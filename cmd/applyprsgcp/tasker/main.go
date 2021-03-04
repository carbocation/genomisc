package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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
		customLayout   string
		hasHeader      bool
		variantsPerJob int
	)
	flag.StringVar(&customLayout, "custom-layout", "", "Optional: a PRS layout with 0-based columns as follows: EffectAlleleCol,Allele1Col,Allele2Col,ChromosomeCol,PositionCol,ScoreCol. E.g., 1,1,2,0,3,5")
	flag.StringVar(&inputBucket, "input", "", "Google Storage file with to source PRS containing SNP weights. This must have the same file content as --prs.")
	flag.StringVar(&outputBucket, "output", "", "Google Storage file which is a 'root name' that will be slightly modified for each output chunk")
	flag.StringVar(&prsPath, "prs", "", "Local file containing your polygenic risk score. Must be sorted by chromosome. Will be parsed to create an appropriate number of tasks per job.")
	flag.StringVar(&layout, "layout", "LDPRED", fmt.Sprint("Layout of your prs file. Currently, options include: ", prsparser.LayoutNames()))
	flag.BoolVar(&hasHeader, "header", true, "(Optional) Does the input file have a header that needs to be skipped?")
	flag.StringVar(&sourceOverride, "column_name", "", "(Optional) If set, overrides the default 'source' column name (derived from the PRS filename) with this value.")
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

	if err := tasker.CreateTasks(inputBucket, outputBucket, prsPath, layout, hasHeader, variantsPerJob, sourceOverride, customLayout); err != nil {
		log.Fatalln(err)
	}
}
