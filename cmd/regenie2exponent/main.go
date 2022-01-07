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

func main() {
	var regenie, logPCol, outDelim, traitName string
	var delimChar rune = ' '
	flag.StringVar(&regenie, "regenie", "", "Path to result file that you want to negate and exponentiate.")
	flag.StringVar(&logPCol, "log_p_col", "LOG10P", "Name of the column that you want to negate and exponentiate.")
	flag.StringVar(&outDelim, "output_delim", "\t", "Delimiter for the output.")
	flag.StringVar(&traitName, "trait_name", "", "Trait name. If not empty, will add a column with header trait_name.")
	flag.Func("delim", "Column delimiter.", func(v string) error {
		if len(v) < 1 {
			return nil
		}

		delimChar = rune(v[0])

		return nil
	})
	flag.Parse()

	if regenie == "" {
		log.Println("regenie2exponent")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var geneMap map[string]Gene
	geneList, err := FetchGenes()
	if err != nil {
		geneMap = make(map[string]Gene)
	} else {
		geneMap = GeneMapFromGeneSlice(geneList, GeneStableID)
	}

	r, closer, err := openWithDelim(regenie, delimChar)
	if err != nil {
		log.Fatalln(err)
	}
	defer closer()

	logPColNum := -1
	IDColNum := -1
	for i := 0; ; i++ {
		line, err := r.Read()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln(err)
		}

		if i == 0 {
			line = append(line, "Symbol")
			line = append(line, "P")
			if traitName != "" {
				line = append(line, "trait_name")
			}
			fmt.Println(strings.Join(line, outDelim))

			for j, v := range line {
				if v == logPCol {
					logPColNum = j
				}
				if v == "ID" {
					IDColNum = j
				}
			}

			if logPColNum < 0 {
				log.Fatalf("%s not found in %v", logPCol, line)
			}

			continue
		}

		symbol := ""
		if IDColNum >= 0 {
			symbol = geneMap[strings.Split(line[IDColNum], ".")[0]].Symbol
		}
		line = append(line, symbol)

		negLogP, err := negLogPToScientificNotationP(line[logPColNum])
		if err != nil {
			log.Fatalln(err)
		}
		line = append(line, negLogP)
		if traitName != "" {
			line = append(line, traitName)
		}
		fmt.Println(strings.Join(line, outDelim))
	}
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
	if mantissa < 1.0 {
		mantissa *= 10.0
		exponent -= 1.0
	}

	return fmt.Sprintf("%.1fE%.0f", mantissa, exponent), nil
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
	r.Comment = '#'

	return r, f.Close, nil
}
