// Convert BOLT-LMM input into input satisfactory for ldhub
package main

import (
	"compress/gzip"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/carbocation/genomisc"
	_ "github.com/carbocation/genomisc/compileinfoprint"
)

const (
	SNP     = 0
	ALLELE0 = 5
	ALLELE1 = 4
	BETA    = 10
	SE      = 11
	P       = 15
)

func main() {
	var filename string
	var sampleSize int

	flag.StringVar(&filename, "file", "", "Path to file containg BOLT-LMM 2.3.4-style summary statistics. P_BOLT_LMM will be used as the P-value column.")
	flag.IntVar(&sampleSize, "samplesize", 0, "Sample size used for this GWAS")
	flag.Parse()

	fmt.Fprintln(os.Stderr, `As of the time of the creation of this tool, the BOLT-LMM header was as follows:

# SNP	CHR	BP	GENPOS	ALLELE1	ALLELE0	A1FREQ	INFO	CHISQ_LINREG	P_LINREG	BETA	SE	CHISQ_BOLT_LMM_INF	P_BOLT_LMM_INF	CHISQ_BOLT_LMM	P_BOLT_LMM

It produces an output file with a header as follows:
# snpid	A1	A2	Zscore	N	P-value

Note that this tool maps ALLELE0 -> A1 and ALLELE1 -> A2`)

	if filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(filename, sampleSize); err != nil {
		log.Fatalln(err)
	}

	log.Println("bolt2ldhub completed")
}

func run(filename string, sampleSize int) error {
	f, err := os.Open(genomisc.ExpandHome(filename))
	if err != nil {
		return err
	}
	defer f.Close()

	var r io.Reader
	if strings.HasSuffix(filename, ".gz") {
		log.Println("Parsing a gzipped file")
		r, err = gzip.NewReader(f)
		if err != nil {
			return err
		}
	} else {
		r = f
	}

	c := csv.NewReader(r)
	c.Comma = '\t'
	c.Comment = '#'

	for i := 0; ; i++ {
		row, err := c.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if i == 0 {
			fmt.Println(strings.Join([]string{"snpid", "A1", "A2", "Zscore", "N", "P-value"}, "\t"))
			continue
		}

		beta, err := strconv.ParseFloat(row[BETA], 64)
		if err != nil {
			return err
		}

		se, err := strconv.ParseFloat(row[SE], 64)
		if err != nil {
			return err
		}

		// We exclude sites with impossible standard errors
		if se <= 0 {
			continue
		}

		z := strconv.FormatFloat(beta/se, 'G', -1, 64)

		fmt.Println(strings.Join([]string{
			row[SNP],
			row[ALLELE0],
			row[ALLELE1],
			z,
			strconv.Itoa(sampleSize),
			row[P],
		}, "\t"))
	}

	return nil
}
