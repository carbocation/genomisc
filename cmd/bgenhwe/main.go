package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/carbocation/genomisc/hwecgo"

	"github.com/carbocation/bgen"
	_ "github.com/mattn/go-sqlite3"
)

const (
	// Columns in the SNP file
	SNP = iota
	CHR
)

// Compute HWE chi square values for all SNPs
func main() {
	bgenPath, snpfile, bgiPath, sampleFile, sampleIDFile := "", "", "", "", ""
	flag.StringVar(&bgenPath, "bgen", "", "Path to the BGEN file (if iterating over the full BGEN) or a template with %s in place of its chromosome number.")
	flag.StringVar(&bgiPath, "bgi", "", "Path to the BGEN index. If blank, will assume it's the BGEN path suffixed with .bgi")
	flag.StringVar(&snpfile, "snps", "", "Tab-delimited SNP file containing rsid and chromosome (in that order). If blank, and a proper BGEN is passed, then the full BGEN will be parsed.")
	flag.StringVar(&sampleFile, "sample", "", "File that maps samples to the blank rows in the BGEN. Must be in the Oxford .sample file format.")
	flag.StringVar(&sampleIDFile, "sample_ids", "", "File that has one sample ID per row. (A subset of the IDs in the sample file.) Must have a header row (or will skip your first entry).")
	flag.Parse()

	if bgiPath == "" {
		bgiPath = bgenPath + ".bgi"
	}

	if bgenPath == "" {
		flag.PrintDefaults()
		log.Fatalln()
	}

	subset := false
	var err error
	shouldCount := make([]bool, 0)
	if sampleFile != "" && sampleIDFile != "" {
		shouldCount, subset, err = SampleLookup(sampleFile, sampleIDFile)
		if err != nil {
			log.Fatalln(err)
		}
	}

	fmt.Printf("SNP\tCHR\tBP\tA1\tA2\tAC\tMAF\tAA\tAa\taa\tHWE_Exact_P\n")

	if snpfile != "" {
		if !strings.Contains(bgenPath, "%s") {
			log.Println("Iterating over a list of SNPs, but passed a full bgen path instead of a BGEN template path with %s in place of the chromosome identifier")
			flag.PrintDefaults()
			os.Exit(1)
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

		for _, row := range allSites {
			func(row []string) {
				rsID := row[SNP]
				bgPath := fmt.Sprintf(bgenPath, row[CHR])
				bgiPath := bgPath + ".bgi"

				bg, err := bgen.Open(bgPath)
				if err != nil {
					log.Fatalln(err)
				}
				defer bg.Close()

				bgi, err := bgen.OpenBGI(bgiPath)
				if err != nil {
					log.Fatalln(err)
				}
				defer bgi.Close()
				bgi.Metadata.FirstThousandBytes = nil
				// log.Printf("%+v\n", *bgi.Metadata)

				rdr := bg.NewVariantReader()

				idx, err := FindOneVariant(bgi, rsID)
				if err != nil {
					log.Fatalln(err)
				}

				variant := rdr.ReadAt(int64(idx.FileStartPosition))

				handleVariant(variant, subset, shouldCount)
			}(row)
		}

		return
	}

	// Else just iterating over a full bgen

	bg, err := bgen.Open(bgenPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer bg.Close()

	bgi, err := bgen.OpenBGI(bgiPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer bgi.Close()
	bgi.Metadata.FirstThousandBytes = nil
	log.Printf("%+v\n", *bgi.Metadata)

	rdr := bg.NewVariantReader()

	for {
		variant := rdr.Read()
		if err := rdr.Error(); err != nil {
			log.Fatalln(err)
		} else if variant == nil {
			break
		}

		handleVariant(variant, subset, shouldCount)
	}
}

func SampleLookup(sampleFile, sampleIDFile string) ([]bool, bool, error) {
	subset := false
	if sampleFile == "" || sampleIDFile == "" {
		return nil, subset, nil
	}

	subset = true

	sampleKeyF, err := os.Open(sampleFile)
	if err != nil {
		return nil, subset, err
	}

	sampleKeyCSV := csv.NewReader(sampleKeyF)
	sampleKeyCSV.Comma = ' '

	recs, err := sampleKeyCSV.ReadAll()
	if err != nil {
		return nil, subset, err
	}
	realRecs := make([][]string, 0, len(recs))
	truthMap := make([]bool, 0, len(recs))
	lookups := make(map[string]int)
	for i, line := range recs {
		if i <= 1 {
			continue
		}
		truthMapI := i - 2

		realRecs = append(realRecs, line)
		truthMap = append(truthMap, false)
		lookups[line[0]] = truthMapI
	}

	// Set the values to true if we want that sample
	subsetKeyF, err := os.Open(sampleIDFile)
	if err != nil {
		return nil, subset, err
	}
	subsetKeyCSV := csv.NewReader(subsetKeyF)
	subsetRecs, err := subsetKeyCSV.ReadAll()
	if err != nil {
		return nil, subset, err
	}
	for _, sample := range subsetRecs {
		truthMap[lookups[sample[0]]] = true
	}

	return truthMap, subset, nil
}

func handleVariant(variant *bgen.Variant, subset bool, shouldCount []bool) {
	ac, _, minaf, AAf, Aaf, aaf, exactP := ComputeHWEChiSq(variant.SampleProbabilities, subset, shouldCount)

	fmt.Printf("%s\t%s\t%d\t%s\t%s\t%.6e\t%.3e\t%.3e\t%.3e\t%.3e\t%.3e\n", variant.RSID, variant.Chromosome, variant.Position, variant.Alleles[0], variant.Alleles[1], ac, minaf, AAf, Aaf, aaf, exactP)
}

// ComputeHWEChiSq calculates the Hardy-Weinberg equilibrium chi square value at
// a given site, based on the observed and expected alleles.
func ComputeHWEChiSq(samples []bgen.SampleProbability, subset bool, shouldCount []bool) (ac, majaf, minaf, AAf, Aaf, aaf, exactP float64) {
	// Genotype count observations
	AA, Aa, aa := 0.0, 0.0, 0.0
	for i, v := range samples {
		if subset {
			if !shouldCount[i] {
				continue
			}
		}

		AA += v.Probabilities[0]
		Aa += v.Probabilities[1]
		aa += v.Probabilities[2]
	}

	// Observed N (sample count) may be smaller than the number of samples
	N := AA + Aa + aa

	// Note: we truncate to integer here
	exactP = hwecgo.Exact(int64(AA), int64(Aa), int64(aa))

	AlleleCount := 2.0 * N

	// Allele frequencies depends on the allele count, not the sample count
	A := (2.0*AA + Aa) / (AlleleCount)
	a := (Aa + 2.0*aa) / (AlleleCount)

	// Assign AF to major or minor correctly
	majaf = A
	minaf = a
	if majaf < minaf {
		minaf, majaf = majaf, minaf
	}

	return AlleleCount,
		majaf,
		minaf,
		AA / N,
		Aa / N,
		aa / N,
		exactP
}

func FindOneVariant(bgi *bgen.BGIIndex, rsID string) (bgen.VariantIndex, error) {
	row := bgen.VariantIndex{}
	if err := bgi.DB.Get(&row, "SELECT * FROM Variant WHERE rsid=? LIMIT 1", rsID); err != nil { // ORDER BY file_start_position ASC
		return row, err
	}

	return row, nil
}
