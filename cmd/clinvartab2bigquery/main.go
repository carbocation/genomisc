// clinvar2bigquery is designed to take clinvar data, filter it using criteria
// that are useful to my line of work, and then format it so that it can be
// ingested in bigquery. Specifically, this ingests the file located at
// https://ftp.ncbi.nlm.nih.gov/pub/clinvar/tab_delimited/variant_summary.txt.gz
//
// See
// https://ftp.ncbi.nlm.nih.gov/pub/clinvar/tab_delimited/README
// for info on interpreting the fields.
package main

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/carbocation/genomisc/compileinfoprint"
	"github.com/carbocation/pfx"
)

const YMDFormat = "2006/01/02"

var (
	BufferSize = 4096 * 32
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
)

func main() {
	defer STDOUT.Flush()

	var tabFile, assembly string
	flag.StringVar(&tabFile, "result", "", "A ClinVar release file in tab format, e.g., https://ftp.ncbi.nlm.nih.gov/pub/clinvar/tab_delimited/variant_summary.txt.gz")
	flag.StringVar(&assembly, "assembly", "", "Human genome assembly. Currently must be one of 'GRCh37' or 'GRCh38'")
	flag.Parse()

	loc, err := time.LoadLocation("")
	if err != nil {
		log.Fatalln(err)
	}

	if tabFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if strings.ToLower(assembly) == "grch37" {
		assembly = "GRCh37"
	} else if strings.ToLower(assembly) == "grch38" {
		assembly = "GRCh38"
	} else {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Println("Started running at", time.Now())
	defer func() {
		log.Println("Completed at", time.Now())
	}()

	fraw, err := os.Open(tabFile)
	if err != nil {
		log.Fatalf("Error opening VCF: %s\n", err)
	}
	defer fraw.Close()

	var f io.Reader
	f, err = gzip.NewReader(fraw)
	if err != nil {

		// Not gzipped? close and reopen

		fraw.Close()

		f, err = os.Open(tabFile)
		if err != nil {
			log.Fatalf("Error opening VCF: %s\n", err)
		}
		defer fraw.Close()
	}

	c := csv.NewReader(f)
	c.Comma = '\t'

	header := []string{"siteid", "CHR", "POS", "REF", "ALT", "stars", "year"}
	hid := make(map[string]int)

	for i := 0; ; i++ {
		cols, err := c.Read()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln(err)
		}

		if i == 0 {
			r := strings.NewReplacer(
				" ", "_",
				"#", "",
				"(", "",
				")", "",
				"/", "_",
			)

			// cols[0] = strings.ReplaceAll(cols[0], "#", "")
			for k := range cols {
				// cols[k] = strings.ReplaceAll(cols[k], " ", "_")

				cols[k] = r.Replace(cols[k])
			}

			header = append(header, cols...)
			for k, v := range cols {
				hid[v] = k
			}

			fmt.Fprintln(STDOUT, strings.Join(header, "\t"))
			continue
		}

		// Use a proper empty field instead of a dash to represent empty
		for k := range cols {
			if cols[k] == "-" {
				cols[k] = ""
			}
		}

		// We need to know the last eval because we can't rely on very old evals:
		if cols[hid["LastEvaluated"]] == "" {
			continue
		}

		// Subset to your chosen assembly
		if cols[hid["Assembly"]] != assembly {
			continue
		}

		// Ignore "included" alleles that have no specific assessment and other
		// entries that have no values for clinical significance at all for this
		// variant or set of variants; used for the "included" variants that are
		// only in ClinVar because they are included in a haplotype or genotype
		// with an interpretation
		if cols[hid["ClinSigSimple"]] == "-1" {
			continue
		}

		// Fix time to a cleaner format
		dt := ParseDate(cols[hid["LastEvaluated"]])
		cols[hid["LastEvaluated"]] = dt.Format(YMDFormat)
		year := dt.Format("2006")

		// Skip any data that hasn't been re-evaluated since the 2015 ACMG
		// guidelines were updated, and as shorthand we use 2016/01/01 as the
		// cutoff for that. See PMID 31752965 "Is 'Likely Pathogenic' Really 90%
		// Likely? Reclassification Data in ClinVar" by Harrison & Rehm.
		if dt.Before(time.Date(2016, time.January, 1, 0, 0, 0, 0, loc)) {
			continue
		}

		// See https://www.ncbi.nlm.nih.gov/clinvar/docs/review_status/ for the rules
		stars := "0"
		switch cols[hid["ReviewStatus"]] {
		case "practice guideline":
			stars = "4"
		case "reviewed by expert panel":
			stars = "3"
		case "criteria provided, multiple submitters, no conflicts":
			stars = "2"
		case "criteria provided, single submitter", "criteria provided, conflicting interpretations":
			stars = "1"
		}

		siteid := fmt.Sprintf("%s:%s:%s:%s", cols[hid["Chromosome"]], cols[hid["Start"]], cols[hid["ReferenceAllele"]], cols[hid["AlternateAllele"]])

		line := append(
			[]string{
				siteid,
				cols[hid["Chromosome"]],
				cols[hid["Start"]],
				cols[hid["ReferenceAllele"]],
				cols[hid["AlternateAllele"]],
				stars,
				year,
			},
			cols...,
		)

		fmt.Fprintln(STDOUT, strings.Join(line, "\t"))

	}

}

func ParseDate(d string) time.Time {
	if d == "" {
		return time.Time{}
	}

	dt, err := time.Parse("Jan 02, 2006", d)
	if err != nil {
		log.Println(pfx.Err(err))
		return time.Time{}
	}

	return dt
}
