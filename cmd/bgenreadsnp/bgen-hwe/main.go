package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/carbocation/bgen"
	"github.com/carbocation/genomisc"
	_ "github.com/mattn/go-sqlite3"
)

// Compute HWE chi square values for all SNPs
func main() {
	bgenPath, rsID, bgiPath, sampleFile, sampleIDFile := "", "", "", "", ""
	flag.StringVar(&bgenPath, "bgen", "", "Path to the BGEN file")
	flag.StringVar(&bgiPath, "bgi", "", "Path to the BGEN index. If blank, will assume it's the BGEN path suffixed with .bgi")
	flag.StringVar(&rsID, "snp", "", "SNP ID. If blank, this will run on all SNPs found in the BGEN.")
	flag.StringVar(&rsID, "snp", "", "SNP ID. If blank, this will run on all SNPs found in the BGEN.")
	flag.StringVar(&sampleFile, "sample", "", "File that maps samples to the blank rows in the BGEN. Must be in .sample file format.")
	flag.StringVar(&sampleIDFile, "sample_ids", "", "File that has one sample ID per row. (A subset of the IDs in the sample file.)")

	if bgiPath == "" {
		bgiPath = bgenPath + ".bgi"
	}

	if bgenPath == "" {
		flag.PrintDefaults()
		log.Fatalln()
	}

	subset := false
	shouldCount := make([]bool, 0)
	var err error
	if sampleFile != "" && sampleIDFile != "" {
		shouldCount, subset, err = SampleLookup(sampleFile, sampleIDFile)
		if err != nil {
			log.Fatalln(err)
		}
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
	fmt.Printf("SNP\tCHR\tBP\tA1\tA2\tHWEChiSq\tMAF\tAA\tAa\taa\n")

	if rsID != "" {
		idx, err := FindOneVariant(bgi, rsID)
		if err != nil {
			log.Fatalln(err)
		}

		variant := rdr.ReadAt(int64(idx.FileStartPosition))

		handleVariant(variant, subset, shouldCount)

		return
	}

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
	delim := genomisc.DetermineDelimiter(sampleKeyF)
	sampleKeyCSV.Comma = delim

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
	delim = genomisc.DetermineDelimiter(subsetKeyF)
	subsetKeyCSV.Comma = delim
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
	hwe, _, minaf, AAf, Aaf, aaf := ComputeHWEChiSq(variant.Probabilities.SampleProbabilities, subset, shouldCount)
	fmt.Printf("%s\t%s\t%d\t%s\t%s\t%.3e\t%.3e\t%.3e\t%.3e\t%.3e\n", variant.RSID, variant.Chromosome, variant.Position, variant.Alleles[0], variant.Alleles[1], hwe, minaf, AAf, Aaf, aaf)
}

// ComputeHWEChiSq calculates the Hardy-Weinberg equilibrium chi square value at
// a given site, based on the observed and expected alleles.
func ComputeHWEChiSq(samples []*bgen.SampleProbability, subset bool, shouldCount []bool) (chisquare, majaf, minaf, AAf, Aaf, aaf float64) {
	Nint := len(samples)
	if subset {
		Nint = 0
	}

	// Genotype count observations
	AA, Aa, aa := 0.0, 0.0, 0.0
	for i, v := range samples {
		if subset {
			if !shouldCount[i] {
				continue
			}

			Nint++
		}

		AA += v.Probabilities[0]
		Aa += v.Probabilities[1]
		aa += v.Probabilities[2]
	}

	N := float64(Nint)

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
		minaf,
		AA / N,
		Aa / N,
		aa / N
}

func FindOneVariant(bgi *bgen.BGIIndex, rsID string) (bgen.VariantIndex, error) {
	row := bgen.VariantIndex{}
	if err := bgi.DB.Get(&row, "SELECT * FROM Variant WHERE rsid=? LIMIT 1", rsID); err != nil { // ORDER BY file_start_position ASC
		return row, err
	}

	return row, nil
}
