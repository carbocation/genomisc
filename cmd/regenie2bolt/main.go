package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

// See https://github.com/rgcgithub/regenie/issues/50#issuecomment-723664958 for
// confirmation of the meaning of BETA and A1FREQ in REGENIE with regard to SNP.

// # REGENIE-v2 (space-delim)
// # CHROM GENPOS ID ALLELE0 ALLELE1 A1FREQ INFO N TEST BETA SE CHISQ LOG10P
// # BOLT (tab-delim)
// # SNP	CHR	BP	GENPOS	ALLELE1	ALLELE0	A1FREQ	INFO	CHISQ_LINREG	P_LINREG	BETA	SE	CHISQ_BOLT_LMM_INF	P_BOLT_LMM_INF	CHISQ_BOLT_LMM	P_BOLT_LMM

// {print $3 OFS $1 OFS $2 OFS "NA" OFS $5 OFS $4 OFS $6 OFS $7 OFS $12 OFS 10 ^ (-1 * $13) OFS $10 OFS $11 OFS $12 OFS 10 ^ (-1 * $13) OFS $12 OFS 10 ^ (-1 * $13) }' -

// REGENIE header that we expect
var expectedHeader = map[int]string{
	0:  "CHROM",
	1:  "GENPOS",
	2:  "ID",
	3:  "ALLELE0",
	4:  "ALLELE1",
	5:  "A1FREQ",
	6:  "INFO",
	9:  "BETA",
	10: "SE",
	11: "CHISQ",
	12: "LOG10P",
}

// Mapping between BOLT column name and its corresponding column number from
// REGENIE file
const (
	SNP                = 2
	CHR                = 0
	POS                = 1
	ALLELE1            = 4
	ALLELE0            = 3
	A1FREQ             = 5
	INFO               = 6
	CHISQ_LINREG       = 11
	P_LINREG           = 12
	BETA               = 9
	SE                 = 10
	CHISQ_BOLT_LMM_INF = 11
	P_BOLT_LMM_INF     = 12
	CHISQ_BOLT_LMM     = 11
	P_BOLT_LMM         = 12
)

func main() {
	var regenie string
	flag.StringVar(&regenie, "regenie", "", "Path to REGENIE result file that you want to convert to BOLT format.")
	flag.Parse()

	if regenie == "" {
		log.Println("regenie2bolt")
		fmt.Fprintln(os.Stderr,
			`Consumes a REGENIE output summary stats file and reorients it so that it behaves like a BOLT-LMM summary stats file.
  1. Zeroes preceding the CHR will be removed.
  2. BOLT ALLELE1 [== ref == effect] and REGENIE ALLELE1 [== ref == effect]. In both cases, ALLELE0 is alt. At least, when computing from UKBB BGEN files.
  So we just need to reformat the file to mimic that of BOLT and compute the P value.
  `)
		flag.PrintDefaults()
		os.Exit(1)
	}

	header, r, closer, err := OpenDelimited(regenie)
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
		"CHISQ_BOLT_LMM",
		"P_BOLT_LMM")
}

func negLogPToScientificNotationP(negLogPString string) (string, error) {
	// REGENIE provides -log10(P)
	negLogP, err := strconv.ParseFloat(negLogPString, 64)
	if err != nil {
		return "", fmt.Errorf("PrintLine: %w", err)
	}

	mantissa := math.Pow(10.0, math.Mod(-1*negLogP, 1.0))
	exponent := math.Floor((-1 * negLogP) / 1.0)

	// Make it pretty (should get mantissa into the 1-10 range)
	if mantissa < 1.0 {
		mantissa *= 10.0
		exponent -= 1.0
	}

	return fmt.Sprintf("%.1fE%.0f", mantissa, exponent), nil
}

func PrintLine(line []string) error {

	// Remove preceding zeroes from CHR
	line[CHR] = strings.TrimPrefix(line[CHR], "0")

	// Convert to the [Mantissa]E[Exponent] formatted string:
	Pstring, err := negLogPToScientificNotationP(line[P_BOLT_LMM])
	if err != nil {
		return fmt.Errorf("PrintLine: %w", err)
	}
	line[P_BOLT_LMM] = Pstring

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
		line[CHISQ_BOLT_LMM],
		line[P_BOLT_LMM])

	return err
}

func validateHeader(header []string) error {
	for key, observed := range header {
		if expected, exists := expectedHeader[key]; exists && expected != observed {
			return fmt.Errorf("for column %d, expected %s but saw %s", key, expected, observed)
		}
	}

	return nil
}

func OpenDelimited(inputFile string) ([]string, *csv.Reader, func() error, error) {

	// Space first
	r, closer, err := openWithDelim(inputFile, ' ')
	if err != nil {
		return nil, nil, nil, err
	}

	line, err := r.Read()
	if err != nil {
		return nil, nil, nil, err
	}
	if len(line) >= len(expectedHeader) {
		// Space is likely the right delim
		return line, r, closer, nil
	}

	// Space isn't the right delim; try tab
	closer()

	r, closer, err = openWithDelim(inputFile, '\t')
	if err != nil {
		return nil, nil, nil, err
	}

	line, err = r.Read()
	if err != nil {
		return nil, nil, nil, err
	}

	if len(line) >= len(expectedHeader) {
		// Tab is likely the right delim
		return line, r, closer, nil
	}

	// Tab isn't right either
	closer()

	return nil, nil, nil, fmt.Errorf("%s seems to be delimited with neither spaces nor tabs, since we don't find the number of columns that we expect (%d). Expected header: %v", inputFile, len(expectedHeader), expectedHeader)
}

func openWithDelim(inputFile string, delim rune) (*csv.Reader, func() error, error) {
	f, err := os.Open(inputFile)
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
