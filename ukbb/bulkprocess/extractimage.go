package bulkprocess

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"image"
	"io"
	"io/ioutil"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"

	"cloud.google.com/go/storage"
)

// ExtractImageFromTarGzMaybeFromGoogleStorage looks for tarGz (a fully
// qualified path to a .tar.gz file, which may optionally be a Google Storage
// path). If that succeeds, it then looks for imageFilename within the ungzipped
// untarred file and processes that into an image.Image.
func ExtractImageFromTarGzMaybeFromGoogleStorage(tarGz, imageFilename string, client *storage.Client) (image.Image, error) {
	imReader, _, err := MaybeOpenFromGoogleStorage(tarGz, client)
	if err != nil {
		return nil, err
	}
	defer imReader.Close()

	gzr, err := gzip.NewReader(imReader)
	if err != nil {
		return nil, err
	}

	tarReader := tar.NewReader(gzr)

	return ExtractImageFromTarReader(tarReader, imageFilename)
}

// ExtractImageFromTarReader consumes a .tar reader, loops over all files until
// it finds the image with the specified name, and returns that image. TODO: If
// extracting more than one file this way, it would be good to loop once to
// create a map, and index into the right position in the tar file, if that is
// possible.
func ExtractImageFromTarReader(tarReader *tar.Reader, imageFilename string) (image.Image, error) {

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if header.Name == imageFilename && header.Typeflag == tar.TypeReg {
			imageBytes, err := ioutil.ReadAll(tarReader)
			if err != nil {
				return nil, err
			}

			br := bytes.NewReader(imageBytes)
			return DecodeImageFromReader(br)
		}
	}

	return nil, fmt.Errorf("ExtractImageFromTarGz: file %s not found", imageFilename)
}

// ExtractImageFromLocalFile pulls an image with the specified suffix (derived
// from the DICOM name) from a local folder. Now just a wrapper.
func ExtractImageFromLocalFile(dicomName, suffix, folderPath string) (image.Image, error) {
	return ExtractImageFromGoogleStorage(dicomName, suffix, folderPath, nil)
}

// ExtractImageFromGoogleStorage is now just a deprecated wrapper around
// MaybeExtractImageFromGoogleStorage and may be removed in the future.
func ExtractImageFromGoogleStorage(dicomName, suffix, folderPath string, storageClient *storage.Client) (image.Image, error) {
	return MaybeExtractImageFromGoogleStorage(folderPath+"/"+dicomName+suffix, storageClient)
}

// MaybeExtractImageFromGoogleStorage pulls an image from a (possibly remote)
// folder
func MaybeExtractImageFromGoogleStorage(imagePath string, storageClient *storage.Client) (image.Image, error) {

	// Read the PNG into memory still compressed - either from a local file, or
	// from Google storage, depending on the prefix you provide.
	f, _, err := MaybeOpenFromGoogleStorage(imagePath, storageClient)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return DecodeImageFromReader(f)
}

// DecodeImageFromReader decodes a reader containing PNG, GIF, BMP, or
// JPEG-formatted data and returns a single-frame image.
func DecodeImageFromReader(f io.Reader) (image.Image, error) {
	// Extract and decode the image. Must be PNG, GIF, BMP, or JPEG formatted
	// (based on the decoders we have imported)
	img, _, err := image.Decode(f)

	return img, err
}
