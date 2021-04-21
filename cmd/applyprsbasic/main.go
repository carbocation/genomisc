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
	"github.com/carbocation/bgen"
	"github.com/carbocation/genomisc/applyprsgcp"
	"github.com/carbocation/genomisc/applyprsgcp/prsworker"
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
	log.Println(bgen.WhichSQLiteDriver())

	defer STDOUT.Flush()

	defer func() {
		log.Println("Completed applyprsbasic")
	}()

	var (
		bgenTemplatePath string
		bgiTemplatePath  string
		inputBucket      string
		layout           string
		sourceFile       string
		customLayout     string
		samplePath       string
		alwaysIncrement  bool
		maxConcurrency   int
	)
	flag.StringVar(&customLayout, "custom-layout", "", "Optional: a PRS layout with 0-based columns as follows: EffectAlleleCol,Allele1Col,Allele2Col,ChromosomeCol,PositionCol,ScoreCol,SNPCol. Either PositionCol or SNPCol (but not both) may be set to -1, indicating that it is not present.")
	flag.StringVar(&bgenTemplatePath, "bgen-template", "", "Templated path to bgen with %s in place of its chromosome number")
	flag.StringVar(&bgiTemplatePath, "bgi-template", "", "Optional: Templated path to bgi with %s in place of its chromosome number. If empty, will be replaced with the bgen-template path + '.bgi'")
	flag.StringVar(&inputBucket, "input", "", "Local path to the PRS input file")
	flag.StringVar(&layout, "layout", "LDPRED", fmt.Sprint("Layout of your prs file. Currently, options include: ", prsparser.LayoutNames()))
	flag.StringVar(&sourceFile, "source", "", "Source of your score (e.g., a trait and a version, or whatever you find convenient to track)")
	flag.StringVar(&samplePath, "sample", "", "Path to sample file, which is an Oxford-format file that contains sample IDs for each row in the BGEN")
	flag.BoolVar(&alwaysIncrement, "alwaysincrement", true, "If true, flips effect (and effect allele) at sites with negative effect sizes so that scores will always be > 0.")
	flag.IntVar(&maxConcurrency, "maxconcurrency", 0, "(Optional) If greater than 0, will only parallelize to maxConcurrency parallel processes, insted of 2*number of cores (the default).")
	flag.Parse()

	if sourceFile == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide --source")
	}

	if samplePath == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide --sample")
	}

	if bgenTemplatePath == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide --bgen-template")
	}

	if inputBucket == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide --input")
	}

	if bgiTemplatePath == "" {
		bgiTemplatePath = bgenTemplatePath + ".bgi"
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
			p.Chromosome = strings.TrimPrefix(row[layout.ColChromosome], "chrom_")
			p.Chromosome = strings.TrimPrefix(row[layout.ColChromosome], "chr")

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

	// Header
	fmt.Printf("sample_id\tsource\tscore\tn_incremented\n")

	// Load the .sample file:
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

	var score []Sample
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

	for _, chromsomalPRSChunk := range chromosomalPRSChunks {
		taskCount++
		wg.Add(1)
		go func(chromosome string, chromosomalSites []prsparser.PRS) {
			subScore, err := accumulateLoop(chromosome, chromosomalSites, bgenTemplatePath, bgiTemplatePath)
			if err != nil {
				log.Fatalln(err)
			}
			scoreChan <- subScore
		}(chromsomalPRSChunk.Chrom, chromsomalPRSChunk.PRSSites)
	}

	log.Println("Launched", taskCount, "tasks")

	// Accumulate
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

			wg.Done()
		}
	}()

	wg.Wait()

	for fileRow, v := range score {
		// +2 because the sample file contains a header of 2 rows: (1) the true
		// header, and (2) a second header indicating the value type of the
		// column
		sampleID := sampleFileContents[fileRow+2][0]

		fmt.Fprintf(STDOUT, "%s\t%s\t%f\t%d\n", sampleID, sourceFile, v.SumScore, v.NIncremented)
	}

}

type PRSSitesOnChrom struct {
	Chrom    string
	PRSSites []prsparser.PRS
}

// splitter parallelizes the computation. Each separate goroutine will process a
// set of sites. Each goroutine will have its own BGEN and BGI filehandles.
func splitter(chromosomalPRS map[string][]prsparser.PRS, chunkSize int) []PRSSitesOnChrom {
	out := make([]PRSSitesOnChrom, 0)

	var subChunk PRSSitesOnChrom

	// For each chromosome
	for chrom, prsSites := range chromosomalPRS {
		if subChunk.Chrom == "" {
			// We are initializing from scratch
			subChunk.Chrom = chrom
			subChunk.PRSSites = make([]prsparser.PRS, 0, chunkSize)
		}

		if subChunk.Chrom != chrom {
			// We have moved to a new chromosome
			out = append(out, subChunk)
			subChunk = PRSSitesOnChrom{
				Chrom:    chrom,
				PRSSites: make([]prsparser.PRS, 0, chunkSize),
			}
		}

		// Within this chromosome, add sites into different chunks
		for k, prsSite := range prsSites {
			subChunk.PRSSites = append(subChunk.PRSSites, prsSite)

			// But if we have iterated to our chunk size, then create a new
			// chunk
			if k > 0 && k%chunkSize == 0 {
				out = append(out, subChunk)
				subChunk = PRSSitesOnChrom{
					Chrom:    chrom,
					PRSSites: make([]prsparser.PRS, 0, chunkSize),
				}
			}
		}
	}

	// Clean up at the end
	if subChunk.Chrom != "" {
		out = append(out, subChunk)
	}

	return out
}

func accumulateLoop(chromosome string, chromosomalSites []prsparser.PRS, bgenTemplatePath, bgiTemplatePath string) ([]Sample, error) {
	// Place to accumulate scores
	var score []Sample
	var err error

	// Load the BGEN Index for this chromosome
	bgenPath := fmt.Sprintf(bgenTemplatePath, chromosome)
	if !strings.Contains(bgenTemplatePath, "%s") {
		// Permit explicit paths (e.g., when all data is in one BGEN)
		bgenPath = bgenTemplatePath
	}

	bgiPath := fmt.Sprintf(bgiTemplatePath, chromosome) //bgenPath + ".bgi"
	if !strings.Contains(bgiTemplatePath, "%s") {
		// Permit explicit paths (e.g., when all data is in one BGEN)
		bgiPath = bgiTemplatePath
	}

	// Open the BGI
	var bgi *bgen.BGIIndex
	var b *bgen.BGEN

	// Repeatedly reading SQLite files over-the-wire is slow. So localize
	// them.
	if strings.HasPrefix(bgiPath, "gs://") {
		bgiFilePath, err := applyprsgcp.ImportBGIFromGoogleStorage(bgiPath, client)
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("Copied BGI file from %s to %s\n", bgiPath, bgiFilePath)

		bgiPath = bgiFilePath
	}

	bgi, b, err = prsworker.OpenBGIAndBGEN(bgenPath, bgiPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer bgi.Close()
	defer b.Close()

	// if len(score) == 0 {
	// 	// If we haven't initialized the score slice yet, do so now by
	// 	// reading a random variant and creating a slice that has one sample
	// 	// per sample found in that variant.
	// 	vr := b.NewVariantReader()
	// 	randomVariant := vr.Read()
	// 	score = make([]Sample, len(randomVariant.SampleProbabilities))
	// }

	// Iterate over each chromosomal position on this current chromosome
	for _, oneSite := range chromosomalSites {

		atomic.AddUint64(&nSitesProcessed, 1)

		var sites []bgen.VariantIndex

		// Find the site in the BGEN index file. Here we also handle a
		// specific filesystem error (input/output error) and retry in that
		// case. Namely, exiting early is superior to producing incorrect
		// output in the face of i/o errors. But, since we expect other
		// "error" messages, we permit those to just be printed and ignored.
		for loadAttempts, maxLoadAttempts := 1, 10; loadAttempts <= maxLoadAttempts; loadAttempts++ {
			sites, err = FindPRSSiteInBGI(bgi, oneSite)
			if err != nil && loadAttempts == maxLoadAttempts {
				// Ongoing failure at maxLoadAttempts is a terminal error
				log.Fatalln(err)
			} else if err != nil && strings.Contains(err.Error(), "input/output error") {
				// If we had an error, often due to an unreliable underlying
				// filesystem, wait for a substantial amount of time before
				// retrying.
				log.Println("FindPRSSiteInBGI: Sleeping 5s to recover from", err.Error(), "attempt", loadAttempts)
				time.Sleep(5 * time.Second)

				// We also seemingly need to reopen the handles.
				b.Close()
				bgi.Close()
				bgi, b, err = prsworker.OpenBGIAndBGEN(bgenPath, bgiPath)
				if err != nil {
					log.Fatalln(err)
				}

				continue
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "Skipping %v\n:", oneSite)
				log.Println(err)
			}

			// No errors? Don't retry.
			break
		}

		// Multiple variants may be present at each chromosomal position
		// (e.g., multiallelic sites, long deletions/insertions, etc) so
		// test each of them for being an exact match.
	SitesLoop:
		for _, site := range sites {

			// Process a variant from the BGEN file. Again, we also handle a
			// specific filesystem error (input/output error) and retry in
			// that case.
			for loadAttempts, maxLoadAttempts := 1, 10; loadAttempts <= maxLoadAttempts; loadAttempts++ {

				err = ProcessOneVariant(b, site, &oneSite, &score)
				if err != nil && loadAttempts == maxLoadAttempts {
					// Ongoing failure at maxLoadAttempts is a terminal error
					log.Fatalln(err)
				} else if err != nil && strings.Contains(err.Error(), "input/output error") {
					// If we had an error, often due to an unreliable underlying
					// filesystem, wait for a substantial amount of time before retrying.
					log.Println("ProcessOneVariant: Sleeping 5s to recover from", err.Error(), "attempt", loadAttempts)
					time.Sleep(5 * time.Second)

					// We also seemingly need to reopen the handles.
					b.Close()
					bgi.Close()
					bgi, b, err = prsworker.OpenBGIAndBGEN(bgenPath, bgiPath)
					if err != nil {
						log.Fatalln(err)
					}

					continue
				} else if err != nil {
					// Non i/o errors usually indicate that this is the wrong allele of a multiallelic. Print and move on.
					log.Println(err)
					continue SitesLoop
				}

				// No errors? Don't retry.
				break
			}
		}
	}

	// Clean up our file handles
	bgi.Close()
	b.Close()

	return score, nil
}
