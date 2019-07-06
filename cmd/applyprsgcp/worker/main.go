package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

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
	)
	flag.StringVar(&bgenTemplatePath, "bgen-template", "", "Templated path to bgen with %s in place of its chromosome number")
	flag.StringVar(&inputBucket, "input", "", "Local path to the PRS input file")
	flag.StringVar(&layout, "layout", "LDPRED", fmt.Sprint("Layout of your prs file. Currently, options include: ", prsparser.LayoutNames()))
	flag.StringVar(&chromosome, "chromosome", "", "Chromosome this job is looking at")
	flag.IntVar(&firstLine, "first_line", 0, "First line in the file to start counting toward the score")
	flag.IntVar(&lastLine, "last_line", 0, "Last line in the file to count toward the score")
	flag.Parse()

	if chromosome == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide --chromosome")
	}

	if bgenTemplatePath == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide --bgen-template")
	}

	if inputBucket == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide --input")
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

	fmt.Printf("sample_file_row_id\tscore\tn_incremented\n")
	for file_row, v := range score {
		fmt.Fprintf(STDOUT, "%d\t%f\t%d\n", file_row, v.SumScore, v.NIncremented)
	}

}
