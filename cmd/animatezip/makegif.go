package main

import (
	"image"
	"image/gif"
	"os"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
)

func makeOneGifFromImageMap(dicomNames []string, imgMap map[string]image.Image, outName string, delay int) error {

	outGif, err := bulkprocess.MakeOneGIFFromMap(dicomNames, imgMap, delay)
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
