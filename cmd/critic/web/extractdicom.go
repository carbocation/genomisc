package main

import (
	"archive/zip"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"

	"github.com/gradienthealth/dicom"
	"github.com/gradienthealth/dicom/dicomtag"
)

// ExtractDicom constructs a native go Image type from the dicom image with the
// given name in the given zip file.
func ExtractDicom(zipPath, dicomName string) (image.Image, error) {
	rc, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	for _, v := range rc.File {
		// Iterate over all of the dicoms in the zip til we find the one with
		// the desired name

		if v.Name != dicomName {
			continue
		}
		dicomReader, err := v.Open()
		if err != nil {
			return nil, err
		}

		defer dicomReader.Close()

		dcm, err := ioutil.ReadAll(dicomReader)
		if err != nil {
			return nil, err
		}

		p, err := dicom.NewParserFromBytes(dcm, nil)
		if err != nil {
			return nil, err
		}

		parsedData, err := p.Parse(dicom.ParseOptions{
			DropPixelData: false,
		})

		if parsedData == nil || err != nil {
			return nil, fmt.Errorf("Error reading %s: %v", zipPath, err)
		}

		for _, elem := range parsedData.Elements {
			if elem.Tag == dicomtag.PixelData {
				data := elem.Value[0].(dicom.PixelDataInfo)

				for _, frame := range data.Frames {
					if frame.IsEncapsulated {
						return nil, fmt.Errorf("Frame is encapsulated, which we did not expect")
					}

					i := image.NewRGBA(image.Rect(0, 0, frame.NativeData.Cols, frame.NativeData.Rows))
					for j := 0; j < len(frame.NativeData.Data); j++ {
						intensity := byte(frame.NativeData.Data[j][0])
						i.SetRGBA(j%frame.NativeData.Cols, j/frame.NativeData.Rows, color.RGBA{R: intensity, B: 255 - intensity, G: intensity, A: 255}) // for now, assume we're not overflowing uint16, assume gray image
					}

					return i, nil
				}

			}
		}

	}

	return nil, fmt.Errorf("Did not find the requested Dicom %s in the Zip %s", dicomName, zipPath)
}
