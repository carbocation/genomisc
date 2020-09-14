package main

import (
	"image/gif"
	"os"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
)

func makeOneGif(pngs []string, outName string, delay int) error {
	outGif, err := bulkprocess.MakeOneGIFFromPaths(pngs, delay, client)
	if err != nil {
		return err
	}

	// Save file
	f, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	return gif.EncodeAll(f, outGif)
}
