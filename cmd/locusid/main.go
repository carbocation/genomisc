package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

var snpFile, chr, pos, comma, locusIDName string
var distanceThreshold float64

func main() {

	flag.StringVar(&snpFile, "snp_file", "", "File with SNPs. Loci should be pre-sorted.")
	flag.Float64Var(&distanceThreshold, "distance_threshold", 500000, "Maximum distance to collapse into the same ID")
	flag.StringVar(&chr, "chr", "CHR", "Chromosome column.")
	flag.StringVar(&pos, "pos", "BP", "Position column.")
	flag.StringVar(&comma, "delimiter", "\t", "Delimiter.")
	flag.StringVar(&locusIDName, "locusid_name", "GlobalLocusID", "Name of 'GlobalLocusID' column.")
	flag.Parse()

	if snpFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	f, err := os.Open(snpFile)
	if err != nil {
		return err
	}

	r := csv.NewReader(f)
	r.Comma = []rune(comma)[0]

	rows, err := r.ReadAll()
	if err != nil {
		return err
	}

	if len(rows) < 1 {
		return fmt.Errorf("No entries found")
	}

	chrCol := -1
	posCol := -1
	header := rows[0]
	for i, text := range header {
		if text == chr {
			chrCol = i
		}

		if text == pos {
			posCol = i
		}
	}

	if chrCol < 0 {
		return fmt.Errorf("chrCol '%s' was not found", chr)
	}
	if posCol < 0 {
		return fmt.Errorf("posCol '%s' was not found", pos)
	}

	snps := make([][]string, 0, len(rows)-1)
	for i, v := range rows {
		if i == 0 {
			continue
		}
		v = append(v, "")
		snps = append(snps, v)
	}

	header = append(header, locusIDName)

	IDCol := len(header) - 1

	i := 1
	for _, snp := range snps {
		if snp[IDCol] == "" {
			// Assign the ID
			snp[IDCol] = strconv.Itoa(i)
			i++

			// Assign the same ID to any other SNPs that match
			for _, snp2 := range snps {
				if snp2[IDCol] == "" && atSameLocus(snp, snp2, chrCol, posCol, distanceThreshold) {
					snp2[IDCol] = snp[IDCol]
				}
			}
		}
	}

	fmt.Println(strings.Join(header, "\t"))
	for _, snp := range snps {
		fmt.Println(strings.Join(snp, "\t"))
	}

	return nil
}

func atSameLocus(snp1, snp2 []string, chrCol, posCol int, distanceThreshold float64) bool {
	if snp1[chrCol] != snp2[chrCol] {
		return false
	}

	bp1, err := strconv.ParseFloat(snp1[posCol], 64)
	if err != nil {
		return false
	}

	bp2, err := strconv.ParseFloat(snp2[posCol], 64)
	if err != nil {
		return false
	}

	if math.Abs(bp1-bp2) <= distanceThreshold {
		return true
	}

	return false
}
