package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	glo "github.com/carbocation/GLO"
	_ "github.com/carbocation/genomisc/compileinfoprint"
)

var client *storage.Client

func main() {
	var chainFile, inputFile, outputFile string
	var chromColID, posColID int
	var hasHeader, addMissingChr, appendOld bool
	flag.StringVar(&chainFile, "chain", "", "Path to the chain file from UCSC. Optionally, may be a google storage URL (gs://)")
	flag.StringVar(&inputFile, "input", "", "Path to the input file whose reference is to be converted by the chain. Optionally, may be a google storage URL (gs://)")
	flag.StringVar(&outputFile, "output", "", "Path to the output file.")
	flag.IntVar(&chromColID, "chromcol", 0, "0-based column index of the chromosome in the input file")
	flag.IntVar(&posColID, "poscol", 1, "0-based column index of the position in the input file")
	flag.BoolVar(&hasHeader, "header", true, "Whether the input file has a header line")
	flag.BoolVar(&addMissingChr, "addmissingchr", true, "Whether to add a 'chr' prefix to the chromosome name if it is missing (e.g. '1' becomes 'chr1')")
	flag.BoolVar(&appendOld, "appendold", true, "Whether to append the original chromosome and position to the output")
	flag.Parse()

	if chainFile == "" {
		flag.Usage()
		log.Fatalln("Must specify a --chain file")
	}

	if inputFile == "" {
		flag.Usage()
		log.Fatalln("Must specify an --input file")
	}

	if strings.HasPrefix(chainFile, "gs://") ||
		strings.HasPrefix(inputFile, "gs://") {
		var err error
		client, err = storage.NewClient(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}

	liftover, fromRef, toRef, err := initLiftoverFromChainFile(chainFile)
	if err != nil {
		log.Fatalln(err)
	}

	inputReader, clsr, err := initInputFile(inputFile, chromColID, posColID, hasHeader)
	if err != nil {
		log.Fatalln(err)
	}
	defer clsr.Close()

	var out io.WriteCloser
	var outUnmapped io.WriteCloser
	if outputFile == "" {
		out = os.Stdout
		outUnmapped = os.Stderr
	} else {
		out, err = os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalln(err)
		}

		outUnmapped, err = os.OpenFile(outputFile+".unmapped", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalln(err)
		}
	}
	defer out.Close()
	defer outUnmapped.Close()

	bw := bufio.NewWriter(out)
	buw := bufio.NewWriter(outUnmapped)

	defer bw.Flush()
	defer buw.Flush()

	mappedCount := 0
	unMappedCount := 0
	var origHeader []string
	for {
		line, err := inputReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln(err)
		}

		if hasHeader {
			origHeader = line
			if appendOld {
				origChrom := "ORIG_" + line[chromColID]
				origPos := "ORIG_" + line[posColID]
				line = append(line, origChrom, origPos)
			}
			fmt.Fprintln(bw, strings.Join(line, string(inputReader.Comma)))
			hasHeader = false
			continue
		}
		mappedCount++

		chr := line[chromColID]
		if addMissingChr && !strings.HasPrefix(chr, "chr") {
			chr = "chr" + chr
		}
		posS := line[posColID]
		pos, err := strconv.ParseInt(posS, 10, 64)
		if err != nil {
			log.Fatalln(err)
		}

		o := liftover.Lift(fromRef, toRef, glo.NewChainInterval(chr, pos, pos+1))
		for _, x := range o {

			if appendOld {
				line = append(line, line[chromColID], line[posColID])
			}

			line[chromColID] = x.Contig
			line[posColID] = strconv.FormatInt(x.Start, 10)

			fmt.Fprintln(bw, strings.Join(line, string(inputReader.Comma)))
		}

		if len(o) == 0 {
			if unMappedCount == 0 {
				fmt.Fprintln(buw, strings.Join(origHeader, string(inputReader.Comma)))
			}
			unMappedCount++
			fmt.Fprintln(buw, strings.Join(line, string(inputReader.Comma)))
		}
	}

	log.Printf("Finished. Mapped: %d. Unmapped: %d\n", mappedCount, unMappedCount)
}
