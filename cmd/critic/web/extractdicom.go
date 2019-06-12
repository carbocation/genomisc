package main

import (
	"archive/zip"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/dicomtag"
	"github.com/suyashkumar/dicom/element"
)

// ExtractDicom constructs a native go Image type from the dicom image with the
// given name in the given zip file.
func ExtractDicom(zipPath, dicomName string, includeOverlay bool) (image.Image, error) {
	var err error

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

		var bitsAllocated, bitsStored, highBit uint16
		_, _, _ = bitsAllocated, bitsStored, highBit

		var windowCenter, windowWidth uint16
		var nOverlayRows, nOverlayCols int

		var img *image.Gray16
		var overlayPixels []int

		for _, elem := range parsedData.Elements {

			// The typical approach is to extract bitsAllocated, bitsStored, and the highBit
			// and to do transformations on the raw pixel values

			if elem.Tag == dicomtag.BitsAllocated {
				// log.Printf("BitsAllocated: %+v %T\n", elem.Value, elem.Value[0])
				bitsAllocated = elem.Value[0].(uint16)
			} else if elem.Tag == dicomtag.BitsStored {
				// log.Printf("BitsStored: %+v %T\n", elem.Value, elem.Value[0])
				bitsStored = elem.Value[0].(uint16)
			} else if elem.Tag == dicomtag.HighBit {
				// log.Printf("HighBit: %+v %T\n", elem.Value, elem.Value[0])
				highBit = elem.Value[0].(uint16)
			} else if elem.Tag == dicomtag.WindowCenter {
				// log.Printf("WindowCenter: %+v %T\n", elem.Value, elem.Value[0])
				windowCenter64, err := strconv.ParseUint(elem.Value[0].(string), 10, 16)
				if err != nil {
					return nil, err
				}
				windowCenter = uint16(windowCenter64)

			} else if elem.Tag == dicomtag.WindowWidth {
				// log.Printf("WindowWidth: %+v %T\n", elem.Value, elem.Value[0])
				windowWidth64, err := strconv.ParseUint(elem.Value[0].(string), 10, 16)
				if err != nil {
					return nil, err
				}
				windowWidth = uint16(windowWidth64)
			} else if elem.Tag.Compare(dicomtag.Tag{Group: 0x6000, Element: 0x0010}) == 0 {
				nOverlayRows = int(elem.Value[0].(uint16))
			} else if elem.Tag.Compare(dicomtag.Tag{Group: 0x6000, Element: 0x0011}) == 0 {
				nOverlayCols = int(elem.Value[0].(uint16))
			}

			if false {
				// Keeping for debugging
				if elem.Tag == dicomtag.PixelRepresentation {
					log.Printf("PixelRepresentation: %+v %T\n", elem.Value, elem.Value[0])
				} else if elem.Tag == dicomtag.RescaleSlope {
					log.Printf("RescaleSlope: %+v %T\n", elem.Value, elem.Value[0])
				} else if elem.Tag == dicomtag.RescaleIntercept {
					log.Printf("RescaleIntercept: %+v %T\n", elem.Value, elem.Value[0])
				} else if elem.Tag == dicomtag.RescaleType {
					log.Printf("RescaleType: %+v %T\n", elem.Value, elem.Value[0])
				} else if elem.Tag == dicomtag.PixelIntensityRelationship {
					log.Printf("PixelIntensityRelationship: %+v %T\n", elem.Value, elem.Value[0])
				} else if elem.Tag == dicomtag.PhotometricInterpretation {
					log.Printf("PhotometricInterpretation: %+v %T\n", elem.Value, elem.Value[0])
				} else if elem.Tag == dicomtag.SamplesPerPixel {
					log.Printf("SamplesPerPixel: %+v %T\n", elem.Value, elem.Value[0])
				} else if elem.Tag == dicomtag.TransferSyntaxUID {
					log.Printf("TransferSyntaxUID: %+v %T\n", elem.Value, elem.Value[0])
				} else if elem.Tag == dicomtag.SmallestImagePixelValue {
					log.Printf("SmallestImagePixelValue: %+v %T\n", elem.Value, elem.Value[0])
				} else if elem.Tag == dicomtag.LargestImagePixelValue {
					log.Printf("LargestImagePixelValue: %+v %T\n", elem.Value, elem.Value[0])
				} else if elem.Tag == dicomtag.VOILUTFunction {
					log.Printf("VOILUTFunction: %+v %T\n", elem.Value, elem.Value[0])
				}
			}

			// Main image
			if elem.Tag == dicomtag.PixelData {

				data := elem.Value[0].(element.PixelDataInfo)

				for _, frame := range data.Frames {
					if frame.IsEncapsulated {
						return nil, fmt.Errorf("Frame is encapsulated, which we did not expect")
					}

					img = image.NewGray16(image.Rect(0, 0, frame.NativeData.Cols, frame.NativeData.Rows))
					for j := 0; j < len(frame.NativeData.Data); j++ {
						leVal := uint16(frame.NativeData.Data[j][0])

						// Should be %cols and /cols -- row count is not necessary here
						img.SetGray16(j%frame.NativeData.Cols, j/frame.NativeData.Cols, color.Gray16{Y: uint16(float64(1<<16) * ApplyWindowScaling(leVal, windowCenter, windowWidth))})
					}
				}

				log.Println("Image bounds:", img.Bounds().Dx(), img.Bounds().Dy())
			}

			// Overlay
			if elem.Tag.Compare(dicomtag.Tag{Group: 0x6000, Element: 0x3000}) == 0 {
				log.Println("Found the Overlay")

				log.Println("Overlay bounds:", nOverlayCols, nOverlayRows)

				// We're in the overlay data
				for _, enclosed := range elem.Value {
					// There should be one enclosure, and it should contain a slice of
					// bytes, one byte per pixel.

					cellVals, ok := enclosed.([]byte)
					if !ok {
						continue
					}

					n_bits := 8

					// Fill an array with zeroes, sized the nRows * nCols ( == n_bits *
					// len(cellVals) )
					overlayPixels = make([]int, n_bits*len(cellVals), n_bits*len(cellVals))

					log.Println("Created a", len(overlayPixels), "array to hold the output")

					for i := range cellVals {
						byte_as_int := cellVals[i]
						for j := 0; j < n_bits; j++ {
							// Should be %cols and /cols -- row count is not necessary here
							overlayPixels[i*n_bits+j] = int((byte_as_int >> uint(j)) & 1)
						}
					}
				}
			}
		}

		if includeOverlay && img != nil && overlayPixels != nil {
			// Iterate over the bytes. There will be 1 value for each cell.
			// So in a 1024x1024 overlay, you will expect 1,048,576 cells.
			for i, overlayValue := range overlayPixels {
				row := i / nOverlayCols
				col := i % nOverlayCols

				if overlayValue != 0 {
					img.SetGray16(col, row, color.White)
				}
			}
		}

		return img, nil

	}

	return nil, fmt.Errorf("Did not find the requested Dicom %s in the Zip %s", dicomName, zipPath)
}

// Algorithm from https://www.dabsoft.ch/dicom/3/C.11.2.1.2/
func ApplyWindowScaling(intensity, windowCenter, windowWidth uint16) float64 {
	x := float64(intensity)
	center := float64(windowCenter)
	width := float64(windowWidth)

	if x < center-0.5-(width-1)/2 {
		return 0.0
	}

	if x > center-0.5+(width-1)/2 {
		return 1.0
	}

	return ((x-(center-0.5))/(width-1) + 0.5)
}
