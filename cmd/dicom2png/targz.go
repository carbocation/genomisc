package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"image"
	"image/png"
	"os"
	"path"
	"path/filepath"
)

// NewTarGzWriter provides a closer which will sequentially close the tar
// writer, the gzip writer, and finally the underlying file writer in correct
// order.
func NewTarGzWriter(filePath, fileName string) (tw *tar.Writer, Close func() error, err error) {
	outFile, err := os.Create(filepath.Join(filePath, fileName))
	if err != nil {
		return nil, func() error { return nil }, err
	}
	gw := gzip.NewWriter(outFile)
	tw = tar.NewWriter(gw)

	closer := func() error {
		var err error

		if err = tw.Flush(); err != nil {
			return err
		}

		if err = tw.Close(); err != nil {
			return err
		}

		if err = gw.Close(); err != nil {
			return err
		}

		if err = outFile.Close(); err != nil {
			return err
		}

		return nil
	}

	return tw, closer, nil
}

func addPNGToArchive(tw *tar.Writer, img image.Image, baseFilename string) error {
	// writing to .tar files is done by first calling WriteHeader, then actually
	// writing the file. One of the components of the tar header is the file
	// size. As a consequence, you need to know the number of bytes in the file
	// before it is written. Hence, we first buffered the image bytes rather
	// than writing directly to the tar writer.
	buf := new(bytes.Buffer)
	bw := bufio.NewWriter(buf)

	// Write the PNG representation of our ID-encoded image
	if err := png.Encode(bw, img); err != nil {
		return err
	}
	bw.Flush()

	filename := baseFilename + ".png"

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
