package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"

	_ "github.com/carbocation/genomisc/compileinfoprint"
)

// See https://github.com/rgcgithub/regenie/issues/50#issuecomment-723664958 for
// confirmation of the meaning of BETA and A1FREQ in REGENIE with regard to SNP.

// # REGENIE-v2 (space-delim)
// # CHROM GENPOS ID ALLELE0 ALLELE1 A1FREQ INFO N TEST BETA SE CHISQ LOG10P
// # BOLT (tab-delim)
// # SNP	CHR	BP	GENPOS	ALLELE1	ALLELE0	A1FREQ	INFO	CHISQ_LINREG	P_LINREG	BETA	SE	CHISQ_BOLT_LMM_INF	P_BOLT_LMM_INF	CHISQ_BOLT_LMM	P_BOLT_LMM

// # REGENIE-v2.2.4 (space-delim)
// # CHROM GENPOS ID ALLELE0 ALLELE1 A1FREQ INFO N TEST BETA SE CHISQ LOG10P EXTRA
// # REGENIE-v2.2.4 (space-delim) binary trait with --af-cc flag
// # CHROM GENPOS ID ALLELE0 ALLELE1 A1FREQ A1FREQ_CASES A1FREQ_CONTROLS INFO N TEST BETA SE CHISQ LOG10P EXTRA

// # REGENIE-v2.2.4 (no imputation data)
// # CHROM GENPOS ID ALLELE0 ALLELE1 A1FREQ N TEST BETA SE CHISQ LOG10P EXTRA
// # REGENIE-v2.2.4 (no imputation data) binary trait with --af-cc flag
// # CHROM GENPOS ID ALLELE0 ALLELE1 A1FREQ A1FREQ_CASES A1FREQ_CONTROLS N TEST BETA SE CHISQ LOG10P EXTRA

// {print $3 OFS $1 OFS $2 OFS "NA" OFS $5 OFS $4 OFS $6 OFS $7 OFS $12 OFS 10 ^ (-1 * $13) OFS $10 OFS $11 OFS $12 OFS 10 ^ (-1 * $13) OFS $12 OFS 10 ^ (-1 * $13) }' -

const (
	HeaderTypeFixedOrderDefault = iota
	HeaderTypeFixedOrderCC
	HeaderTypeFixedOrderNoINFO
	HeaderTypeFixedOrderCCNoINFO
)

var headerType = HeaderTypeFixedOrderDefault

// REGENIE header that we expect

// 0-based regenie column ID -> column name
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
	// New:
	13: "EXTRA",
}

// 0-based regenie column ID -> column name
var expectedHeaderNoINFO = map[int]string{
	0:  "CHROM",
	1:  "GENPOS",
	2:  "ID",
	3:  "ALLELE0",
	4:  "ALLELE1",
	5:  "A1FREQ",
	8:  "BETA",
	9:  "SE",
	10: "CHISQ",
	11: "LOG10P",
	// New:
	12: "EXTRA",
}

// 0-based regenie column ID -> column name
var expectedHeaderCaseControl = map[int]string{
	0: "CHROM",
	1: "GENPOS",
	2: "ID",
	3: "ALLELE0",
	4: "ALLELE1",
	5: "A1FREQ",
	// New:
	6: "A1FREQ_CASES",
	7: "A1FREQ_CONTROLS",
	// Different locations:
	8:  "INFO",
	11: "BETA",
	12: "SE",
	13: "CHISQ",
	14: "LOG10P",
	// New
	15: "EXTRA",
}

// 0-based regenie column ID -> column name
var expectedHeaderCaseControlNoINFO = map[int]string{
	0: "CHROM",
	1: "GENPOS",
	2: "ID",
	3: "ALLELE0",
	4: "ALLELE1",
	5: "A1FREQ",
	// New:
	6: "A1FREQ_CASES",
	7: "A1FREQ_CONTROLS",
	// Different locations:
	10: "BETA",
	11: "SE",
	12: "CHISQ",
	13: "LOG10P",
	// New
	14: "EXTRA",
}

// Mapping between BOLT column name and its corresponding column number from
// REGENIE file
var (
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

func overrideColMapping(m map[int]string) {
	for regenieColumnKey, regenieColumnName := range m {
		switch regenieColumnName {
		case "CHROM":
			CHR = regenieColumnKey
		case "GENPOS":
			POS = regenieColumnKey
		case "ID":
			SNP = regenieColumnKey
		case "ALLELE0":
			ALLELE0 = regenieColumnKey
		case "ALLELE1":
			ALLELE1 = regenieColumnKey
		case "A1FREQ":
			A1FREQ = regenieColumnKey
		case "INFO":
			INFO = regenieColumnKey
		case "BETA":
			BETA = regenieColumnKey
		case "SE":
			SE = regenieColumnKey
		case "CHISQ":
			CHISQ_BOLT_LMM = regenieColumnKey
			CHISQ_LINREG = regenieColumnKey
			CHISQ_BOLT_LMM_INF = regenieColumnKey
		case "LOG10P":
			P_BOLT_LMM = regenieColumnKey
			P_LINREG = regenieColumnKey
			P_BOLT_LMM_INF = regenieColumnKey
		}
	}
}

// var kvMap = map[string]int{
// 	"SNP":                SNP,
// 	"CHR":                CHR,
// 	"POS":                POS,
// 	"ALLELE1":            ALLELE1,
// 	"ALLELE0":            ALLELE0,
// 	"A1FREQ":             A1FREQ,
// 	"INFO":               INFO,
// 	"CHISQ_LINREG":       CHISQ_LINREG,
// 	"P_LINREG":           P_LINREG,
// 	"BETA":               BETA,
// 	"SE":                 SE,
// 	"CHISQ_BOLT_LMM_INF": CHISQ_BOLT_LMM_INF,
// 	"P_BOLT_LMM_INF":     P_BOLT_LMM_INF,
// 	"CHISQ_BOLT_LMM":     CHISQ_BOLT_LMM,
// 	"P_BOLT_LMM":         P_BOLT_LMM,
// }

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
	fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		"SNP",
		"CHR",
		"BP",
		"GENPOS",
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
	exponent := math.Ceil((-1 * negLogP) / 1.0)

	// Make it pretty (should get mantissa into the 1-10 range)

	// If you don't round during this comparison check, then you end up with
	// things like "10.0E-3" when the -logP is 2.001.

	// Via https://stackoverflow.com/a/49175144/199475 . This is overkill here,
	// but we certainly won't overflow this way...
	f := new(big.Float).SetMode(big.ToNearestEven).SetFloat64(mantissa)
	f = f.SetPrec(1)
	mantissaRounded, _ := f.Float64()
	if mantissaRounded < 1.0 {
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

	var infoField string
	if headerType == HeaderTypeFixedOrderCCNoINFO || headerType == HeaderTypeFixedOrderNoINFO {
		// TODO: make this an option in the command line. 1.0 makes sense for
		// whole genome sequencing, but may not make sense in all cases where
		// INFO is not provided.
		infoField = "1.0"
	} else {
		infoField = line[INFO]
	}

	_, err = fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		line[SNP],
		line[CHR],
		line[POS],
		"NA",
		line[ALLELE1],
		line[ALLELE0],
		line[A1FREQ],
		infoField,
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
	var err error

	for headerTypeOrder, expectedHeaderGroup := range []map[int]string{
		// Note: order is important here
		expectedHeader,
		expectedHeaderCaseControl,
		expectedHeaderNoINFO,
		expectedHeaderCaseControlNoINFO,
	} {
		log.Println("testing header type", headerTypeOrder)
		err = nil

		for key, observed := range header {
			if expected, exists := expectedHeaderGroup[key]; exists && expected != observed {
				err = fmt.Errorf("for column %d, expected %s but saw %s", key, expected, observed)

				// Break the inner loop as soon as we see an error, since this
				// cannot be the right format
				break
			}
		}

		if err == nil {
			headerType = headerTypeOrder
			overrideColMapping(expectedHeaderGroup)
			break
		}
	}

	// If we make it to here and err is still not nil, then we return the error,
	// since none of the candidate header types matched.

	return err
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
