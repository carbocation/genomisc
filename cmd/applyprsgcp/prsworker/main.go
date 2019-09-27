package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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

	var (
		bgenTemplatePath string
		inputBucket      string
		layout           string
		chromosome       string
		firstLine        int
		lastLine         int
		sourceFile       string
		customLayout     string
	)
	flag.StringVar(&customLayout, "custom-layout", "", "Optional: a PRS layout with 0-based columns as follows: EffectAlleleCol,Allele1Col,Allele2Col,ChromosomeCol,PositionCol,ScoreCol")
	flag.StringVar(&bgenTemplatePath, "bgen-template", "", "Templated path to bgen with %s in place of its chromosome number")
	flag.StringVar(&inputBucket, "input", "", "Local path to the PRS input file")
	flag.StringVar(&layout, "layout", "LDPRED", fmt.Sprint("Layout of your prs file. Currently, options include: ", prsparser.LayoutNames()))
	flag.StringVar(&chromosome, "chromosome", "", "Chromosome this job is looking at")
	flag.StringVar(&sourceFile, "source", "", "Source of your score (e.g., a trait and a version, or whatever you find convenient to track)")
	flag.IntVar(&firstLine, "first_line", 0, "First line in the file to start counting toward the score")
	flag.IntVar(&lastLine, "last_line", 0, "Last line in the file to count toward the score")
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

		udf := prsparser.Layout{
			Delimiter:       '\t', // TODO: make this configurable
			Comment:         '#',  // TODO: make this configurable
			ColEffectAllele: intCols[0],
			ColAllele1:      intCols[1],
			ColAllele2:      intCols[2],
			ColChromosome:   intCols[3],
			ColPosition:     intCols[4],
			ColScore:        intCols[5],
		}

		log.Println("Using custom parser:")
		fmt.Fprintf(os.Stderr, "%+v\n", udf)

		prsparser.Layouts["CUSTOM"] = udf
	}

	if err := LoadPRSInRange(inputBucket, layout, chromosome, firstLine, lastLine); err != nil {
		log.Fatalln(err)
	}
	log.Println("There are", len(currentVariantScoreLookup), "variants in the PRS database within this job's range")

	// Load the BGEN Index
	bgenPath := fmt.Sprintf(bgenTemplatePath, chromosome)

	// Extract sites from the BGEN file that overlap with our PRS sites
	sites, err := BGENPreprocessor(bgenPath, chromosome)
	if err != nil {
		log.Fatalln(err)
	}

	score, err := AccumulateScore(bgenPath, sites)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("sample_file_row\tsource\tscore\tn_incremented\n")
	for file_row, v := range score {
		fmt.Fprintf(STDOUT, "%d\t%s\t%f\t%d\n", file_row, sourceFile, v.SumScore, v.NIncremented)
	}

}
