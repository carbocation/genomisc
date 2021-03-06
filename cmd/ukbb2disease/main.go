package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"cloud.google.com/go/bigquery"
)

type WrappedBigQuery struct {
	Context        context.Context
	Client         *bigquery.Client
	Project        string
	Database       string
	MaterializedDB string
	UseGP          bool
}

// Special value that is to be set using ldflags
// E.g.: go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"
// Consider aliasing in .profile: alias gobuild='go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"'
var builddate string

var (
	BufferSize = 4096 * 8
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
)

func init() {
	flag.Usage = func() {
		flag.PrintDefaults()
		describeDateFields(false)
	}
}

func main() {
	defer STDOUT.Flush()

	fmt.Fprintf(os.Stderr, "This ukbb2disease binary was built at: %s\n", builddate)

	var BQ = &WrappedBigQuery{
		Context: context.Background(),
	}
	var tabfile string
	var displayQuery bool
	var override bool
	var allowUndated bool
	var diseaseName string
	var verbose bool

	flag.StringVar(&BQ.Project, "project", "", "Google Cloud project you want to use for billing purposes only")
	flag.StringVar(&BQ.Database, "database", "", "BigQuery source database name (note: must be formatted as project.database, e.g., ukbb-analyses.ukbb7089_201904)")
	flag.StringVar(&tabfile, "tabfile", "", "Tabfile-formatted phenotype definition")
	flag.StringVar(&BQ.MaterializedDB, "materialized", "", "(Deprecated, not used - will be removed in future versions)")
	flag.BoolVar(&displayQuery, "display-query", false, "Display the constructed query and exit?")
	flag.BoolVar(&override, "override", false, "Force run, even if this tool thinks your tabfile is inadequate?")
	flag.BoolVar(&allowUndated, "allow-undated", false, "Force run, even if your tabfile has fields whose date is unknown (which will cause matching participants to be set to prevalent)?")
	flag.BoolVar(&verbose, "verbose", false, "Print all ~ 2,000 fields whose dates are known?")
	flag.StringVar(&diseaseName, "disease", "", "If not specified, the tabfile will be parsed and become the disease name.")
	flag.BoolVar(&BQ.UseGP, "usegp", false, "")
	flag.Parse()

	flag.Usage = func() {
		flag.PrintDefaults()
		describeDateFields(verbose)
	}

	// Deprecating BQ.MaterializedDB - adds unnecessary complexity; reasonable
	// to assume that all data is in the same database
	BQ.MaterializedDB = BQ.Database

	if BQ.Project == "" {
		fmt.Fprintln(os.Stderr, "Please provide --project")
		flag.Usage()
		os.Exit(1)
	}

	if BQ.Database == "" {
		fmt.Fprintln(os.Stderr, "Please provide --database")
		flag.Usage()
		os.Exit(1)
	}

	if tabfile == "" {
		fmt.Fprintln(os.Stderr, "Please provide --tabfile")
		flag.Usage()
		os.Exit(1)
	}

	if BQ.MaterializedDB == "" {
		fmt.Fprintln(os.Stderr, "Please provide --materialized")
		flag.Usage()
		os.Exit(1)
	}

	tabs, err := ParseTabFile(tabfile)
	if err != nil {
		log.Fatalln(err)
	}

	if diseaseName == "" {
		diseaseName = path.Base(tabfile)
		if parts := strings.Split(diseaseName, "."); len(parts) > 1 {
			diseaseName = strings.Join(parts[0:len(parts)-1], ".")
		}
	}

	log.Println("Processing disease", diseaseName)

	missingFields, err := tabs.CheckSensibility()
	if err != nil && !override {
		log.Println(err)
		log.Fatalf("%s: Add the missing fields to your tabfile, or re-run with the -override flag to process anyway.\n", diseaseName)
	} else if err != nil && override {
		log.Println(diseaseName, err)
		log.Printf("%s: Overriding error check for missing fields and continuing.\n", diseaseName)
	}

	// Check for the use of fields where we don't know the date (which forces
	// them to set disease status to prevalent)
	_, err = tabs.CheckUndatedFields()
	if err != nil && !allowUndated {
		log.Println(err)
		log.Fatalf("%s: Remove the undated fields from your tabfile, update this software with information about the dates of those fields, or re-run with the -allow-undated flag to process anyway.\n", diseaseName)
	} else if err != nil && allowUndated {
		log.Println(diseaseName, err)
		log.Printf("%s: Overriding failed undated field check and continuing.\n", diseaseName)
	}

	fmt.Fprintln(os.Stderr, "Including:")
	for _, v := range tabs.AllIncluded() {
		fmt.Fprintf(os.Stderr, "\tFieldID %v values:\n", v.FieldID)
		fmt.Fprintf(os.Stderr, "\t\t")
		fmt.Fprintf(os.Stderr, "%s", strings.Join(v.Values, ", "))
		fmt.Fprintf(os.Stderr, "\n")
	}
	fmt.Fprintf(os.Stderr, "\n")

	fmt.Fprintln(os.Stderr, "Excluding:")
	for _, v := range tabs.AllExcluded() {
		fmt.Fprintf(os.Stderr, "\tFieldID %v values:\n", v.FieldID)
		fmt.Fprintf(os.Stderr, "\t\t")
		fmt.Fprintf(os.Stderr, "%s", strings.Join(v.Values, ", "))
		fmt.Fprintf(os.Stderr, "\n")
	}
	if len(tabs.AllExcluded()) < 1 {
		fmt.Fprintf(os.Stderr, "\t(No exclusion criteria)\n")
	}

	BQ.Client, err = bigquery.NewClient(BQ.Context, BQ.Project)
	if err != nil {
		log.Fatalln("Connecting to BigQuery:", err)
	}
	defer BQ.Client.Close()

	query, err := BuildQuery(BQ, tabs, displayQuery)
	if err != nil {
		log.Fatalln(diseaseName, err)
	}

	if displayQuery {
		return
	}

	if err := ExecuteQuery(BQ, query, diseaseName, missingFields); err != nil {
		log.Fatalln(diseaseName, err)
	}

	fmt.Fprintf(os.Stderr, "Finished producing output for %s\n", diseaseName)
}
