package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/carbocation/bgen"
	"github.com/carbocation/genomisc/prsparser"
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

	fmt.Fprintf(os.Stderr, "%q\n", os.Args)

	var (
		bgenTemplatePath string
		inputBucket      string
		layout           string
		chromosome       string
		firstLine        int
		lastLine         int
		sourceFile       string
		customLayout     string
		alwaysIncrement  bool
	)
	flag.StringVar(&customLayout, "custom-layout", "", "Optional: a PRS layout with 0-based columns as follows: EffectAlleleCol,Allele1Col,Allele2Col,ChromosomeCol,PositionCol,ScoreCol")
	flag.StringVar(&bgenTemplatePath, "bgen-template", "", "Templated path to bgen with %s in place of its chromosome number")
	flag.StringVar(&inputBucket, "input", "", "Local path to the PRS input file")
	flag.StringVar(&layout, "layout", "LDPRED", fmt.Sprint("Layout of your prs file. Currently, options include: ", prsparser.LayoutNames()))
	flag.StringVar(&chromosome, "chromosome", "", "Chromosome this job is looking at")
	flag.StringVar(&sourceFile, "source", "", "Source of your score (e.g., a trait and a version, or whatever you find convenient to track)")
	flag.IntVar(&firstLine, "first_line", 0, "First line in the file to start counting toward the score")
	flag.IntVar(&lastLine, "last_line", 0, "Last line in the file to count toward the score")
	flag.BoolVar(&alwaysIncrement, "alwaysincrement", true, "If true, flips effect (and effect allele) at sites with negative effect sizes so that scores will always be > 0.")
	flag.Parse()

	if chromosome == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide --chromosome")
	}

	if sourceFile == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide --source")
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
				return p, err
			}

			// ... remove common prefixes from the chomosome column
			p.Chromosome = strings.TrimPrefix(row[layout.ColChromosome], "chrom_")
			p.Chromosome = strings.TrimPrefix(row[layout.ColChromosome], "chr")

			return p, err
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

	if err := LoadPRSInRange(inputBucket, layout, chromosome, firstLine, lastLine, alwaysIncrement); err != nil {
		log.Fatalln(err)
	}
	log.Println("There are", len(currentVariantScoreLookup), "variants in the PRS database within this job's range")
	for _, v := range currentVariantScoreLookup {
		log.Println("Example PRS entry from your score file:")
		log.Printf("%+v\n", v)
		break
	}

	bgenPath := fmt.Sprintf(bgenTemplatePath, chromosome)

	var sites []bgen.VariantIndex
	var err error

	// Because of i/o errors on GCP, may need to loop this
	for loadAttempts, maxLoadAttempts := 1, 10; loadAttempts <= maxLoadAttempts; loadAttempts++ {

		// Extract sites from the BGEN file that overlap with our PRS sites
		sites, err = BGENPreprocessor(bgenPath, chromosome)
		if err != nil && loadAttempts == maxLoadAttempts {

			// Ongoing failure at maxLoadAttempts is a terminal error
			log.Fatalln(err)

		} else if err != nil && strings.Contains(err.Error(), "input/output error") {

			// If we had an error, often due to an unreliable underlying
			// filesystem, wait for a substantial amount of time before
			// retrying.
			log.Println("BGENPreprocessor: Sleeping 5s to recover from", err.Error(), "attempt", loadAttempts)
			time.Sleep(5 * time.Second)

			continue

		} else if err != nil {

			// Not simply an i/o error:
			log.Fatalln(err)

		}

		// If loading the data was error-free, no additional attempts are
		// required.
		break
	}

	score, err := AccumulateScore(bgenPath, sites)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("sample_file_row\tsource\tscore\tn_incremented\n")
	for fileRow, v := range score {
		fmt.Fprintf(STDOUT, "%d\t%s\t%f\t%d\n", fileRow, sourceFile, v.SumScore, v.NIncremented)
	}

}
