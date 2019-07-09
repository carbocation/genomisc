package main

import (
	"archive/zip"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/dicomtag"
	"github.com/suyashkumar/dicom/element"
)

// ExtractDicomFromLocalFile constructs a native go Image type from the dicom image with the
// given name in the given zip file.
func ExtractDicomFromLocalFile(zipPath, dicomName string, includeOverlay bool) (image.Image, error) {
	f, err := os.Open(zipPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// the zip reader wants to know the # of bytes in advance
	nBytes, err := f.Stat()
	if err != nil {
		return nil, err
	}

	img, err := ExtractDicomFromReaderAt(f, nBytes.Size(), dicomName, includeOverlay)
	if err != nil {
		return nil, fmt.Errorf("Err parsing zip %s: %s", zipPath, err.Error())
	}

	return img, nil
}

// ExtractDicomFromReaderAt operates directly on a ReaderAt object, which makes
// it possible to use in conjunction with, e.g., a Google Storage object.
func ExtractDicomFromReaderAt(readerAt io.ReaderAt, zipNBytes int64, dicomName string, includeOverlay bool) (image.Image, error) {
	var err error

	rc, err := zip.NewReader(readerAt, zipNBytes)
	if err != nil {
		return nil, err
	}

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
			return nil, fmt.Errorf("Error reading zip: %v", err)
		}

		var bitsAllocated, bitsStored, highBit uint16
		_, _, _ = bitsAllocated, bitsStored, highBit

		var nOverlayRows, nOverlayCols int

		var img *image.Gray16

		var imgRows, imgCols int
		var imgPixels []int
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
			} else if elem.Tag.Compare(dicomtag.Tag{Group: 0x6000, Element: 0x0010}) == 0 {
				nOverlayRows = int(elem.Value[0].(uint16))
			} else if elem.Tag.Compare(dicomtag.Tag{Group: 0x6000, Element: 0x0011}) == 0 {
				nOverlayCols = int(elem.Value[0].(uint16))
			} else if elem.Tag == dicomtag.Rows {
				imgRows = int(elem.Value[0].(uint16))
			} else if elem.Tag == dicomtag.Columns {
				imgCols = int(elem.Value[0].(uint16))
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

					for j := 0; j < len(frame.NativeData.Data); j++ {
						imgPixels = append(imgPixels, frame.NativeData.Data[j][0])
					}
				}

			}

			// Extract the overlay, if it exists
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

		// Identify the brightest pixel
		maxIntensity := 0
		for _, v := range imgPixels {
			if v > maxIntensity {
				maxIntensity = v
			}
		}

		// Draw the image
		img = image.NewGray16(image.Rect(0, 0, imgCols, imgRows))
		for j := 0; j < len(imgPixels); j++ {
			leVal := imgPixels[j]

			// Should be %cols and /cols -- row count is not necessary here
			img.SetGray16(j%imgCols, j/imgCols, color.Gray16{Y: ApplyPythonicWindowScaling(leVal, maxIntensity)})
		}

		// Draw the overlay
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

	return nil, fmt.Errorf("Did not find the requested Dicom %s", dicomName)
}

func ApplyPythonicWindowScaling(intensity, maxIntensity int) uint16 {
	if intensity < 0 {
		intensity = 0
	}

	return uint16(float64(math.MaxUint16) * float64(intensity) / float64(maxIntensity))
}
