// clinvar2bigquery is designed to take clinvar data, filter it using criteria
// that are useful to my line of work, and then format it so that it can be
// ingested in bigquery. Specifically, this ingests the file located at
// https://ftp.ncbi.nlm.nih.gov/pub/clinvar/xml/ClinVarFullRelease_00-latest.xml.gz
package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	var result string
	flag.StringVar(&result, "result", "", "A ClinVar release file, e.g., https://ftp.ncbi.nlm.nih.gov/pub/clinvar/xml/ClinVarFullRelease_00-latest.xml.gz")
	flag.Parse()

	if result == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Println("Started running at", time.Now())
	defer func() {
		log.Println("Completed at", time.Now())
	}()

	f, err := os.Open(result)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	d := xml.NewDecoder(f)
Loop:
	for {
		tok, err := d.Token()
		if tok == nil || err == io.EOF {
			// EOF means we're done.
			break
		} else if err != nil {
			log.Fatalf("Error decoding token: %s", err)
		}

		switch ty := tok.(type) {
		case xml.StartElement:
			if ty.Name.Local == "ClinVarSet" {

				// If this is a start element named "ClinVarSet", parse this element
				// fully.
				var rec ClinVarSet
				if err = d.DecodeElement(&rec, &ty); err != nil {
					log.Fatalf("Error decoding item: %s", err)
				}

				// Only print active records
				if rec.RecordStatus != "current" {
					continue
				}

				// Keep these fields:
				// fmt.Println(rec.ID)

				fmt.Println(rec.ReferenceClinVarAssertion)
			} else if ty.Name.Local == "VariationArchive" {

				// If this is a start element named "VariationArchive", parse
				// this element fully.
				var rec VariationArchive
				if err = d.DecodeElement(&rec, &ty); err != nil {
					log.Fatalf("Error decoding item: %s", err)
				}

				spew.Dump(rec.InterpretedRecord)
				// rec.IncludedRecord.SimpleAllele.Interpretations.Interpretation.Text
				break Loop

			}
		default:
		}
	}

}
