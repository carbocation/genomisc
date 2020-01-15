package bulkprocess

import (
	"image"
	"os"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"

	"cloud.google.com/go/storage"
)

// ExtractImageFromLocalFile pulls an image with the specified suffix (derived
// from the DICOM name) from a local folder
func ExtractImageFromLocalFile(dicomName, suffix, folderPath string) (image.Image, error) {
	f, err := os.Open(folderPath + "/" + dicomName + suffix)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Extract and decode the image. Must be PNG, GIF, BMP, or JPEG formatted
	// (based on the decoders we have imported)
	img, _, err := image.Decode(f)

	return img, err
}

// ExtractImageFromGoogleStorage pulls an image with the specified suffix (derived
// from the DICOM name) from a possibly remote folder
func ExtractImageFromGoogleStorage(dicomName, suffix, folderPath string, storageClient *storage.Client) (image.Image, error) {

	// Read the PNG into memory still compressed - either from a local
	// file, or from Google storage, depending on the prefix you provide.
	f, _, err := MaybeOpenFromGoogleStorage(folderPath+"/"+dicomName+suffix, storageClient)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Extract and decode the image. Must be PNG, GIF, BMP, or JPEG formatted
	// (based on the decoders we have imported)
	img, _, err := image.Decode(f)

	return img, err
}
