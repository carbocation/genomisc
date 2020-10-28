package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"io/ioutil"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/dicomtag"
	"github.com/suyashkumar/dicom/element"
)

// DicomToImage constructs native go Image types from the dicom images. Usually
// (perhaps always?) there will be just one image.
func DicomToImages(dicomReader io.Reader) ([]image.Image, error) {
	dcm, err := ioutil.ReadAll(dicomReader)
	if err != nil {
		return nil, err
	}

	p, err := dicom.NewParserFromBytes(dcm, nil)
	if err != nil {
		return nil, err
	}

	parsedData, err := bulkprocess.SafelyDicomParse(p, dicom.ParseOptions{
		DropPixelData: false,
	})

	if parsedData == nil || err != nil {
		return nil, fmt.Errorf("Error reading dicom")
	}

	var output []image.Image

	for _, elem := range parsedData.Elements {
		if elem.Tag == dicomtag.PixelData {
			data := elem.Value[0].(element.PixelDataInfo)

			for _, frame := range data.Frames {
				if frame.IsEncapsulated() {
					return nil, fmt.Errorf("Frame is encapsulated, which we did not expect")
				}

				//i := image.NewRGBA(image.Rect(0, 0, frame.NativeData.Cols, frame.NativeData.Rows))
				i := image.NewGray(image.Rect(0, 0, frame.NativeData.Cols, frame.NativeData.Rows))

				for j := 0; j < len(frame.NativeData.Data); j++ {
					// To the extent that I can tell, we are receiving 8-bit grayscale information.
					intensity := byte(frame.NativeData.Data[j][0])

					//i.SetRGBA(j%frame.NativeData.Cols, j/frame.NativeData.Rows, color.RGBA{R: intensity, B: intensity, G: intensity, A: 255})
					i.SetGray(j%frame.NativeData.Cols, j/frame.NativeData.Rows, color.Gray{intensity})
				}

				output = append(output, i)
			}

		}
	}

	return output, nil
}

// Takes in a dicom file (in bytes), outputs one or more jpeg file equivalents
// (in bytes)
func DicomToJpeg(dicomReader io.Reader) ([][]byte, error) {
	dcm, err := ioutil.ReadAll(dicomReader)
	if err != nil {
		return nil, err
	}

	p, err := dicom.NewParserFromBytes(dcm, nil)
	if err != nil {
		return nil, err
	}

	parsedData, err := bulkprocess.SafelyDicomParse(p, dicom.ParseOptions{
		DropPixelData: false,
	})
	if parsedData == nil || err != nil {
		return nil, fmt.Errorf("Error reading dicom: %v", err)
	}

	var output [][]byte

	for _, elem := range parsedData.Elements {
		if elem.Tag != dicomtag.PixelData {
			continue
		}

		data := elem.Value[0].(element.PixelDataInfo)

		for _, frame := range data.Frames {

			// Encapsulated

			if frame.IsEncapsulated() {
				output = append(output, frame.EncapsulatedData.Data)
				continue
			}

			// Unencapsulated

			img := image.NewGray16(image.Rect(0, 0, frame.NativeData.Cols, frame.NativeData.Rows))
			for j := 0; j < len(frame.NativeData.Data); j++ {
				// for now, assume we're not overflowing uint16, assume gray image
				img.SetGray16(j%frame.NativeData.Cols, j/frame.NativeData.Rows, color.Gray16{Y: uint16(frame.NativeData.Data[j][0])})
			}
			buf := new(bytes.Buffer)
			jpeg.Encode(buf, img, &jpeg.Options{Quality: 100})
			output = append(output, buf.Bytes())
		}
	}

	return output, nil
}
