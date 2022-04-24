package main

import (
	"flag"
	"io"
	"log"
	"os"

	_ "github.com/carbocation/genomisc/compileinfoprint"
)

var (
	MapDelim     = " "
	MapCHRColumn = 0
	MapBPColumn  = 3
	MapCMColumn  = 2
)

const (
	PVARPosColumn = 1
)

func main() {
	// Consumes a .pvar file that has a "CM" column. Consumes a map file with
	// chr, basepair, and value columns. Interpolates the .pvar CM values based
	// on the map file and writes the pvar file.
	var pvarFile, mapFile, outFile string
	flag.StringVar(&pvarFile, "pvar", "", ".pvar file. Must have a 'CM' column.")
	flag.StringVar(&mapFile, "map", "", "map file. Expected to be headerless. By default, column 0 is chr, column 3 is basepair, and column 2 is the centiMorgan value")
	flag.IntVar(&MapCHRColumn, "chr", 0, "0-based column of the map file that contains the chromosome")
	flag.IntVar(&MapBPColumn, "bp", 3, "0-based column of the map file that contains the basepair")
	flag.IntVar(&MapCMColumn, "cm", 2, "0-based column of the map file that contains the centiMorgan value")
	flag.StringVar(&outFile, "out", "", "output file. If not specified, writes to stdout")
	flag.StringVar(&MapDelim, "delim", " ", "delimiter for the map file")
	flag.Parse()

	if pvarFile == "" || mapFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Writer
	var outWriter io.WriteCloser
	if outFile == "" {
		outWriter = os.Stdout
	} else {
		var err error
		outWriter, err = os.Create(outFile)
		if err != nil {
			log.Fatalln(err)
		}
	}
	defer outWriter.Close()

	// Map
	mapReader, err := os.Open(mapFile)
	if err != nil {
		log.Fatalln(err)
	}
	defer mapReader.Close()
	if err := loadMap(mapReader); err != nil {
		log.Fatalln(err)
	}

	// PVAR
	pvarReader, err := os.Open(pvarFile)
	if err != nil {
		log.Fatalln(err)
	}
	defer pvarReader.Close()

	// Process
	if err := processPVAR(pvarReader, outWriter); err != nil {
		log.Fatalln(err)
	}
}
