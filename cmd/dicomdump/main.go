package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strings"
)

var (
	BufferSize = 4096
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
)

// Makes one big combined manifest
// Emits to stdout
func main() {
	defer STDOUT.Flush()

	var path string

	flag.StringVar(&path, "path", "", "Path to a single raw .dcm file, or to a folder with UKBB bulk .zip files.")
	flag.Parse()

	if path == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Single DICOM file
	if strings.HasSuffix(path, ".dcm") {
		f, err := os.Open(path)
		if err != nil {
			log.Fatalln(err)
		}

		if err := ProcessDicom(f); err != nil {
			log.Fatalln(err)
		}

		return
	}

	// Folder of UKBB Zip files
	if err := IterateOverFolder(path); err != nil {
		log.Fatalln(err)
	}
}
