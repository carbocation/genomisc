package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"cloud.google.com/go/storage"
	_ "github.com/carbocation/genomisc/compileinfoprint"
	"github.com/carbocation/genomisc/prsparser"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/pfx"
)

func init() {
	// Add a tab-delimited LDPred processor in addition to the space-delimited
	// one
	ldp := prsparser.Layouts["LDPRED"]
	ldp.Delimiter = '\t'
	prsparser.Layouts["LDPREDTAB"] = ldp
}

var (
	BufferSize = 4096 * 8
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
	client     *storage.Client

	// Yes, an ugly global counter that is atomically updated across goroutines
	nSitesProcessed uint64
)

func main() {

	// defer profile.Start().Stop()

	fmt.Fprintf(os.Stderr, "%q\n", os.Args)

	defer STDOUT.Flush()

	defer func() {
		log.Println("Completed applyprsbasic")
	}()

	var (
		bgenTemplatePath string
		bgiTemplatePath  string
		vcfTemplatePath  string
		vcfiTemplatePath string
		inputBucket      string
		layout           string
		sourceFile       string
		customLayout     string
		samplePath       string
		alwaysIncrement  bool
		stripPRSChrom    bool
		maxConcurrency   int
	)
	flag.StringVar(&customLayout, "custom-layout", "", "Optional: a PRS layout with 0-based columns as follows: EffectAlleleCol,Allele1Col,Allele2Col,ChromosomeCol,PositionCol,ScoreCol,SNPCol. Either PositionCol or SNPCol (but not both) may be set to -1, indicating that it is not present.")
	flag.StringVar(&bgenTemplatePath, "bgen-template", "", "Templated path to bgen with %s in place of its chromosome number")
	flag.StringVar(&bgiTemplatePath, "bgi-template", "", "Optional: Templated path to bgi with %s in place of its chromosome number. If empty, will be replaced with the bgen-template path + '.bgi'")
	flag.StringVar(&vcfTemplatePath, "vcf-template", "", "Templated path to vcf with %s in place of its chromosome number.")
	flag.StringVar(&vcfiTemplatePath, "vcfi-template", "", "Optional: Templated path to vcfi with %s in place of its chromosome number. If empty, will be replaced with the bgen-template path + '.vcf.gz'")
	flag.StringVar(&inputBucket, "input", "", "Local path to the PRS input file")
	flag.StringVar(&layout, "layout", "LDPRED", fmt.Sprint("Layout of your prs file. Currently, options include: ", prsparser.LayoutNames()))
	flag.StringVar(&sourceFile, "source", "", "Source of your score (e.g., a trait and a version, or whatever you find convenient to track)")
	flag.StringVar(&samplePath, "sample", "", "Path to sample file, which is an Oxford-format file that contains sample IDs for each row in the BGEN")
	flag.BoolVar(&alwaysIncrement, "alwaysincrement", true, "If true, flips effect (and effect allele) at sites with negative effect sizes so that scores will always be > 0.")
	flag.BoolVar(&stripPRSChrom, "stripprschr", true, "If true, strips the 'chr' or 'chrom' prefix from the PRS file's chromosome names before processing.")
	flag.IntVar(&maxConcurrency, "maxconcurrency", 0, "(Optional) If greater than 0, will only parallelize to maxConcurrency parallel processes, insted of 2*number of cores (the default).")
	flag.Parse()

	if sourceFile == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide --source")
	}

	if bgenTemplatePath != "" && samplePath == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide --sample")
	}

	if bgenTemplatePath == "" && vcfTemplatePath == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide --bgen-template or --vcf-template")
	}

	if inputBucket == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide --input")
	}

	if bgiTemplatePath == "" && bgenTemplatePath != "" {
		bgiTemplatePath = bgenTemplatePath + ".bgi"
	} else if vcfiTemplatePath == "" && vcfTemplatePath != "" {
		vcfiTemplatePath = vcfTemplatePath + ".vcf.gz.tbi"
	}

	if customLayout != "" {
		layout = "CUSTOM"

		cols := strings.Split(customLayout, ",")
		if x := len(cols); x != 6 && x != 7 {
			log.Fatalf("--custom-layout was toggled; 6 or 7 column numbers were expected, but %d were given\n", x)
		}
		intCols := make([]int, 0, len(cols))
		for i, col := range cols {
			j, err := strconv.ParseInt(col, 10, 32)
			if err != nil {
				log.Fatalf("The identifier for column %d (value %s) is not an integer", i, col)
			}
			intCols = append(intCols, int(j))
		}

		parseRule := func(layout *prsparser.Layout, row []string) (prsparser.PRS, error) {
			p, err := prsparser.DefaultParseRow(layout, row)
			if err != nil {
				return p, pfx.Err(err)
			}

			// ... remove common prefixes from the chomosome column
			if stripPRSChrom {
				p.Chromosome = strings.TrimPrefix(row[layout.ColChromosome], "chrom_")
				p.Chromosome = strings.TrimPrefix(row[layout.ColChromosome], "chr")
			}

			return p, pfx.Err(err)
		}

		udf := prsparser.Layout{
			Delimiter:       '\t', // TODO: make this configurable
			Comment:         '#',  // TODO: make this configurable
			ColEffectAllele: intCols[0],
			ColAllele1:      intCols[1],
			ColAllele2:      intCols[2],
			ColChromosome:   intCols[3],
			ColPosition:     intCols[4], // May be set to -1 if the user is setting ColSNP
			ColScore:        intCols[5],
			ColSNP:          -1,
			Parser:          &parseRule,
		}

		if len(intCols) > 6 && intCols[6] >= 0 {
			udf.ColSNP = intCols[6]
		}

		log.Println("Using custom parser:")
		fmt.Fprintf(os.Stderr, "%+v\n", udf)

		prsparser.Layouts["CUSTOM"] = udf
	}

	// Connect to Google Storage if requested
	if strings.HasPrefix(inputBucket, "gs://") ||
		strings.HasPrefix(samplePath, "gs://") ||
		strings.HasPrefix(vcfTemplatePath, "gs://") ||
		strings.HasPrefix(vcfiTemplatePath, "gs://") ||
		strings.HasPrefix(bgenTemplatePath, "gs://") ||
		strings.HasPrefix(bgiTemplatePath, "gs://") {
		var err error
		client, err = storage.NewClient(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}

	if err := LoadPRS(inputBucket, layout, alwaysIncrement); err != nil {
		log.Fatalln(err)
	}
	log.Println("There are", len(currentVariantScoreLookup), "variants in the PRS database")
	for _, v := range currentVariantScoreLookup {
		log.Println("Example PRS entry from your score file:")
		log.Printf("%+v\n", v)
		break
	}

	chromosomalPRS := ChromosomalPRS(currentVariantScoreLookup)

	// Increase parallelism. The goal is to split the job into 2 * sqrt(jobs)
	// chunks of tasks, or ncpu chunks of tasks, whichever is lower.
	desiredChunks := 2 * int(math.Floor(math.Sqrt(float64(len(currentVariantScoreLookup)))))
	log.Println("Desire about", desiredChunks, "chunks of length", desiredChunks, "(", len(currentVariantScoreLookup), ") in total. NCPU ==", runtime.NumCPU())
	if desiredChunks > 2*runtime.NumCPU() {
		desiredChunks = 2 * runtime.NumCPU()
	}
	if desiredChunks > maxConcurrency && maxConcurrency > 0 {
		desiredChunks = maxConcurrency
	}
	if desiredChunks < 1 {
		desiredChunks = 1
	}
	chunkSize := len(currentVariantScoreLookup) / desiredChunks

	chromosomalPRSChunks := splitter(chromosomalPRS, chunkSize)

	log.Println("Split into", len(chromosomalPRSChunks), "chunks")

	wg := sync.WaitGroup{}
	scoreChan := make(chan []Sample)
	taskCount := 0

	go func() {
		tick := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-tick.C:
				log.Println(atomic.LoadUint64(&nSitesProcessed), "sites have been processed")
			}
		}
	}()

	defer func() {
		log.Println(atomic.LoadUint64(&nSitesProcessed), "sites were processed")
	}()

	// Iterate over the chunks of PRS entries. The only guarantee is that each
	// chunk of PRS SNPs to be scored will be on the same chromosome.
	for whichChunk, chromsomalPRSChunk := range chromosomalPRSChunks {
		taskCount++
		wg.Add(1)
		go func(whichChunk int, chromosome string, chromosomalSites []prsparser.PRS, scoreChan chan []Sample) {

			switch {
			case bgenTemplatePath != "":
				subScore, err := scoreBGEN(chromosome, chromosomalSites, bgenTemplatePath, bgiTemplatePath)
				if err != nil {
					log.Fatalln(err)
				}
				scoreChan <- subScore
			case vcfTemplatePath != "":
				subScore, err := scoreVCF(whichChunk, chromosome, chromosomalSites, vcfTemplatePath, vcfiTemplatePath)
				if err != nil {
					log.Fatalln(err)
				}
				scoreChan <- subScore
			}
		}(whichChunk, chromsomalPRSChunk.Chrom, chromsomalPRSChunk.PRSSites, scoreChan)
	}

	log.Println("Launched", taskCount, "tasks")

	// Accumulate
	score := make([]Sample, 0)
	go func() {
		for i := 0; i < taskCount; i++ {
			scores := <-scoreChan

			if len(score) == 0 {
				score = append(score, scores...)
			} else {
				for k, v := range scores {
					score[k].NIncremented += v.NIncremented
					score[k].SumScore += v.SumScore
				}
			}

			log.Println("Completed", i+1, "of", taskCount, "tasks")
			wg.Done()
		}
	}()

	wg.Wait()

	// Header
	fmt.Printf("sample_id\tsource\tscore\tn_incremented\n")

	// Create a row-number-to-sample-ID mapping
	var sampleFileContentsLookup func(int) string

	if samplePath != "" {
		sf, _, err := bulkprocess.MaybeOpenFromGoogleStorage(samplePath, client)
		if err != nil {
			log.Fatalln(err)
		}
		sfCSV := csv.NewReader(sf)
		sfCSV.Comma = ' '
		sampleFileContents, err := sfCSV.ReadAll()
		if err != nil {
			log.Fatalln(err)
		}
		sampleFileContentsLookup = func(row int) string {
			// .sample files have a header row and a 'header column type' row,
			// so the first 2 entries are not sample IDs and need to be skipped.
			return sampleFileContents[row+2][0]
		}
	} else if vcfTemplatePath != "" {
		sampleFileContentsLookup = func(row int) string {
			return score[row].ID
		}
	}

	for fileRow, v := range score {
		sampleID := sampleFileContentsLookup(fileRow)

		fmt.Fprintf(STDOUT, "%s\t%s\t%f\t%d\n", sampleID, sourceFile, v.SumScore, v.NIncremented)
	}

}
