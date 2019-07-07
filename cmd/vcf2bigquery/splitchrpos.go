package main

import (
	"broad/ghgwas/lib/vcf"
	"bytes"
	"encoding/csv"
	"log"
	"strconv"

	"github.com/gobuffalo/packr/v2"
)

func chrPosSlice(assembly string, chromosome string) ([]vcf.TabixLocus, error) {
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

		// Only process autosomes, chrX, and chrY
		_, err := strconv.Atoi(v[header["name"]])
		if err != nil && v[header["name"]] != "X" && v[header["name"]] != "Y" {
			log.Println("Skipping chromosome", v[header["name"]])
			continue
		}

		// If a specific chromosome is requested, only process that chromosome
		if chromosome != "" && chromosome != v[header["name"]] {
			log.Println("Ignoring chromosome", v[header["name"]])
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

func SplitChrPos(chunksize int, assembly string, chromosome string, startPos, endPos int) ([]vcf.TabixLocus, error) {
	loci, err := chrPosSlice(assembly, chromosome)

	output := make([]vcf.TabixLocus, 0)

	start := 0
	if chromosome != "" {
		start = startPos
	}

	for _, locus := range loci {

		end := int(locus.End())
		if chromosome != "" && endPos != 0 {
			end = endPos
		}

		for locationInChromosome := start; locationInChromosome < end; locationInChromosome += chunksize {
			output = append(output, vcf.MakeTabixLocus(
				locus.Chrom(),
				locationInChromosome,
				locationInChromosome+chunksize,
			))
		}
	}

	return output, err
}
