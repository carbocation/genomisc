package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/carbocation/bgen"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	bgenPath, rsID := "", ""
	flag.StringVar(&bgenPath, "bgen", "", "Path to the BGEN file")
	flag.StringVar(&rsID, "snp", "", "SNP ID")
	flag.Parse()

	bgiPath := bgenPath + ".bgi"
	if bgiPath == ".bgi" {
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
	fmt.Printf("%+v\n", *bgi.Metadata)

	idx, err := FindOneVariant(bgi, rsID)
	if err != nil {
		log.Fatalln(err)
	}

	rdr := bg.NewVariantReader()
	variant := rdr.ReadAt(int64(idx.FileStartPosition))

	fmt.Println(rsID)
	fmt.Println(variant.Alleles)

	ac := make(map[bgen.Allele]float64)
	for _, v := range variant.SampleProbabilities {
		for allele, prob := range v.Probabilities {
			ac[variant.Alleles[allele]] += prob
		}
	}

	fmt.Println("Summed allelic dosage")
	fmt.Println(ac)
}

func FindOneVariant(bgi *bgen.BGIIndex, rsID string) (bgen.VariantIndex, error) {
	row := bgen.VariantIndex{}
	if err := bgi.DB.Get(&row, "SELECT * FROM Variant WHERE rsid=? LIMIT 1", rsID); err != nil { // ORDER BY file_start_position ASC
		return row, err
	}

	return row, nil
}
