package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"log"
	"os"
	"strings"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
)

func loadImagesFromTarGz(filePath1 string) ([]image.Image, error) {
	images := make([]image.Image, 0)

	// Reader: Open and stream/ungzip the tar.gz
	f, _, err := bulkprocess.MaybeOpenFromGoogleStorage(filePath1, client)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()
	tarReader := tar.NewReader(gzr)

	// Iterate over tarfile contents, processing all non-directory files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		} else if header.Typeflag != tar.TypeReg {
			continue
		}

		// Decode image bytes from the tar reader
		newImg, err := bulkprocess.DecodeImageFromReader(tarReader)
		if err != nil {
			log.Println(fmt.Errorf("%s->%s: %w", filePath1, header.Name, err))
			continue
		}

		images = append(images, newImg)
	}

	return images, nil
}

func loadImagesFromFolder(folder string) ([]image.Image, error) {
	files, err := os.Open(folder)
	if err != nil {
		return nil, err
	}
	defer files.Close()

	names, err := files.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	images := make([]image.Image, 0, len(names))
	for _, name := range names {
		file, err := os.Open(folder + "/" + name)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		if !strings.HasSuffix(name, ".png") {
			continue
		}

		img, err := png.Decode(file)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
		file.Close()
	}
	return images, nil
}

func loadImagesFromGIF(gifFilePath string) ([]image.Image, error) {
	// Reader: Open and stream/ungzip the tar.gz
	f, _, err := bulkprocess.MaybeOpenFromGoogleStorage(gifFilePath, client)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Decode the GIF
	gif, err := gif.DecodeAll(f)
	if err != nil {
		return nil, err
	}

	// Convert to PNG
	images := make([]image.Image, 0, len(gif.Image))
	for _, frame := range gif.Image {
		img := image.NewRGBA(image.Rect(0, 0, frame.Bounds().Dx(), frame.Bounds().Dy()))
		draw.Draw(img, img.Bounds(), frame, image.Point{}, draw.Src)
		images = append(images, img)
	}

	return images, nil
}
