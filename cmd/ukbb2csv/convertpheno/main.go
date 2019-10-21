package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/carbocation/genomisc/ukbb2csv/convertpheno"
)

func main() {
	var (
		checkMethod    string
		phenoPathList  string
		dictionaryFile string
		acknowledge    bool
		BQ             = &convertpheno.WrappedBigQuery{}
	)
	flag.StringVar(&checkMethod, "method", "bigquery", "Method of checking for pre-existing fields. Options: bigquery, none")
	flag.StringVar(&BQ.Project, "project", "", "Name of the Google Cloud project that hosts your BigQuery dataset (only if method == 'bigquery')")
	flag.StringVar(&BQ.Database, "bigquery", "", "BigQuery phenotype dataset that will/does contain the 'phenotype' table (only if method == 'bigquery')")
	flag.StringVar(&dictionaryFile, "dictionary", "", "File containing the parsed output from `convertdict`. Required if method == 'none'")
	flag.StringVar(&phenoPathList, "pheno", "", "File containing the paths of each phenotype file for the UKBB that you want to process in this run. Each file should be on its own line. The files with the newest data should be listed first: every FieldID that is seen in an earlier file will be ignored in later files.")
	flag.BoolVar(&acknowledge, "ack", false, "Acknowledge the limitations of the tool")
	flag.Parse()

	if phenoPathList == "" {
		log.Println("Please pass --pheno a file containing paths to the phenotype data.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if checkMethod != "bigquery" && checkMethod != "none" {
		log.Println("Valid --method options include bigquery and none")
		flag.PrintDefaults()
		os.Exit(1)
	}

	phenoPaths, err := convertpheno.ExtractPhenoFileNames(phenoPathList)
	if err != nil {
		log.Fatalln(err)
	}

	for _, phenoPath := range phenoPaths {
		if strings.HasSuffix(phenoPath, ".gz") {
			log.Printf("\n**\n**\nWARNING This tool does not currently operate on gzipped files. Based on your filename, this will likely crash. Please gunzip %s\n**\n**\n", phenoPath)
		}

		if _, err := os.Stat(phenoPath); os.IsNotExist(err) {
			log.Fatalf("Fatal error: %v does not exist\n", phenoPath)
		} else if err != nil {
			log.Fatalf("Fatal error: %v (possibly disk or permissions issues?): %v\n", phenoPath, err)
		}
	}

	if !acknowledge && checkMethod == "bigquery" {
		fmt.Fprintln(os.Stderr, "!! -- !!")
		fmt.Fprintln(os.Stderr, "NOTE")
		fmt.Fprintln(os.Stderr, "!! -- !!")
		fmt.Fprintf(os.Stderr, `This tool checks the %s:%s.phenotype table to avoid duplicating data. 
However, it only looks at data that is already present in the BigQuery table. 
Any data that has been emitted to disk but not yet loaded into BigQuery cannot be seen by this tool.
If you pass multiple files to this tool, it will behave as expected.
If you run this tool multiple times (once per file) without first loading the output from the last round into BigQuery, you may get duplicated data.
!! -- !!
Please re-run this tool with the --ack flag to demonstrate that you understand this limitation.%s`, BQ.Project, BQ.Database, "\n")
		os.Exit(1)
	}

	log.Println("Processing", len(phenoPaths), "files:", phenoPaths)

	switch checkMethod {
	case "bigquery":
		if err := convertpheno.RunAllBackendBQ(phenoPaths, BQ); err != nil {
			log.Fatalln(err)
		}

	case "none":
		if err := convertpheno.RunAllBackendCSV(phenoPaths, dictionaryFile); err != nil {
			log.Fatalln(err)
		}
	}
}
