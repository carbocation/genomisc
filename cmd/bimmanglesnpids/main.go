// bimmanglesnpids is designed to handle a specific situation where you want to
// mangle rsids intentionally. This arises when subsetting from a format (e.g.,
// bgen) which does not have expressive tools for requesting not only rsID but
// also ref and alt allele (since multiallelics are so common). So, this tool
// allows you to make sure that the SNPs in your BIM file contain not only a
// matching rsID, but also a matching ref and alt allele from a lookup file.
// And, if not,
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
	var file, otherFile string
	flag.StringVar(&file, "file", "", "BIM file path")
	flag.StringVar(&otherFile, "other_file", "", "Tab-delimited file with 3 columns and no header: rsid, ref, alt (order of ref and alt is irrelevant)")
	flag.Parse()

	if file == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(file, otherFile); err != nil {
		log.Fatalln(err)
	}
}

func run(file, other string) error {
	bimFile, err := os.Open(file)
	if err != nil {
		return pfx.Err(err)
	}
	defer bimFile.Close()

	otherFileMap, err := OtherFileMap(other)
	if err != nil {
		return err
	}

	sr := tabToSpaceReader{bimFile}

	r := csv.NewReader(sr)
	r.Comma = ' '

	saw := 0
	rejected := 0

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

		v1 := otherFileRow{RSID: line[SNP], Ref: line[Allele1], Alt: line[Allele2]}
		v2 := otherFileRow{RSID: line[SNP], Ref: line[Allele2], Alt: line[Allele1]}
		if _, exists := otherFileMap[v1]; exists {
			// Found the SNP, print it as-is
		} else if _, exists := otherFileMap[v2]; exists {
			// Found the SNP, print it as-is
		} else {
			// The SNP was not found. Mangle it.
			line[SNP] = "bad_" + RandHeteroglyphs(10)
			rejected++
		}

		fmt.Println(strings.Join(line, " "))
	}

	log.Println("Saw", saw, "total SNP IDs")
	log.Println("Mangled", rejected, "unwanted SNP IDs")
	log.Println("Yielded", saw-rejected, "permissible SNP IDs")

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

type otherFileRow struct {
	RSID string
	Ref  string
	Alt  string
}

// OtherFileMap converts your alternative input file into a map based on rsid,
// ref, and alt.
func OtherFileMap(other string) (map[otherFileRow]struct{}, error) {
	otherFile, err := os.Open(other)
	if err != nil {
		return nil, pfx.Err(err)
	}
	defer otherFile.Close()

	r := csv.NewReader(otherFile)
	r.Comma = '\t'
	rows, err := r.ReadAll()
	if err != nil {
		return nil, pfx.Err(err)
	}

	out := make(map[otherFileRow]struct{})

	for _, row := range rows {
		if len(row) < 3 {
			continue
		}

		out[otherFileRow{RSID: row[0], Ref: row[1], Alt: row[2]}] = struct{}{}
	}

	return out, nil
}
