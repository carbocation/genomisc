// clinvar2bigquery is designed to take clinvar data, filter it using criteria
// that are useful to my line of work, and then format it so that it can be
// ingested in bigquery. Specifically, this ingests the file located at
// https://ftp.ncbi.nlm.nih.gov/pub/clinvar/xml/ClinVarFullRelease_00-latest.xml.gz
//
// See
// https://ftp.ncbi.nih.gov/pub/clinvar/vcf_GRCh38/archive_2.0/2018/README_VCF.txt
// for info on interpreting the VCF fields.
package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/carbocation/vcfgo"
)

var (
	BufferSize = 4096 * 32
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
)

func main() {
	var vcfFile string
	flag.StringVar(&vcfFile, "result", "", "A ClinVar release file, e.g., https://ftp.ncbi.nlm.nih.gov/pub/clinvar/xml/ClinVarFullRelease_00-latest.xml.gz")
	flag.Parse()

	if vcfFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Println("Started running at", time.Now())
	defer func() {
		log.Println("Completed at", time.Now())
	}()

	fraw, err := os.Open(vcfFile)
	if err != nil {
		log.Fatalf("Error opening VCF: %s\n", err)
	}
	defer fraw.Close()

	var f io.Reader
	f, err = gzip.NewReader(fraw)
	if err != nil {
		f = fraw
	}

	// Buffered reader
	buffRead := bufio.NewReaderSize(f, BufferSize)

	rdr, err := vcfgo.NewReader(buffRead, true) // Lazy genotype parsing, so we can avoid doing it on the main thread
	if err != nil {
		log.Printf("Invalid VCF. Invalid features include:\n%s\n", err)
		if rdr != nil {
			log.Println("Attempting to continue.")
			rdr.Clear()
		} else {
			log.Println("VCF reader could not be initialized due to the errors. Exiting.")
			return
		}
	}
	if err := rdr.Error(); err != nil {
		log.Println("Invalid VCF. Attempting to continue. Invalid features include:")
		log.Println(err)
		rdr.Clear()
	}

	if err := ReadAllVCF(rdr); err != nil {
		log.Fatalln(err)
	}
}

func ReadAllVCF(rdr *vcfgo.Reader) error {
	if rdr == nil {
		panic("Nil reader")
	}

	log.Println(rdr.Header.Extras)

	fmt.Println(header())

	i := 0
	skipped := 0
	for ; ; i++ {
		variant := rdr.Read()
		if variant == nil {
			log.Println("Finished")
			break
		}

		if i%50000 == 0 {
			log.Printf("Processed %d variants. Last %s:%d\n", i, variant.Chrom(), variant.Pos)
		}

		if variant.Filter != "PASS" && variant.Filter != "." {
			// Skip non-pass filter variants if toggled
			continue
		}

		// Interpretations may be made on a single variant or a set of variants,
		// such as a haplotype. Variants that have only been interpreted as part
		// of a set of variants (i.e. no direct interpretation for the variant
		// itself) are considered "included" variants. The VCF files include
		// both variants with a direct interpretation and included variants.
		// Included variants do not have an associated disease (CLNDN, CLNDISDB)
		// or a clinical significance (CLNSIG). Instead there are three tags are
		// specific to the included variants - CLNDNINCL, CLNDISDBINCL, and
		// CLNSIGINCL (see below).

		// So, skip anything without CLNSIG, because lacking CLNSIG implies that
		// we are looking at an "included" variant which has not had a clinical
		// interpretation.
		if v, err := variant.Info().Get(("CLNSIG")); err != nil || v == "" || v == nil {
			// log.Println("No CLNSIG data - skipping")
			skipped++
			continue
		}

		for alleleID := range variant.Alt() {
			fmt.Println(worker(variant, alleleID))
		}
	}

	log.Printf("Processed %d variants.\n", i)
	log.Printf("Skipped %d variants.\n", skipped)

	if err := rdr.Error(); err != nil {
		return err
	}

	return nil
}
