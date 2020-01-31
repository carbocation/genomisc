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
	Context  context.Context
	Client   *bigquery.Client
	Project  string
	Database string
}

// Special value that is to be set using ldflags
// E.g.: go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"
// Consider aliasing in .profile: alias gobuild='go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"'
var builddate string

var (
	BufferSize = 4096 * 8
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
)

var materializedDB string

func init() {
	flag.Usage = func() {
		flag.PrintDefaults()
		describeDateFields()
	}
}

func main() {
	defer STDOUT.Flush()

	fmt.Fprintf(os.Stderr, "This ukbb2disease binary was built at: %s\n", builddate)

	var tabfile string
	var diseaseName string
	var dictPath string
	var codingPath string
	var override bool
	var allowUndated bool

	flag.StringVar(&codingPath, "coding", "https://biobank.ctsu.ox.ac.uk/~bbdatan/Codings_Showcase.csv", "URL or path to CSV file with the UKBB data encodings")
	flag.StringVar(&dictPath, "dict", "https://biobank.ndph.ox.ac.uk/~bbdatan/Data_Dictionary_Showcase.csv", "URL or path to CSV file with the UKBB data dictionary")
	flag.StringVar(&tabfile, "tabfile", "", "Tabfile-formatted phenotype definition")
	flag.BoolVar(&override, "override", false, "Force run, even if this tool thinks your tabfile is inadequate?")
	flag.BoolVar(&allowUndated, "allow-undated", false, "Force run, even if your tabfile has fields whose date is unknown (which will cause matching participants to be set to prevalent)?")
	flag.Parse()

	if tabfile == "" {
		fmt.Fprintln(os.Stderr, "Please provide --tabfile")
		flag.Usage()
		os.Exit(1)
	}

	dict, err := ImportDictionary(dictPath)
	if err != nil {
		log.Fatalln(err)
	}

	coding, err := ImportCoding(codingPath)
	if err != nil {
		log.Fatalln(err)
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
		log.Fatalf("%s: Add the missing fields to your tabfile [%s], or re-run with the -override flag to process anyway.\n", diseaseName, strings.Join(missingFields, ","))
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
		dictEntry := dict[v.FieldID]

		fmt.Fprintf(os.Stderr, "\t%s (FieldID %v) values:\n", dictEntry.Field, v.FieldID)

		codingEntry, exists := coding[dictEntry.Coding.String()]
		if exists {
			for _, entryValue := range v.FormattedValues() {
				fmt.Fprintf(os.Stderr, "\t\t%s\n", codingEntry[entryValue])
			}
		} else {
			fmt.Fprintf(os.Stderr, "\t\t")
			fmt.Fprintf(os.Stderr, "%s", strings.Join(v.Values, ", "))
			fmt.Fprintf(os.Stderr, "\n")
		}

	}
	fmt.Fprintf(os.Stderr, "\n")

	fmt.Fprintln(os.Stderr, "Excluding:")
	for _, v := range tabs.AllExcluded() {
		dictEntry := dict[v.FieldID]

		fmt.Fprintf(os.Stderr, "\t%s (FieldID %v) values:\n", dictEntry.Field, v.FieldID)

		codingEntry, exists := coding[dictEntry.Coding.String()]
		if exists {
			for _, entryValue := range v.FormattedValues() {
				fmt.Fprintf(os.Stderr, "\t\t%s", codingEntry[entryValue])
			}
		} else {
			fmt.Fprintf(os.Stderr, "\t\t")
			fmt.Fprintf(os.Stderr, "%s", strings.Join(v.Values, ", "))
			fmt.Fprintf(os.Stderr, "\n")
		}
	}
	if len(tabs.AllExcluded()) < 1 {
		fmt.Fprintf(os.Stderr, "\t(No exclusion criteria)\n")
	}

}
