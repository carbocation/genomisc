package main

import (
	"image"
	"image/gif"
	"os"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/pfx"
	"github.com/unixpickle/ffmpego"
)

func makeOneGIF(sortedImages []image.Image, outName string, delay int, withTransparency bool) error {
	outGIF, err := bulkprocess.MakeOneGIF(sortedImages, delay, withTransparency)
	if err != nil {
		return pfx.Err(err)
	}

	// Save file
	f, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return pfx.Err(err)
	}

	defer f.Close()

	return pfx.Err(gif.EncodeAll(f, outGIF))
}

func makeOneMPEG(sortedImages []image.Image, outName string, fps float64) error {
	w, h := 0, 0
	for _, v := range sortedImages {
		w = v.Bounds().Dx()
		h = v.Bounds().Dy()
		break
	}
	vw, err := ffmpego.NewVideoWriter(outName, w, h, fps)
	if err != nil {
		return err
	}
	defer vw.Close()

	for _, v := range sortedImages {
		if err := vw.WriteFrame(v); err != nil {
			return err
		}
	}

	return nil
}
