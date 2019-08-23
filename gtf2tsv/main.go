// gtf2tsv flattens a gtf file to a tsv
// seems to require a 2-pass approach
package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const (
	// Delim is the character used to delimit the output
	Delim = '\t'
)

var (
	STDOUT = bufio.NewWriterSize(os.Stdout, 4096)
)

func main() {
	defer STDOUT.Flush()

	var filename string
	flag.StringVar(&filename, "file", "", "Path to the gtf file to convert to tsv.")
	flag.Parse()

	if filename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Println("Pass 1: discovering column names")
	cols, colMap, err := BuildDict(filename)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Pass 2: printing")
	if err := PrintFile(filename, Delim, cols, colMap); err != nil {
		log.Fatalln(err)
	}
}

func PrintFile(filename string, delim rune, cols []string, colMap map[string]int) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)

	w := csv.NewWriter(STDOUT)
	w.Comma = delim
	defer w.Flush()

	// Print the header
	w.Write(cols)

	var line string
	for i := 0; ; i++ {
		line, err = r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("GTF 0-based row %d error %s: %s", i, err, line)
		}

		lineCandidate := strings.TrimSuffix(line, "\n")
		if strings.HasPrefix(lineCandidate, "#") {
			continue
		}

		row := strings.Split(lineCandidate, "\t")

		// Populate the mandatory fields
		lineOut := make([]string, len(cols), len(cols))
		for i, col := range row {
			if i >= 8 || col == "." {
				continue
			}
			lineOut[i] = col
		}

		// Populate whatever attributes we received
		attributes, err := ParseAttributes(row[8])
		if err != nil {
			return err
		}

		for _, attr := range attributes {
			lineOut[colMap[attr.Key]] = attr.Value
		}

		w.Write(lineOut)

	}

	return nil
}

// BuildDict performs the first pass to identify exhaustively all possible
// columns and what order they were seen in
func BuildDict(filename string) ([]string, map[string]int, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	r := bufio.NewReader(f)

	cols := []string{"seqname", "source", "feature", "start", "end", "score", "strand", "frame"}
	colMap := map[string]int{}
	for k, v := range cols {
		colMap[v] = k
	}

	var line string
	for i := 0; ; i++ {
		// row, err := r.Read()
		line, err = r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, fmt.Errorf("GTF 0-based row %d error %s: %s", i, err, line)
		}

		lineCandidate := strings.TrimSuffix(line, "\n")
		if strings.HasPrefix(lineCandidate, "#") {
			continue
		}

		row := strings.Split(lineCandidate, "\t")

		if x := len(row); x < 9 {
			return nil, nil, fmt.Errorf("GTF 0-based row %d had %d lines, expected 9", i, x)
		}

		attributes, err := ParseAttributes(row[8])
		if err != nil {
			return nil, nil, fmt.Errorf("Line %d: %s (%+v)", i, err, row[8])
		}

		for _, attr := range attributes {
			if _, exists := colMap[attr.Key]; !exists {
				log.Println("Found new column:", attr.Key)
				cols = append(cols, attr.Key)
				colMap[attr.Key] = len(cols) - 1
			}
		}
	}

	return cols, colMap, nil
}

type KeyValue struct {
	Key   string
	Value string
}

func ParseAttributes(attr string) ([]KeyValue, error) {
	out := make([]KeyValue, 0, 0)

	attributes := strings.Split(attr, ";")
	for i, attribute := range attributes {
		parts := strings.SplitN(strings.TrimSpace(attribute), " ", 2)
		if x := len(parts); x < 2 {
			// Line ends in a semicolon
			break
		} else if x != 2 {
			return nil, fmt.Errorf("Expected 2 parts; attribute %d had %d (%+v)", i, x, parts)
		}

		out = append(out, KeyValue{Key: parts[0], Value: strings.Trim(parts[1], "\"")})
	}

	return out, nil
}
