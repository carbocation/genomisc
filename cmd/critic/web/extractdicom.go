package main

import (
	"archive/zip"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"math"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/dicomtag"
	"github.com/suyashkumar/dicom/element"
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
			// ReturnTags:    []dicomtag.Tag{
			// 	// Need to understand the range of data in the pixels, a la
			// 	// https://stackoverflow.com/a/35055589/199475
			// 	// dicomtag.BitsAllocated,
			// 	// dicomtag.BitsStored,
			// 	// dicomtag.HighBit,
			// 	// dicomtag.NumberOfFrames,
			// },
		})

		if parsedData == nil || err != nil {
			return nil, fmt.Errorf("Error reading %s: %v", zipPath, err)
		}

		var bitsAllocated, bitsStored, highBit uint16
		_, _, _ = bitsAllocated, bitsStored, highBit

		for _, elem := range parsedData.Elements {

			// The typical approach is to extract bitsAllocated, bitsStored, and the highBit
			// and to do transformations on the raw pixel values

			if elem.Tag == dicomtag.BitsAllocated {
				log.Printf("BitsAllocated: %+v %T\n", elem.Value, elem.Value[0])
				bitsAllocated = elem.Value[0].(uint16)
			} else if elem.Tag == dicomtag.BitsStored {
				log.Printf("BitsStored: %+v %T\n", elem.Value, elem.Value[0])
				bitsStored = elem.Value[0].(uint16)
			} else if elem.Tag == dicomtag.HighBit {
				log.Printf("HighBit: %+v %T\n", elem.Value, elem.Value[0])
				highBit = elem.Value[0].(uint16)
			}

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
			} else if elem.Tag == dicomtag.WindowCenter {
				log.Printf("WindowCenter: %+v %T\n", elem.Value, elem.Value[0])
			} else if elem.Tag == dicomtag.WindowWidth {
				log.Printf("WindowWidth: %+v %T\n", elem.Value, elem.Value[0])
			} else if elem.Tag == dicomtag.VOILUTFunction {
				log.Printf("VOILUTFunction: %+v %T\n", elem.Value, elem.Value[0])
			}

			if elem.Tag == dicomtag.PixelData {
				data := elem.Value[0].(element.PixelDataInfo)

				for _, frame := range data.Frames {
					if frame.IsEncapsulated {
						return nil, fmt.Errorf("Frame is encapsulated, which we did not expect")
					}

					// We will read our 2-byte (16-bit) little endian values into this buffer
					pxBuf := make([]byte, 2)

					// For normalization
					var lowest uint16 = math.MaxUint16
					var highest uint16

					var lowestLE uint16 = math.MaxUint16
					var highestLE uint16
					for j := 0; j < len(frame.NativeData.Data); j++ {
						leVal := uint16(frame.NativeData.Data[j][0])
						binary.LittleEndian.PutUint16(pxBuf, leVal)
						px := binary.BigEndian.Uint16(pxBuf)

						if px < lowest {
							// lowest = px
							// lowestLE = leVal
						}

						if px > highest {
							// highest = px
							// highestLE = leVal
						}

						if leVal < lowestLE {
							lowest = px
							lowestLE = leVal
						}
						if leVal > highestLE {
							// highestLE = leVal
							highest = px
							highestLE = leVal
						}
					}

					log.Println("Lowest", lowest, "Highest", highest)
					log.Println("Lowest little endian", lowestLE, "Highest little endian", highestLE)

					i := image.NewGray16(image.Rect(0, 0, frame.NativeData.Cols, frame.NativeData.Rows))
					// pxBuf = make([]byte, 2) // 2 bytes == 16 bits
					for j := 0; j < len(frame.NativeData.Data); j++ {

						// The UK Biobank images I've seen so far are 16-bit
						binary.LittleEndian.PutUint16(pxBuf, uint16(frame.NativeData.Data[j][0]))

						px := binary.BigEndian.Uint16(pxBuf)

						// i.SetGray16(j%frame.NativeData.Cols, j/frame.NativeData.Rows, color.Gray16{Y: uint16(px << (bitsAllocated - bitsStored))})
						// i.SetGray16(j%frame.NativeData.Cols, j/frame.NativeData.Rows, color.Gray16{Y: uint16(px)})
						i.SetGray16(j%frame.NativeData.Cols, j/frame.NativeData.Rows, color.Gray16{Y: uint16(float64(1<<16) * math.Min(1.0, (float64(px)-float64(lowest))/(float64(highest)-float64(lowest))))})
					}

					return i, nil
				}

			}
		}

	}

	return nil, fmt.Errorf("Did not find the requested Dicom %s in the Zip %s", dicomName, zipPath)
}
