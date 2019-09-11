package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/carbocation/bgen"
	_ "github.com/mattn/go-sqlite3"
)

const (
	// Columns in the SNP file
	SNP = iota
	CHR
)

const (
	// MissingAltAlleleIndicatorString is the string that we will use to
	// represent a missing site. For the purposes of BigQuery, the empty string
	// is most convenient. TODO: Make this value modifiable.
	MissingAltAlleleIndicatorString = ""
)

var (
	BufferSize = 4096 * 8
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
)

func main() {
	defer STDOUT.Flush()

	var bgenTemplatePath, assembly, snpfile string

	flag.StringVar(&bgenTemplatePath, "bgen-template", "", "Templated full path to bgens, with %s in place of its chromosome number. Index file is assumed to be .bgi at the same path.")
	flag.StringVar(&snpfile, "snps", "", "Tab-delimited SNP file containing rsid and chromosome")
	flag.StringVar(&assembly, "assembly", "", "Name of assembly. Must be grch37 or grch38.")
	flag.Parse()

	if bgenTemplatePath == "" {
		flag.PrintDefaults()
		log.Fatalln("Please specify --bgen-template")
	}

	if snpfile == "" {
		flag.PrintDefaults()
		log.Fatalln("Please specify --snps")
	}

	if assembly != "grch37" && assembly != "grch38" {
		flag.PrintDefaults()
		log.Fatalln("Please specify assembly version")
	}

	snps, err := os.Open(snpfile)
	if err != nil {
		log.Fatalln(err)
	}

	r := csv.NewReader(snps)
	r.Comma = '\t'

	allSites, err := r.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Fprintf(STDOUT, "chr\tpos_%s\trsid\tref\talt\tsample_row_id\talt_allele_dosage\n", assembly)

	for i, row := range allSites {
		if len(row) != 2 {
			log.Printf("Skipping row %d with != 2 columns: %+v\n", i, row)
			continue
		}

		rsID := row[SNP]
		bgenPath := fmt.Sprintf(bgenTemplatePath, row[CHR])
		bgiPath := bgenPath + ".bgi"

		if err := PrintOneVariant(rsID, bgenPath, bgiPath); err != nil {
			log.Println(err)
		}
	}
}

func PrintOneVariant(rsID string, bgenPath, bgiPath string) error {
	// Open the BGEN
	bg, err := bgen.Open(bgenPath)
	if err != nil {
		return err
	}
	defer bg.Close()

	// Open the BGI
	bgi, err := bgen.OpenBGI(bgiPath)
	if err != nil {
		return err
	}
	defer bgi.Close()
	bgi.Metadata.FirstThousandBytes = nil
	log.Printf("%+v\n", *bgi.Metadata)

	// Look up the variant in the BGI
	idx := bgen.VariantIndex{}
	if err := bgi.DB.Get(&idx, "SELECT * FROM Variant WHERE rsid=? LIMIT 1", rsID); err != nil {
		return err
	}
	if idx.Chromosome == "" && idx.Position == 0 {
		// Didn't read any data
		return fmt.Errorf("'%s' was not found in index file '%s'", rsID, bgiPath)
	}

	// Read & print the variant from the BGEN
	rdr := bg.NewVariantReader()
	variant := rdr.ReadAt(int64(idx.FileStartPosition))
	if err := rdr.Error(); err != nil {
		return fmt.Errorf("'%s' error: %v", rsID, err)
	}

	fixedChromosome := FixChromosomeIfNumeric(variant.Chromosome)

	var aacText string
	for sampleFileRow, v := range variant.SampleProbabilities {
		aac := 0.0
		for allele, prob := range v.Probabilities {
			// 0 for AA
			// 1 * prob for AB
			// 2 * prob for BB
			aac += float64(allele) * prob

			// ac[allele] += prob

		}

		if v.Missing {
			aacText = MissingAltAlleleIndicatorString
		} else {
			aacText = fmt.Sprintf("%f", aac)
		}

		fmt.Fprintf(STDOUT, "%s\t%d\t%s\t%s\t%s\t%d\t%s\n", fixedChromosome, variant.Position, variant.RSID, variant.Alleles[0], variant.Alleles[1], sampleFileRow, aacText)
	}

	return nil
}

// FixChromosomeIfNumeric removes preceding zeroes from the chromosome, if one
// is present and if the chromosome name is numeric.
func FixChromosomeIfNumeric(chromosome string) string {
	number, err := strconv.ParseInt(chromosome, 10, 64)
	if err != nil {
		return chromosome
	}

	// The value is an integer. Print it without the preceding zero, if one was
	// there in the first place.
	return strconv.FormatInt(number, 10)
}
