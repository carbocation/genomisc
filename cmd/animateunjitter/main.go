// animateunjitter removes lines from a sorted manifest in a fashion that avoids
// returning to a previously seen series_number. This empirically reduces
// jitter.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	_ "github.com/carbocation/genomisc/compileinfoprint"
)

func main() {
	var series, position int

	// Currently, col 16 is the image_y and col 17 is image_z. col 6 is the true
	// series name. 21 is the series_number. 28 is echo_time
	flag.IntVar(&series, "series_col", 21, "0-based column number of the series name to block jitter")
	flag.IntVar(&position, "position_col", 17, "0-based column number of the ascending position value to sort the gif")
	flag.Parse()

	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	// Following tips from https://flaviocopes.com/go-shell-pipes/
	if info.Mode()&os.ModeCharDevice != 0 {
		fmt.Fprintf(os.Stderr, "The command is intended to work with pipes.\n")
		fmt.Fprintf(os.Stderr, "Usage: cat manifest.tsv | animatemanifest\n")
		return
	}

	data, err := tRead(os.Stdin)
	if err != nil {
		log.Fatalln(err)
	}

	sortables, err := linesToSortables(data, series, position)
	if err != nil {
		log.Fatalln(err)
	}

	sort.Slice(sortables, func(i, j int) bool {
		return sortables[i].position > sortables[j].position
	})

	seenSeries := map[string]struct{}{}
	last := ""
	for _, v := range sortables {
		l := fmt.Sprintf(strings.Join(data[v.lineNumber], "\t"))

		if last == "" {
			// First entry
			fmt.Println(l)

			seenSeries[v.series] = struct{}{}
			last = v.series

			continue
		}

		if last != v.series {
			if _, exists := seenSeries[v.series]; !exists {
				// On to a new series
				fmt.Println(l)

				seenSeries[v.series] = struct{}{}
				last = v.series

				continue
			}

			// Trying to return to an old series. Don't allow it (jitter)
			continue
		}

		// Still within the same series as the last line:
		fmt.Println(l)
	}

}

func tRead(file *os.File) ([][]string, error) {
	input := [][]string{}

	c := csv.NewReader(file)
	c.Comma = '\t'
	for {
		line, err := c.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		input = append(input, line)
	}

	return input, nil
}
