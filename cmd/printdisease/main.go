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
		describeDateFields(false)
	}
}

func main() {
	defer STDOUT.Flush()

	fmt.Fprintf(os.Stderr, "This printdisease binary was built at: %s\n", builddate)

	var tabfile string
	var diseaseName string
	var dictPath string
	var codingPath string
	var override bool
	var allowUndated bool
	var verbose bool
	var simplified bool

	flag.StringVar(&codingPath, "coding", "https://biobank.ctsu.ox.ac.uk/~bbdatan/Codings.csv", "URL or path to comma-delimited file with the UKBB data encodings, specified at http://biobank.ctsu.ox.ac.uk/crystal/exinfo.cgi?src=AccessingData")
	flag.StringVar(&dictPath, "dict", "https://biobank.ndph.ox.ac.uk/~bbdatan/Data_Dictionary_Showcase.csv", "URL or path to comma-delimited file with the UKBB data dictionary, specified at http://biobank.ctsu.ox.ac.uk/crystal/exinfo.cgi?src=AccessingData")
	flag.StringVar(&tabfile, "tabfile", "", "Tabfile-formatted phenotype definition")
	flag.BoolVar(&simplified, "simplified", false, "Simplify the output to avoid listing the same codes separately for primary/secondary/death?")
	flag.BoolVar(&override, "override", false, "Force run, even if this tool thinks your tabfile is inadequate?")
	flag.BoolVar(&verbose, "verbose", false, "Print all ~ 2,000 fields whose dates are known?")
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

	if simplified {
		printTabEntryTabular(diseaseName, "include", tabs.AllIncluded(), dict, coding)
		printTabEntryTabular(diseaseName, "exclude", tabs.AllExcluded(), dict, coding)

		return
	}

	fmt.Fprintln(STDOUT, "Disease", diseaseName)

	fmt.Fprintln(STDOUT, "Including:")
	printTabEntry(tabs.AllIncluded(), dict, coding)

	fmt.Fprintln(STDOUT, "Excluding:")
	if len(tabs.AllExcluded()) > 0 {
		printTabEntry(tabs.AllExcluded(), dict, coding)
	} else {
		fmt.Fprintf(STDOUT, "\t(No exclusion criteria)\n")
	}

}

func printTabEntryTabular(diseaseName, prefix string, entries []TabEntry, dict map[int]UKBField, coding map[string]map[string]string) {
	seen := make(map[string]struct{})

	for _, v := range entries {
		dictEntry := dict[v.FieldID]
		codingEntry, exists := coding[dictEntry.Coding.String()]

		if _, exists := seen[dictEntry.Coding.String()]; exists {
			continue
		} else {
			seen[dictEntry.Coding.String()] = struct{}{}
		}

		// Try to print just the chunk with the ICD label or OPCS label.
		fieldText := ""

		chunks := strings.Split(dictEntry.Field, " ")
		foundChunk := false
		for _, chunk := range chunks {
			if strings.Contains(chunk, "ICD") || strings.Contains(chunk, "OPCS") {
				foundChunk = true
				fieldText = chunk
			}
		}

		if !foundChunk {
			fieldText = dictEntry.Field
		}

		if exists {
			for _, entryValue := range v.FormattedValues() {
				fmt.Fprintf(STDOUT, "%s\t%s\t%s\t%s\n", diseaseName, prefix, fieldText, codingEntry[entryValue])
			}
		} else {
			fmt.Fprintf(STDOUT, "%s\t%s\t%s\t%s\n", diseaseName, prefix, fieldText, strings.Join(v.Values, ", "))
		}

	}
}

func printTabEntrySimplified(entries []TabEntry, dict map[int]UKBField, coding map[string]map[string]string) {
	seen := make(map[string]struct{})

	for _, v := range entries {
		dictEntry := dict[v.FieldID]
		codingEntry, exists := coding[dictEntry.Coding.String()]

		if _, exists := seen[dictEntry.Coding.String()]; exists {
			continue
		} else {
			seen[dictEntry.Coding.String()] = struct{}{}
		}

		// Try to print just the chunk with the ICD label or OPCS label.
		chunks := strings.Split(dictEntry.Field, " ")
		foundChunk := false
		for _, chunk := range chunks {
			if strings.Contains(chunk, "ICD") || strings.Contains(chunk, "OPCS") {
				foundChunk = true
				fmt.Fprintf(STDOUT, "\t%s\n", chunk)
			}
		}

		if !foundChunk {
			fmt.Fprintf(STDOUT, "\t%s\n", dictEntry.Field)
		}

		if exists {
			for _, entryValue := range v.FormattedValues() {
				fmt.Fprintf(STDOUT, "\t\t%s\n", codingEntry[entryValue])
			}
			fmt.Fprintf(STDOUT, "\n")
		} else {
			fmt.Fprintf(STDOUT, "\t\t")
			fmt.Fprintf(STDOUT, "%s", strings.Join(v.Values, ", "))
			fmt.Fprintf(STDOUT, "\n")
		}

	}
	fmt.Fprintf(STDOUT, "\n")
}

func printTabEntry(entries []TabEntry, dict map[int]UKBField, coding map[string]map[string]string) {
	for _, v := range entries {
		dictEntry := dict[v.FieldID]

		fmt.Fprintf(STDOUT, "\t%s (FieldID %v) values:\n", dictEntry.Field, v.FieldID)

		codingEntry, exists := coding[dictEntry.Coding.String()]
		if exists {
			for _, entryValue := range v.FormattedValues() {
				fmt.Fprintf(STDOUT, "\t\t%s: %s\n", entryValue, codingEntry[entryValue])
			}
			fmt.Fprintf(STDOUT, "\n")
		} else {
			fmt.Fprintf(STDOUT, "\t\t")
			fmt.Fprintf(STDOUT, "%s", strings.Join(v.Values, ", "))
			fmt.Fprintf(STDOUT, "\n")
		}

	}
	fmt.Fprintf(STDOUT, "\n")
}
