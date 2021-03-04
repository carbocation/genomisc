package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
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
)

func main() {

	defer STDOUT.Flush()

	var (
		bgenTemplatePath string
		bgiTemplatePath  string
		inputBucket      string
		layout           string
		sourceFile       string
		customLayout     string
		samplePath       string
		alwaysIncrement  bool
	)
	flag.StringVar(&customLayout, "custom-layout", "", "Optional: a PRS layout with 0-based columns as follows: EffectAlleleCol,Allele1Col,Allele2Col,ChromosomeCol,PositionCol,ScoreCol")
	flag.StringVar(&bgenTemplatePath, "bgen-template", "", "Templated path to bgen with %s in place of its chromosome number")
	flag.StringVar(&bgiTemplatePath, "bgi-template", "", "Optional: Templated path to bgi with %s in place of its chromosome number. If empty, will be replaced with the bgen-template path + '.bgi'")
	flag.StringVar(&inputBucket, "input", "", "Local path to the PRS input file")
	flag.StringVar(&layout, "layout", "LDPRED", fmt.Sprint("Layout of your prs file. Currently, options include: ", prsparser.LayoutNames()))
	flag.StringVar(&sourceFile, "source", "", "Source of your score (e.g., a trait and a version, or whatever you find convenient to track)")
	flag.StringVar(&samplePath, "sample", "", "Path to sample file, which is an Oxford-format file that contains sample IDs for each row in the BGEN")
	flag.BoolVar(&alwaysIncrement, "alwaysincrement", true, "If true, flips effect (and effect allele) at sites with negative effect sizes so that scores will always be > 0.")
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
		if x := len(cols); x != 6 {
			log.Fatalf("--custom-layout was toggled; 6 column numbers were expected, but %d were given\n", x)
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
			ColPosition:     intCols[4],
			ColScore:        intCols[5],
			Parser:          &parseRule,
		}

		log.Println("Using custom parser:")
		fmt.Fprintf(os.Stderr, "%+v\n", udf)

		prsparser.Layouts["CUSTOM"] = udf
	}

	// Connect to Google Storage if requested
	if strings.HasPrefix(inputBucket, "gs://") || strings.HasPrefix(samplePath, "gs://") || strings.HasPrefix(bgenTemplatePath, "gs://") {
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

	wg := sync.WaitGroup{}

	var score []Sample
	scoreChan := make(chan []Sample)
	for chromosome, chromosomalSites := range chromosomalPRS {

		wg.Add(1)
		go func(chromosome string, chromosomalSites []prsparser.PRS) {
			subScore, err := accumulateLoop(chromosome, chromosomalSites, bgenTemplatePath, bgiTemplatePath)
			if err != nil {
				log.Fatalln(err)
			}
			scoreChan <- subScore
		}(chromosome, chromosomalSites)
	}

	// Accumulate
	go func() {
		for range chromosomalPRS {
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

func accumulateLoop(chromosome string, chromosomalSites []prsparser.PRS, bgenTemplatePath, bgiTemplatePath string) ([]Sample, error) {
	// Place to accumulate scores
	var score []Sample
	var err error

	i := 0

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

	if len(score) == 0 {
		// If we haven't initialized the score slice yet, do so now by
		// reading a random variant and creating a slice that has one sample
		// per sample found in that variant.
		vr := b.NewVariantReader()
		randomVariant := vr.Read()
		score = make([]Sample, len(randomVariant.SampleProbabilities))
	}

	// Iterate over each chromosomal position on this current chromosome
	for _, oneSite := range chromosomalSites {

		i++
		if i%100 == 0 {
			log.Println("Processing site", i, oneSite.Chromosome, oneSite.Position, oneSite.Allele1, oneSite.Allele2)
		}

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
