package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

var expectedHeader = map[int]string{
	0:  "CHR",
	1:  "POS",
	4:  "Allele1",
	5:  "Allele2",
	7:  "AF_Allele2",
	10: "BETA",
	13: "p.value",
}

const (
	SNP                = 2
	CHR                = 0
	POS                = 1
	ALLELE1            = 4
	ALLELE0            = 5
	A1FREQ             = 7
	INFO               = 8
	CHISQ_LINREG       = 12
	P_LINREG           = 13
	BETA               = 10
	SE                 = 11
	CHISQ_BOLT_LMM_INF = 12
	P_BOLT_LMM_INF     = 13
	CHISQL_BOLT_LMM    = 12
	P_BOLT_LMM         = 13
)

func main() {
	log.Println("saige2bolt")
	fmt.Fprintln(os.Stderr,
		`Consumes a SAIGE output summary stats file and reorients it so that it behaves like a BOLT-LMM summary stats file.
  1. Zeroes preceding the CHR will be removed.
  2. Because BOLT ALLELE1 [ref/effect] is non-reference and SAIGE Allele2 [alt/effect] is non-reference, the SAIGE BETA will have its sign reversed 
and A1FREQ will be represented as 1-AF_Allele2
`)

	var saige string
	flag.StringVar(&saige, "saige", "", "Path to SAIGE result file that you want to convert to BOLT format.")
	flag.Parse()

	if saige == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	header, r, closer, err := OpenDelimited(saige)
	if err != nil {
		log.Fatalln(err)
	}
	defer closer()

	if err := validateHeader(header); err != nil {
		log.Fatalln(err)
	}

	PrintBOLTHeader()

	for {
		line, err := r.Read()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln(err)
		}

		if err := PrintLine(line); err != nil {
			log.Fatalln(err)
		}
	}
}

func PrintBOLTHeader() {
	fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		"SNP",
		"CHR",
		"POS",
		"ALLELE1",
		"ALLELE0",
		"A1FREQ",
		"INFO",
		"CHISQ_LINREG",
		"P_LINREG",
		"BETA",
		"SE",
		"CHISQ_BOLT_LMM_INF",
		"P_BOLT_LMM_INF",
		"CHISQL_BOLT_LMM",
		"P_BOLT_LMM")
}

func PrintLine(line []string) error {

	// Remove preceding zeroes from CHR
	line[CHR] = strings.TrimPrefix(line[CHR], "0")

	// TODO: Fix the column numbers in the map

	// Reverse BETA's sign
	betaVal, err := strconv.ParseFloat(line[BETA], 64)
	if err != nil {
		return err
	}
	line[BETA] = strconv.FormatFloat(-1*betaVal, 'g', -1, 64)

	// Replace AF_Allele2 with 1-AF_Allele2 (yielding A1FREQ)
	a1freq, err := strconv.ParseFloat(line[A1FREQ], 64)
	if err != nil {
		return err
	}
	line[A1FREQ] = strconv.FormatFloat(1-a1freq, 'g', -1, 64)

	_, err = fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		line[SNP],
		line[CHR],
		line[POS],
		line[ALLELE1],
		line[ALLELE0],
		line[A1FREQ],
		line[INFO],
		line[CHISQ_LINREG],
		line[P_LINREG],
		line[BETA],
		line[SE],
		line[CHISQ_BOLT_LMM_INF],
		line[P_BOLT_LMM_INF],
		line[CHISQL_BOLT_LMM],
		line[P_BOLT_LMM])

	return err
}

func validateHeader(header []string) error {
	for key, observed := range header {
		if expected, exists := expectedHeader[key]; exists && expected != observed {
			return fmt.Errorf("For column %d, expected %s but saw %s", key, expected, observed)
		}
	}

	return nil
}

func OpenDelimited(saige string) ([]string, *csv.Reader, func() error, error) {

	// Space first
	r, closer, err := openWithDelim(saige, ' ')
	if err != nil {
		return nil, nil, nil, err
	}

	line, err := r.Read()
	if err != nil {
		return nil, nil, nil, err
	}
	if len(line) >= 2 {
		// Space is the right delim
		return line, r, closer, nil
	}

	// Space isn't the right delim; try tab
	closer()

	r, closer, err = openWithDelim(saige, '\t')
	if err != nil {
		return nil, nil, nil, err
	}

	line, err = r.Read()
	if err != nil {
		return nil, nil, nil, err
	}

	if len(line) >= 2 {
		// Tab is the right delim
		return line, r, closer, nil
	}

	// Tab isn't right either
	closer()

	return nil, nil, nil, fmt.Errorf("%s seems to be delimited with neither spaces nor tabs", saige)
}

func openWithDelim(saige string, delim rune) (*csv.Reader, func() error, error) {
	f, err := os.Open(saige)
	if err != nil {
		return nil, nil, err
	}

	r := csv.NewReader(f)
	r.ReuseRecord = true
	r.TrimLeadingSpace = true
	// Try delim
	r.Comma = delim

	return r, f.Close, nil
}
