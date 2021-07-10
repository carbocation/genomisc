package bulkprocess

import (
	"image"
	"io"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"

	"cloud.google.com/go/storage"
)

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
