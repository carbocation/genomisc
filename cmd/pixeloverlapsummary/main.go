// pixeloverlapsummary is a convenience tool to summarize the output of
// pixeloverlap by label
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/montanaflynn/stats"
)

func main() {
	var input string
	var linePrefix string

	// Parse the command line arguments
	flag.StringVar(&input, "input", "", "The input file")
	flag.StringVar(&linePrefix, "line_prefix", "", "Column to add to each line. If empty, no column will be added.")
	flag.Parse()

	if input == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Open the input file
	f, err := os.Open(input)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	// Parse the input file
	if err := parsePixelOverlap(f, linePrefix); err != nil {
		log.Fatalln(err)
	}
}

type StatCollection struct {
	Agree, Only1, Only2, Kappa, Dice, Jaccard, CountAgreement []string
}

func newStatCollection(N int) StatCollection {
	return StatCollection{
		Agree:          make([]string, 0, N),
		Only1:          make([]string, 0, N),
		Only2:          make([]string, 0, N),
		Kappa:          make([]string, 0, N),
		Dice:           make([]string, 0, N),
		Jaccard:        make([]string, 0, N),
		CountAgreement: make([]string, 0, N),
	}
}

func parsePixelOverlap(f io.Reader, linePrefix string) error {
	// Treat the reader f as a csv.Reader
	csvReader := csv.NewReader(f)
	csvReader.Comma = '\t'
	entries, err := csvReader.ReadAll()
	if err != nil {
		return err
	}

	// If entries is not empty, then its first row is the header
	var header = make(map[string]int)
	if len(entries) > 0 {
		// header = entries[0]
		for i, col := range entries[0] {
			header[col] = i
		}
	} else {
		return fmt.Errorf("No entries in the input file")
	}

	labelValues := make(map[string]StatCollection)

	// // Iterate over all entries, except the header
	output := []string{
		"Label",
	}

	if linePrefix != "" {
		output = append(output, "LinePrefix")
	}

	output = append(output, []string{
		"Filter",
		"N_Entries",
		"Agree",
		"AgreeSD",
		"Only1",
		"Only1SD",
		"Only2",
		"Only2SD",
		"Kappa",
		"KappaSD",
		"Dice",
		"DiceSD",
		"Jaccard",
		"JaccardSD",
		"CountAgreement",
		"CountAgreementSD",
	}...)
	fmt.Println(strings.Join(output, "\t"))

	for i, row := range entries {
		if i == 0 {
			continue
		}

		labelValueSlice, ok := labelValues[row[header["Label"]]]
		if !ok {
			labelValueSlice = newStatCollection(len(entries))
			labelValues[row[header["Label"]]] = labelValueSlice
		}

		labelValueSlice.Agree = append(labelValueSlice.Agree, row[header["Agree"]])
		labelValueSlice.Only1 = append(labelValueSlice.Only1, row[header["Only1"]])
		labelValueSlice.Only2 = append(labelValueSlice.Only2, row[header["Only2"]])
		labelValueSlice.Kappa = append(labelValueSlice.Kappa, row[header["Kappa"]])
		labelValueSlice.Dice = append(labelValueSlice.Dice, row[header["Dice"]])
		labelValueSlice.Jaccard = append(labelValueSlice.Jaccard, row[header["Jaccard"]])
		labelValueSlice.CountAgreement = append(labelValueSlice.CountAgreement, row[header["CountAgreement"]])
		labelValues[row[header["Label"]]] = labelValueSlice
	}

	if err := printValues(labelValues, "raw", linePrefix); err != nil {
		return err
	}

	filteredValues, err := filterToNonEmptyValues(labelValues)
	if err != nil {
		return err
	}

	if err := printValues(filteredValues, "nonzero", linePrefix); err != nil {
		return err
	}

	return nil

}

func filterToNonEmptyValues(labelValues map[string]StatCollection) (map[string]StatCollection, error) {
	v2 := make(map[string]StatCollection)
	for k, v0 := range labelValues {

		entry := newStatCollection(len(v0.Agree))

		for i := range v0.Agree {
			agree, err := strconv.ParseFloat(v0.Agree[i], 64)
			if err != nil {
				return nil, err
			}

			only1, err := strconv.ParseFloat(v0.Only1[i], 64)
			if err != nil {
				return nil, err
			}

			only2, err := strconv.ParseFloat(v0.Only2[i], 64)
			if err != nil {
				return nil, err
			}

			if agree+only1+only2 == 0 {
				continue
			}

			entry.Agree = append(entry.Agree, v0.Agree[i])
			entry.Only1 = append(entry.Only1, v0.Only1[i])
			entry.Only2 = append(entry.Only2, v0.Only2[i])
			entry.Kappa = append(entry.Kappa, v0.Kappa[i])
			entry.Dice = append(entry.Dice, v0.Dice[i])
			entry.Jaccard = append(entry.Jaccard, v0.Jaccard[i])
			entry.CountAgreement = append(entry.CountAgreement, v0.CountAgreement[i])
		}

		v2[k] = entry
	}

	return v2, nil
}

func printValues(labelValues map[string]StatCollection, filterType, linePrefix string) error {
	for label, values := range labelValues {
		output := []string{label}

		if linePrefix != "" {
			output = append(output, linePrefix)
		}

		output = append(output, []string{filterType, fmt.Sprintf("%d", len(values.Agree))}...)

		entryV := reflect.ValueOf(values)
		entryK := reflect.VisibleFields(reflect.TypeOf(values))
		for i := range entryK {
			data := stats.LoadRawData(entryV.Field(i).Interface())

			if data.Len() < 1 {
				output = append(output, []string{"N/A", "N/A"}...)
				continue
			}

			fl, err := data.Mean()
			if err != nil {
				return err
			}
			output = append(output, fmt.Sprintf("%.3f", fl))

			fl, err = data.StandardDeviation()
			if err != nil {
				return err
			}
			output = append(output, fmt.Sprintf("%.3f", fl))
		}
		fmt.Println(strings.Join(output, "\t"))
	}

	return nil
}
