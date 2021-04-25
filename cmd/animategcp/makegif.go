package main

import (
	"image"
	"image/gif"
	"os"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/pfx"
)

func makeOneGrid(dicomNames []string, outName string, delay int, withTransparency bool) error {

	// Fetch the images based on the dicom names and shove them into a map
	sortedPngs, err := bulkprocess.FetchGIFComponents(dicomNames, client)
	if err != nil {
		return pfx.Err(err)
	}
	imgMap := make(map[string]image.Image)
	for k, dicomName := range dicomNames {
		imgMap[dicomName] = sortedPngs[k]
	}

	// Construct our image grid with 4 columns. Assumes 50 images per cardiac
	// cycle. Extremely hacky.
	newDicomNames, newImageMap, err := bulkprocess.ImageGrid(dicomNames, imgMap, "", nil, 50, 4)
	if err != nil {
		return pfx.Err(err)
	}

	// Create the GIF
	outGif, err := bulkprocess.MakeOneGIFFromMap(newDicomNames, newImageMap, delay, withTransparency)
	if err != nil {
		return pfx.Err(err)
	}

	// Save file
	f, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return pfx.Err(err)
	}

	defer f.Close()

	return pfx.Err(gif.EncodeAll(f, outGif))
}

func makeOneGif(pngs []string, outName string, delay int, withTransparency bool) error {
	outGif, err := bulkprocess.MakeOneGIFFromPaths(pngs, delay, withTransparency, client)
	if err != nil {
		return pfx.Err(err)
	}

	// Save file
	f, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return pfx.Err(err)
	}

	defer f.Close()

	return pfx.Err(gif.EncodeAll(f, outGif))
}
