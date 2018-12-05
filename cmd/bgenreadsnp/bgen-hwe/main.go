package main

import (
	"flag"
	"fmt"
	"log"
	"math"

	"github.com/carbocation/bgen"
	_ "github.com/mattn/go-sqlite3"
)

// Compute HWE chi square values for all SNPs
func main() {
	bgenPath, rsID, bgiPath := "", "", ""
	flag.StringVar(&bgenPath, "bgen", "", "Path to the BGEN file")
	flag.StringVar(&bgiPath, "bgi", "", "Path to the BGEN index. If blank, will assume it's the BGEN path suffixed with .bgi")
	flag.StringVar(&rsID, "snp", "", "SNP ID. If blank, this will run on all SNPs found in the BGEN.")
	flag.Parse()

	if bgiPath == "" {
		bgiPath = bgenPath + ".bgi"
	}

	if bgenPath == "" {
		flag.PrintDefaults()
		log.Fatalln()
	}

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
	fmt.Printf("SNP\tCHR\tBP\tA1\tA2\tHWEChiSq\tMAF\n")

	if rsID != "" {
		idx, err := FindOneVariant(bgi, rsID)
		if err != nil {
			log.Fatalln(err)
		}

		variant := rdr.ReadAt(int64(idx.FileStartPosition))

		hwe, _, minaf := ComputeHWEChiSq(variant.Probabilities.SampleProbabilities)
		fmt.Printf("%s\t%s\t%d\t%s\t%s\t%.3f\t%.3e\n", variant.RSID, variant.Chromosome, variant.Position, variant.Alleles[0], variant.Alleles[1], hwe, minaf)

		return
	}

	for {
		variant := rdr.Read()
		if err := rdr.Error(); err != nil {
			log.Fatalln(err)
		} else if variant == nil {
			break
		}

		hwe, _, minaf := ComputeHWEChiSq(variant.Probabilities.SampleProbabilities)
		fmt.Printf("%s\t%s\t%d\t%s\t%s\t%.3f\t%.3e\n", variant.RSID, variant.Chromosome, variant.Position, variant.Alleles[0], variant.Alleles[1], hwe, minaf)
	}
}

// ComputeHWEChiSq calculates the Hardy-Weinberg equilibrium chi square value at
// a given site, based on the observed and expected alleles.
func ComputeHWEChiSq(samples []*bgen.SampleProbability) (chisquare, majaf, minaf float64) {
	N := float64(len(samples))

	// Genotype count observations
	AA, Aa, aa := 0.0, 0.0, 0.0
	for _, v := range samples {
		AA += v.Probabilities[0]
		Aa += v.Probabilities[1]
		aa += v.Probabilities[2]
	}

	// Allele frequencies
	A := (2.0*AA + Aa) / (2 * N)
	a := (Aa + 2.0*aa) / (2 * N)

	// Genotype count expectations
	eAA := A * A * N
	eAa := 2.0 * A * a * N
	eaa := a * a * N

	// Assign AF to major or minor correctly
	majaf = A
	minaf = a
	if majaf < minaf {
		minaf, majaf = majaf, minaf
	}

	// ChiSquare
	return math.Pow(eAA-AA, 2)/eAA +
			math.Pow(eAa-Aa, 2)/eAa +
			math.Pow(eaa-aa, 2)/eaa,
		majaf,
		minaf
}

func FindOneVariant(bgi *bgen.BGIIndex, rsID string) (bgen.VariantIndex, error) {
	row := bgen.VariantIndex{}
	if err := bgi.DB.Get(&row, "SELECT * FROM Variant WHERE rsid=? LIMIT 1", rsID); err != nil { // ORDER BY file_start_position ASC
		return row, err
	}

	return row, nil
}
