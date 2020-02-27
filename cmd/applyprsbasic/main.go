package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/carbocation/bgen"
	"github.com/carbocation/genomisc/prsparser"
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
	BufferSize = 4096
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
)

func main() {
	defer STDOUT.Flush()

	var (
		bgenTemplatePath string
		inputBucket      string
		layout           string
		sourceFile       string
		customLayout     string
		samplePath       string
	)
	flag.StringVar(&customLayout, "custom-layout", "", "Optional: a PRS layout with 0-based columns as follows: EffectAlleleCol,Allele1Col,Allele2Col,ChromosomeCol,PositionCol,ScoreCol")
	flag.StringVar(&bgenTemplatePath, "bgen-template", "", "Templated path to bgen with %s in place of its chromosome number")
	flag.StringVar(&inputBucket, "input", "", "Local path to the PRS input file")
	flag.StringVar(&layout, "layout", "LDPRED", fmt.Sprint("Layout of your prs file. Currently, options include: ", prsparser.LayoutNames()))
	flag.StringVar(&sourceFile, "source", "", "Source of your score (e.g., a trait and a version, or whatever you find convenient to track)")
	flag.StringVar(&samplePath, "sample", "", "Path to sample file, which is an Oxford-format file that contains sample IDs for each row in the BGEN")
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

	if err := LoadPRS(inputBucket, layout); err != nil {
		log.Fatalln(err)
	}
	log.Println("There are", len(currentVariantScoreLookup), "variants in the PRS database")
	for _, v := range currentVariantScoreLookup {
		log.Println("Example PRS entry from your score file:")
		log.Printf("%+v\n", v)
		break
	}

	// Place to accumulate scores
	score := make([]Sample, 0)

	// Header
	fmt.Printf("sample_id\tsource\tscore\tn_incremented\n")

	// Load the .sample file:
	sf, err := os.Open(samplePath)
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

	// Iterate over each chromosome present in our PRS
	for chromosome, chomosomalSites := range chromosomalPRS {

		// Load the BGEN Index for this chromosome
		bgenPath := fmt.Sprintf(bgenTemplatePath, chromosome)
		if !strings.Contains(bgenTemplatePath, "%s") {
			// Permit explicit paths (e.g., when all data is in one BGEN)
			bgenPath = bgenTemplatePath
		}
		bgi, err := bgen.OpenBGI(bgenPath + ".bgi?mode=ro")
		if err != nil {
			log.Fatalln(err)
		}
		bgi.Metadata.FirstThousandBytes = nil
		log.Printf("BGI Metadata: %+v\n", bgi.Metadata)

		// Load the BGEN for the chromosome
		b, err := bgen.Open(bgenPath)
		if err != nil {
			panic(err.Error())
		}

		// Iterate over each chromosomal position on this current chromosome
		for _, oneSite := range chomosomalSites {

			// Find the site in the BGEN index file
			sites, err := FindPRSSiteInBGI(bgi, oneSite)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Skipping %v\n:", oneSite)
				log.Println(err)
			}

			// Multiple variants may be present at each chromosomal position
			// (e.g., multiallelic sites, long deletions/insertions, etc) so
			// test each of them for being an exact match.
			for _, site := range sites {
				scoresThisSite, err := ProcessOneVariant(b, site, &oneSite)
				if err != nil {
					log.Println(err)
					continue
				}

				if len(score) == 0 {
					score = scoresThisSite
				} else {
					// TODO: Consider only summing periodically - otherwise you are
					// doing N*M operations
					for k, v := range scoresThisSite {
						prior := score[k]
						prior.NIncremented += v.NIncremented
						prior.SumScore += v.SumScore

						score[k] = prior
					}
				}
			}

		}

		// Clean up our file handles
		bgi.Close()
		b.Close()
	}

	for fileRow, v := range score {
		// +2 because the sample file contains a header of 2 rows: (1) the true
		// header, and (2) a second header indicating the value type of the
		// column
		sampleID := sampleFileContents[fileRow+2][0]

		fmt.Fprintf(STDOUT, "%s\t%s\t%f\t%d\n", sampleID, sourceFile, v.SumScore, v.NIncremented)
	}

}
