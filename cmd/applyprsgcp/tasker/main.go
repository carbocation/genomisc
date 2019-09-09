package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/carbocation/genomisc/applyprsgcp/tasker"
	"github.com/carbocation/genomisc/prsparser"
)

func init() {
	// Add a tab-based LDPRED preprocessor

	ldp := prsparser.Layouts["LDPRED"]
	ldp.Delimiter = '\t'

	tasker.AddPRSParserLayout(ldp, "LDPREDTAB")
}

func main() {
	var (
		inputBucket    string
		outputBucket   string
		prsPath        string
		layout         string
		sourceOverride string
		hasHeader      bool
		variantsPerJob int
	)
	flag.StringVar(&inputBucket, "input", "", "Google Storage bucket path where the PRS file will be found")
	flag.StringVar(&outputBucket, "output", "", "Google Storage bucket path where output files should go")
	flag.StringVar(&prsPath, "prs", "", "Path to text file containing your polygenic risk score. Must be sorted by chromosome.")
	flag.StringVar(&layout, "layout", "LDPRED", fmt.Sprint("Layout of your prs file. Currently, options include: ", prsparser.LayoutNames()))
	flag.BoolVar(&hasHeader, "header", true, "Does the input file have a header that needs to be skipped?")
	flag.StringVar(&sourceOverride, "column_name", "", "Optional. If set, overrides the default column name (derived from the PRS filename) with this value.")
	flag.IntVar(&variantsPerJob, "variants_per_job", 0, "Maximum number of variants to be processed by each individual worker")
	flag.Parse()

	// PRS
	if prsPath == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide the path to a file with your polygenic risk score effects.")
	}

	if outputBucket == "" || inputBucket == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide the path to a google bucket folder where the input can be found and the output can be placed")
	}

	if variantsPerJob <= 0 {
		flag.PrintDefaults()
		log.Fatalln("Variants per job cannot be set to 0 or fewer")
	}

	if err := tasker.CreateTasks(inputBucket, outputBucket, prsPath, layout, hasHeader, variantsPerJob, sourceOverride); err != nil {
		log.Fatalln(err)
	}
}
