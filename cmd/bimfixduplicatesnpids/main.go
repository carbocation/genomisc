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

	r := csv.NewReader(x)
	r.Comma = ' '

	seen := make(map[string]struct{})
	fixed := 0

	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return pfx.Err(err)
		}

		if _, exists := seen[line[1]]; exists {
			fixed++
			line[1] += "_" + RandHeteroglyphs(3)
			seen[line[1]] = struct{}{}
		} else {
			seen[line[1]] = struct{}{}
		}

		fmt.Println(strings.Join(line, " "))
	}

	log.Println("Fixed", fixed, "duplicate SNP IDs")

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
