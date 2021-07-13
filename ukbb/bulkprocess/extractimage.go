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

// ExtractImageFromTarGzMaybeFromGoogleStorage is now a thin wrapper over
// ExtractImagesFromTarGzMaybeFromGoogleStorage.
func ExtractImageFromTarGzMaybeFromGoogleStorage(tarGz, imageFilename string, client *storage.Client) (image.Image, error) {
	imgMap := make(map[string]struct{})
	imgMap[imageFilename] = struct{}{}

	imgs, err := ExtractImagesFromTarGzMaybeFromGoogleStorage(tarGz, imgMap, client)
	if err != nil {
		return nil, err
	}

	return imgs[imageFilename], nil
}

// ExtractImagesFromTarGzMaybeFromGoogleStorage looks for tarGz (a fully
// qualified path to a .tar.gz file, which may optionally be a Google Storage
// path). If that succeeds, it then looks for imageFilenames within the
// ungzipped untarred file and processes them into image.Image files, which it
// returns in a map with the same names as the input map.
func ExtractImagesFromTarGzMaybeFromGoogleStorage(tarGz string, imageFilenames map[string]struct{}, client *storage.Client) (map[string]image.Image, error) {
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

	return ExtractImagesFromTarReader(tarReader, imageFilenames)
}

// ExtractImageFromTarReader consumes a .tar reader, loops over all files until
// it finds the image with the specified name, and returns that image. This is
// now a thin wrapper over ExtractImagesFromTarReader.
func ExtractImageFromTarReader(tarReader *tar.Reader, imageFilename string) (image.Image, error) {
	imgMap := make(map[string]struct{})
	imgMap[imageFilename] = struct{}{}

	imgs, err := ExtractImagesFromTarReader(tarReader, imgMap)
	if err != nil {
		return nil, err
	}

	return imgs[imageFilename], nil
}

// ExtractImagesFromTarReader consumes a .tar reader, loops over all files until
// it finds all images with the specified names, and returns those images in a
// named map. If it fails to find all requested files, it still returns the
// files it found, but also returns an error.
func ExtractImagesFromTarReader(tarReader *tar.Reader, imageFilenames map[string]struct{}) (map[string]image.Image, error) {

	lenImageFilenames := len(imageFilenames)
	out := make(map[string]image.Image)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if _, isWanted := imageFilenames[header.Name]; isWanted && header.Typeflag == tar.TypeReg {
			imageBytes, err := ioutil.ReadAll(tarReader)
			if err != nil {
				return nil, err
			}

			br := bytes.NewReader(imageBytes)
			newImg, err := DecodeImageFromReader(br)
			if err != nil {
				return nil, err
			}
			out[header.Name] = newImg

			if len(out) == lenImageFilenames {
				break
			}
		}
	}

	if len(out) != lenImageFilenames {
		return out, fmt.Errorf("ExtractImagesFromTarGz: found %d images but expected %d", len(out), lenImageFilenames)
	}

	return out, nil
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
