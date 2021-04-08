package chrpos

import (
	"bytes"
	"embed"
	"encoding/csv"
	"fmt"
	"log"
	"strconv"

	"github.com/carbocation/pfx"
)

//go:embed lookups/*
var embeddedTemplates embed.FS

// ChunkChrRange allows you to split genomic data into chunks of size chunksize.
// Valid values for assemply are grch37 and grch38. chromosome is optional; if
// set, will only process that chromosome. startPos and endPos are optional and
// require chromosome to be set. If nonzero, chunks will only be created within
// their limits.
func ChunkChrRange(chunksize int, assembly string, chromosome string, startPos, endPos int) ([]TabixLocus, error) {
	loci, err := chrPosSlice(assembly, chromosome, false)
	if err != nil {
		return nil, fmt.Errorf("SplitChrPos: %w", err)
	}

	output := make([]TabixLocus, 0)

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

			endPoint := locationInChromosome + chunksize
			if chrEnd := int(locus.End()); endPoint > chrEnd {
				endPoint = chrEnd
			}

			output = append(output, MakeTabixLocus(
				locus.Chrom(),
				locationInChromosome,
				endPoint,
			))
		}
	}

	return output, err
}

func chrPosSlice(assembly string, chromosome string, verbose bool) ([]TabixLocus, error) {
	fileBytes, err := embeddedTemplates.ReadFile("lookups/" + assembly)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewReader(fileBytes)
	cr := csv.NewReader(buf)
	cr.Comma = '\t'
	entries, err := cr.ReadAll()
	if err != nil {
		return nil, pfx.Err(err)
	}

	loci := make([]TabixLocus, 0)
	header := make(map[string]int)

	// Iterating over the chromosomes in the embedded file
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
			if verbose {
				log.Println("Skipping chromosome", v[header["name"]])
			}
			continue
		}

		// If a specific chromosome is requested, only process that chromosome
		if chromosome != "" && chromosome != v[header["name"]] {
			if verbose {
				log.Println("Ignoring chromosome", v[header["name"]])
			}
			continue
		}

		end, err := strconv.Atoi(v[header["chromEnd"]])
		if err != nil {
			return nil, err
		}
		loci = append(loci, MakeTabixLocus(v[header["name"]], 0, end))
	}

	return loci, nil
}
