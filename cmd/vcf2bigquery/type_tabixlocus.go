package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

type TabixLocus struct {
	chrom string
	start int
	end   int
}

func MakeTabixLocus(chrom string, start, end int) TabixLocus {
	return TabixLocus{chrom, start, end}
}

func (tl TabixLocus) Chrom() string {
	return tl.chrom
}

func (tl TabixLocus) Start() uint32 {
	return uint32(tl.start)
}

func (tl TabixLocus) End() uint32 {
	return uint32(tl.end)
}

func TabixLociFromPath(filepath string) ([]TabixLocus, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(f)
	r.Comma = '\t'
	regions, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	tabixLoci := make([]TabixLocus, 0, len(regions))
	for _, v := range regions {
		if len(v) < 3 {
			continue
		}
		start, err := strconv.Atoi(v[1])
		if err != nil {
			return nil, err
		}
		end, err := strconv.Atoi(v[2])
		if err != nil {
			return nil, err
		}
		t := MakeTabixLocus(v[0], start, end)
		tabixLoci = append(tabixLoci, t)
	}

	if len(tabixLoci) < 1 {
		return nil, fmt.Errorf("No valid loci identified in %v", filepath)
	}

	return tabixLoci, nil
}
