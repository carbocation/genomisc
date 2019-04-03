package main

import (
	"fmt"
	"strconv"
	"sync"

	"gopkg.in/guregu/null.v3"
)

func printer(work <-chan Work, pool *sync.WaitGroup) {
	// fmt.Print(getHeader(analysis))

	for {
		select {
		case w := <-work:
			fmt.Fprintf(STDOUT, "%s\t%s\t%d\t%s\t%s\t%s\t%s\n", w.SampleID, w.Chrom, w.Pos, w.Ref, w.Alt, w.SNP, NullIntFormatter(w.Genotype))
			pool.Done()
		}
	}
}

func NullIntFormatter(n null.Int) string {
	if !n.Valid {
		return ""
	}

	return strconv.FormatInt(n.Int64, 10)
}
