package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	glo "github.com/carbocation/GLO"
	"github.com/carbocation/genomisc"
)

func initInputFile(inputFile string, chromColID, posColID int, hasHeader bool) (*csv.Reader, io.Closer, error) {
	f, err := genomisc.MaybeOpenSeekerFromGoogleStorage(inputFile, client)
	if err != nil {
		return nil, nil, err
	}

	r, err := genomisc.MaybeDecompressReadCloserFromFile(f)
	if err != nil {
		return nil, nil, err
	}

	delim := genomisc.DetermineDelimiter(r)
	f.Seek(0, 0)

	// Sort of nuts but we have to re-decompress the file since the decompressed
	// reader cannot seek.
	r, err = genomisc.MaybeDecompressReadCloserFromFile(f)
	if err != nil {
		return nil, nil, err
	}

	rdr := csv.NewReader(r)
	rdr.Comma = delim

	return rdr, r, nil
}

func initLiftoverFromChainFile(chainFile string) (liftover *glo.LiftOver, fromRef, toRef string, err error) {
	chunks := strings.Split(strings.Split(filepath.Base(chainFile), ".")[0], "To")
	if len(chunks) != 2 {
		return nil, "", "", fmt.Errorf("Expected chain file name format to be oldToNew.over.chain.*, but found: %s", chainFile)
	}

	fromRef = strings.ToLower(chunks[0])
	toRef = strings.ToLower(chunks[1])

	log.Println("Lifting from", fromRef, "to", toRef)

	f, err := genomisc.MaybeOpenSeekerFromGoogleStorage(chainFile, client)
	if err != nil {
		return nil, "", "", err
	}
	defer f.Close()

	// Detect the chainfile compression type and open a reader to decompress
	// before wrapping it in the buffer
	r, err := genomisc.MaybeDecompressReadCloserFromFile(f)
	if err != nil {
		return nil, "", "", err
	}
	buf := bufio.NewReader(r)

	liftover = new(glo.LiftOver)
	liftover.Init()
	liftover.Load(fromRef, toRef, buf)

	return liftover, fromRef, toRef, nil
}
