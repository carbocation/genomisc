package bulkprocess

import (
	"image"
	"os"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
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
