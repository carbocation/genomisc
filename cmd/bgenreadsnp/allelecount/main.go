package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/carbocation/bgen"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	bgenPath, rsID, bgiPath := "", "", ""
	flag.StringVar(&bgenPath, "bgen", "", "Path to the BGEN file")
	flag.StringVar(&bgiPath, "bgi", "", "Path to the BGEN index. If blank, will assume it's the BGEN path suffixed with .bgi")
	flag.StringVar(&rsID, "snp", "", "SNP ID")
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
	// log.Printf("%+v\n", *bgi.Metadata)

	idx, err := FindOneVariant(bgi, rsID)
	if err != nil {
		log.Fatalln(err)
	}

	rdr := bg.NewVariantReader()
	variant := rdr.ReadAt(int64(idx.FileStartPosition))

	// fmt.Println(rsID)
	// fmt.Println(variant.Alleles)

	ac := make(map[int]float64)
	for _, v := range variant.SampleProbabilities {
		// mac := 0.0
		for allele, prob := range v.Probabilities {
			// 0 for AA
			// 1 * prob for AB
			// 2 * prob for BB
			// mac += float64(allele) * prob

			ac[allele] += prob

		}
		// fmt.Printf("%s\t%d\t%d\t\n", variant.Chromosome, variant.Position, sampleFileRow)

	}

	a0 := ac[0]*2 + ac[1]
	a1 := ac[2]*2 + ac[1]

	fmt.Printf("chr\tpos\trsid\ta0\ta1\ta0_count\ta1_count\n")
	fmt.Printf("%s\t%d\t%s\t%s\t%s\t%f\t%f\n", stripLeadingZero(variant.Chromosome), variant.Position, variant.RSID, variant.Alleles[0].String(), variant.Alleles[1].String(), a0, a1)
}

func stripLeadingZero(in string) string {
	return strings.TrimPrefix(in, "0")
}

func FindOneVariant(bgi *bgen.BGIIndex, rsID string) (bgen.VariantIndex, error) {
	row := bgen.VariantIndex{}
	if err := bgi.DB.Get(&row, "SELECT * FROM Variant WHERE rsid=? LIMIT 1", rsID); err != nil { // ORDER BY file_start_position ASC
		return row, err
	}

	return row, nil
}
