package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/carbocation/pfx"
)

func ReadSNPsnap(fileName string) (Results, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return Results{}, err
	}
	defer f.Close()

	cr := csv.NewReader(f)
	cr.Comment = '#'
	cr.Comma = '\t'
	lines, err := cr.ReadAll()
	if err != nil {
		return Results{}, err
	}

	if len(lines) < 1 {
		return Results{}, fmt.Errorf("0 lines in SNPsnap file")
	}
	if len(lines[0]) < 1 {
		return Results{}, fmt.Errorf("First line of SNPsnap file has no entries")
	}

	// Make one permutation per column
	permutations := make([]Permutation, len(lines[0]))
	for k := range permutations {
		// Initialize
		permutations[k].Loci = make([]Locus, len(lines)-1)
	}

	// Each row is a SNP
Outer:
	for i, row := range lines {
		if len(row) < 1 {
			continue
		}

		if i == 0 {
			continue
		}

		// Each column is a different permutation
		for j, col := range row {
			parts := strings.Split(col, ":")
			if len(parts) < 2 {
				log.Printf("%s had %d parts, expected 2. Skipping\n", col, len(parts))
				continue Outer
			}
			intpos, err := strconv.Atoi(parts[1])
			if err != nil {
				log.Println("Len of bad col:", len(col))
				return Results{}, pfx.Err(err)
			}
			locus := Locus{Index: j, Chromosome: parts[0], Position: intpos}
			permutations[j].Loci[i-1] = locus
		}
	}

	return Results{Permutations: permutations}, nil
}
