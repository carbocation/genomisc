package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/carbocation/pfx"
)

const (
	CHR = iota
	SNP
	Morgans
	Position
	Allele1
	Allele2
)

// Since BIM files can be whitespace delimited but golang's CSV Reader only
// accepts a single delimiter character, use this to replace all tabs with
// spaces on-the-fly.
type tabToSpaceReader struct {
	r io.Reader
}

func (t tabToSpaceReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	for i := 0; i < n; i++ {
		if p[i] == '\t' {
			p[i] = ' '
		}
	}

	return
}

func main() {
	var file string
	flag.StringVar(&file, "file", "", "BIM file path")
	flag.Parse()

	if file == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(file); err != nil {
		log.Fatalln(err)
	}
}

func run(file string) error {
	x, err := os.Open(file)
	if err != nil {
		return pfx.Err(err)
	}
	defer x.Close()

	sr := tabToSpaceReader{x}

	r := csv.NewReader(sr)
	r.Comma = ' '

	seen := make(map[string]struct{})
	saw := 0
	fixed := 0
	random := 0

	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return pfx.Err(err)
		}

		if len(line) < 6 {
			continue
		}

		saw++

		if _, exists := seen[line[SNP]]; exists {
			fixed++

			// Add positional information
			line[SNP] += "_" + line[CHR] + "_" + line[Position] + "_" + line[Allele1] + "_" + line[Allele2]

			// Iteratively add random suffix until the entry is unique
			for {
				// Once we have achieved uniqueness, we are done
				if _, exists := seen[line[SNP]]; !exists {
					break
				}

				// Add random suffix
				if _, exists := seen[line[SNP]]; exists {
					line[SNP] += "_" + RandHeteroglyphs(3)
					random++
				}
			}

			// Mark the modified SNP as seen only after the loop
			seen[line[SNP]] = struct{}{}

		} else {
			seen[line[SNP]] = struct{}{}
		}

		fmt.Println(strings.Join(line, " "))
	}

	log.Println("Saw", saw, "total SNP IDs")
	log.Println("Fixed", fixed-random, "duplicate SNP IDs by adding allele information")
	log.Println("Had to resort to adding random suffixes to", random, "duplicate SNP IDs")

	return nil
}

// RandHeteroglyphs produces a string of n symbols which do
// not look like one another. (Derived to be the opposite of
// homoglyphs, which are symbols which look similar to one
// another and cannot be quickly distinguished.)
func RandHeteroglyphs(n int) string {
	var letters = []rune("abcdefghkmnpqrstwxyz")
	lenLetters := len(letters)
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(lenLetters)]
	}
	return string(b)
}
