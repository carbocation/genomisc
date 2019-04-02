package main

import (
	"broad/ghgwas/lib/vcf"
	"bytes"
	"encoding/csv"
	"log"
	"strconv"

	"github.com/gobuffalo/packr/v2"
)

func chrPosSlice(assembly string) ([]vcf.TabixLocus, error) {
	lookups := packr.New("lookups", "./lookups")

	file, err := lookups.Find(assembly)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(file)
	cr := csv.NewReader(buf)
	cr.Comma = '\t'
	entries, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}

	loci := make([]vcf.TabixLocus, 0)
	header := make(map[string]int)
	for i, v := range entries {
		if i == 0 {
			for key, name := range v {
				header[name] = key
			}
			continue
		}

		// Only process autosomes
		_, err := strconv.Atoi(v[header["name"]])
		if err != nil {
			log.Println("Skipping chromosome", v[header["name"]])
			continue
		}

		end, err := strconv.Atoi(v[header["chromEnd"]])
		if err != nil {
			return nil, err
		}
		loci = append(loci, vcf.MakeTabixLocus(v[header["name"]], 0, end))
	}

	return loci, nil
}

func SplitChrPos(chunksize int, assembly string) ([]vcf.TabixLocus, error) {
	loci, err := chrPosSlice(assembly)

	output := make([]vcf.TabixLocus, 0)

	for _, chromosome := range loci {

		for locationInChromosome := 0; locationInChromosome < int(chromosome.End()); locationInChromosome += chunksize {
			output = append(output, vcf.MakeTabixLocus(
				chromosome.Chrom(),
				locationInChromosome,
				locationInChromosome+chunksize,
			))
		}
	}

	return output, err
}