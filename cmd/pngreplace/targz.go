package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"path"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
)

func processOneTarGZFilepath(filePath1, filePath2, filename string, threshold int, newLabels ReplacementMap) error {
	// Reader: Open and stream/ungzip the tar.gz
	f, _, err := bulkprocess.MaybeOpenFromGoogleStorage(filePath1, client)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	defer f.Close()
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	defer gzr.Close()
	tarReader := tar.NewReader(gzr)

	// Writer: Create tar.gz writer
	outFile, err := os.Create(filePath2 + "/" + filename)
	if err != nil {
		return err
	}
	defer outFile.Close()
	bw := bufio.NewWriter(outFile)
	defer bw.Flush()
	gw := gzip.NewWriter(bw)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Iterate over tarfile contents, processing all non-directory files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("%s: %w", filename, err)
		} else if header.Typeflag != tar.TypeReg {
			continue
		}

		// Decode image bytes from the tar reader
		newImg, err := bulkprocess.DecodeImageFromReader(tarReader)
		if err != nil {
			log.Println(fmt.Errorf("%s->%s: %w", filename, header.Name, err))
			continue
		}

		// Process image
		img, err := processOneImage(newImg, filePath2, header.Name, threshold, newLabels)
		if err != nil {
			log.Printf("%s: %v\n", header.Name, err)
			continue
		}

		// Encode image bytes into the tar.gz writer
		if err := addPNGToArchive(tw, header.Name, img); err != nil {
			return fmt.Errorf("%s: %w", filename, err)
		}
	}

	return nil
}

func addPNGToArchive(tw *tar.Writer, filename string, img image.Image) error {
	// writing to .tar files is done by first calling WriteHeader, then actually
	// writing the file. One of the components of the tar header is the file
	// size. As a consequence, you need to know the number of bytes in the file
	// before it is written. Hence, we first buffered the image bytes rather
	// than writing directly to the tar writer.
	buf := new(bytes.Buffer)
	bw := bufio.NewWriter(buf)

	// Write the PNG representation of our ID-encoded image to disk
	if err := png.Encode(bw, img); err != nil {
		return err
	}
	bw.Flush()

	hdr := &tar.Header{
		Name: path.Base(filename),
		Mode: int64(0644),
		Size: int64(buf.Len()),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}

	if _, err := tw.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}

	return nil
}
