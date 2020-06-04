package main

import (
	"bufio"
	"flag"
	"log"
	"os"
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

	flag.StringVar(&path, "path", "", "Path where the UKBB bulk .zip files are being held.")
	flag.Parse()

	if path == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Replace the output target with a file, if desired
	if err := IterateOverFolder(path); err != nil {
		log.Fatalln(err)
	}
}
