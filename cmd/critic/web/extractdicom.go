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
	"strconv"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/dicomtag"
	"github.com/suyashkumar/dicom/element"
)

// https://stackoverflow.com/a/16641340/199475

// ExtractDicom constructs a native go Image type from the dicom image with the
// given name in the given zip file.
func ExtractDicom(zipPath, dicomName string) (image.Image, error) {
	var err error

	rc, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	// We will read our 2-byte (16-bit) little endian values into this buffer
	pxBuf := make([]byte, 2)

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

		var bitsAllocated, bitsStored, highBit, windowCenter, windowWidth, windowCenterBE, windowWidthBE uint16
		_, _, _, _, _, _, _ = bitsAllocated, bitsStored, highBit, windowCenter, windowWidth, windowCenterBE, windowWidthBE

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
			} else if elem.Tag == dicomtag.WindowCenter {
				log.Printf("WindowCenter: %+v %T\n", elem.Value, elem.Value[0])
				windowCenter64, err := strconv.ParseUint(elem.Value[0].(string), 10, 16)
				if err != nil {
					return nil, err
				}
				windowCenter = uint16(windowCenter64)

				// Get the big-endian representation
				// binary.LittleEndian.PutUint16(pxBuf, windowCenter)
				// windowCenterBE = binary.BigEndian.Uint16(pxBuf)

			} else if elem.Tag == dicomtag.WindowWidth {
				log.Printf("WindowWidth: %+v %T\n", elem.Value, elem.Value[0])
				windowWidth64, err := strconv.ParseUint(elem.Value[0].(string), 10, 16)
				if err != nil {
					return nil, err
				}
				windowWidth = uint16(windowWidth64)

				// Get the big-endian representation
				// binary.LittleEndian.PutUint16(pxBuf, windowWidth)
				// windowWidthBE = binary.BigEndian.Uint16(pxBuf)
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
			} else if elem.Tag == dicomtag.VOILUTFunction {
				log.Printf("VOILUTFunction: %+v %T\n", elem.Value, elem.Value[0])
			}

			if elem.Tag == dicomtag.PixelData {
				// https://stackoverflow.com/a/8765366/199475
				var lowestVisibleValue, highestVisibleValue uint16
				if windowWidth > 0 {
					lowestVisibleValue = uint16(float64(windowCenter) - float64(windowWidth)/float64(2))
					highestVisibleValue = uint16(float64(windowCenter) + float64(windowWidth)/float64(2))

					log.Println("LowestVisibleLE", lowestVisibleValue, "HighestVisibleLE", highestVisibleValue)

					binary.LittleEndian.PutUint16(pxBuf, lowestVisibleValue)
					lowestVisibleValue = binary.BigEndian.Uint16(pxBuf)

					binary.LittleEndian.PutUint16(pxBuf, highestVisibleValue)
					highestVisibleValue = binary.BigEndian.Uint16(pxBuf)

					log.Println("LowestVisibleLETx", lowestVisibleValue, "HighestVisibleLETx", highestVisibleValue)
				}

				data := elem.Value[0].(element.PixelDataInfo)

				for _, frame := range data.Frames {
					if frame.IsEncapsulated {
						return nil, fmt.Errorf("Frame is encapsulated, which we did not expect")
					}

					// For normalization
					var lowest uint16 = math.MaxUint16
					var highest uint16

					var lowestLE uint16 = math.MaxUint16
					var highestLE uint16
					for j := 0; j < len(frame.NativeData.Data); j++ {
						leVal := uint16(frame.NativeData.Data[j][0])

						if leVal < lowestLE {
							lowestLE = leVal
						}
						if leVal > highestLE {
							highestLE = leVal
						}

						// if leVal < lowestVisibleValue {
						// 	leVal = lowestVisibleValue
						// } else if leVal > highestVisibleValue {
						// 	leVal = highestVisibleValue
						// }

						binary.LittleEndian.PutUint16(pxBuf, leVal)
						px := binary.BigEndian.Uint16(pxBuf)

						if px < lowest {
							lowest = px
							// lowestLE = leVal
						}

						if px > highest {
							highest = px
							// highestLE = leVal
						}
					}

					log.Println("Lowest", lowest, "Highest", highest)
					log.Println("Lowest little endian", lowestLE, "Highest little endian", highestLE)
					log.Println("Window Center", windowCenter, "Window Width", windowWidth)
					log.Println("WindowCenterBE", windowCenterBE, "windowWidthBE", windowWidthBE)

					i := image.NewGray16(image.Rect(0, 0, frame.NativeData.Cols, frame.NativeData.Rows))
					// pxBuf = make([]byte, 2) // 2 bytes == 16 bits
					for j := 0; j < len(frame.NativeData.Data); j++ {
						leVal := uint16(frame.NativeData.Data[j][0])

						// if leVal < lowestVisibleValue {
						// 	leVal = lowestVisibleValue
						// } else if leVal > highestVisibleValue {
						// 	leVal = highestVisibleValue
						// }

						// binary.LittleEndian.PutUint16(pxBuf, leVal)
						// px := binary.BigEndian.Uint16(pxBuf)
						// _ = px

						// if px > highestVisibleValue {
						// 	px = 1<<16 - 1
						// }

						// if leVal > 220 && leVal < 270 {
						// 	px = 1<<16 - 1
						// }

						// ((byte[0] & 0x0f) << 8) | byte[1];
						// px = (uint16(pxBuf[1]&0x0f) << 8) | uint16(pxBuf[0])

						i.SetGray16(j%frame.NativeData.Cols, j/frame.NativeData.Rows, color.Gray16{Y: uint16(float64(1<<16) * ApplyWindowScaling(leVal, windowCenter, windowWidth))})

						// i.SetGray16(j%frame.NativeData.Cols, j/frame.NativeData.Rows, color.Gray16{Y: uint16(px << (bitsAllocated - bitsStored))})
						// i.SetGray16(j%frame.NativeData.Cols, j/frame.NativeData.Rows, color.Gray16{Y: uint16(px)})
						// i.SetGray16(j%frame.NativeData.Cols, j/frame.NativeData.Rows, color.Gray16{Y: uint16(float64(1<<16) * math.Min(1.0, (float64(leVal)-float64(lowestLE))/(float64(highestLE)-float64(lowestLE))))})
					}

					return i, nil
				}

			}
		}

	}

	return nil, fmt.Errorf("Did not find the requested Dicom %s in the Zip %s", dicomName, zipPath)
}

// if (x <= c - 0.5 - (w-1)/2), then y = y min
// else if (x > c - 0.5 + (w-1)/2), then y = y max ,
// else y = ((x - (c - 0.5)) / (w-1) + 0.5) * (y max - y min )+ y min
func ApplyWindowScaling(intensity, windowCenter, windowWidth uint16) float64 {
	// https://www.dabsoft.ch/dicom/3/C.11.2.1.2/
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
