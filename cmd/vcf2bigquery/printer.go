package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/guregu/null.v3"
)

func printer(work <-chan Work, pool *sync.WaitGroup, sampleFields []string) {
	// fmt.Print(getHeader(analysis))

	if len(sampleFields) > 0 {
		for {
			select {
			case w := <-work:

				sb := make([]string, 0, len(sampleFields))
				for _, v := range w.SampleFields {
					sb = append(sb, NullStringFormatter(v))
				}

				fmt.Fprintf(STDOUT, "%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\n", w.SampleID, w.Chrom, w.Pos, w.Ref, w.Alt, w.SNP, NullIntFormatter(w.Genotype), strings.Join(sb, "\t"))
				pool.Done()
			}
		}
	} else {
		for {
			select {
			case w := <-work:
				fmt.Fprintf(STDOUT, "%s\t%s\t%d\t%s\t%s\t%s\t%s\n", w.SampleID, w.Chrom, w.Pos, w.Ref, w.Alt, w.SNP, NullIntFormatter(w.Genotype))
				pool.Done()
			}
		}
	}
}

func NullIntFormatter(n null.Int) string {
	if !n.Valid {
		return ""
	}

	return strconv.FormatInt(n.Int64, 10)
}

func NullStringFormatter(n null.String) string {
	if !n.Valid {
		return ""
	}

	return n.String
}
