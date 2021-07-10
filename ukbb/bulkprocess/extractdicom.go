package bulkprocess

import (
	"archive/zip"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"math"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/dicomtag"
	"github.com/suyashkumar/dicom/element"
)

// ExtractDicomFromGoogleStorage fetches a dicom from within a zipped file in
// Google Storage and returns it as a native go image.Image, optionally with the
// overlay displayed on top.
func ExtractDicomFromGoogleStorage(zipPath, dicomName string, includeOverlay bool, storageClient *storage.Client) (image.Image, error) {
	// Read the zip file handle into memory still compressed and turn it into an
	// io.ReaderAt which is appropriate for consumption by the zip reader -
	// either from a local file, or from Google storage, depending on the prefix
	// you provide.
	f, nbytes, err := MaybeOpenFromGoogleStorage(zipPath, storageClient)
	if err != nil {
		return nil, err
	}

	rc, err := zip.NewReader(f, nbytes)
	if err != nil {
		return nil, err
	}

	// Now we have our compressed zip data in an zip.Reader, regardless of its
	// origin.
	return ExtractDicomFromZipReader(rc, dicomName, includeOverlay)
}

// ExtractDicomFromLocalFile constructs a native go Image type from the dicom
// image with the given name in the given zip file. Now just wraps the
// GoogleStorage variant, since it has the capability of loading local files as
// well as remote ones.
func ExtractDicomFromLocalFile(zipPath, dicomName string, includeOverlay bool) (image.Image, error) {
	return ExtractDicomFromGoogleStorage(zipPath, dicomName, includeOverlay, nil)
}

// ExtractDicomFromZipReader consumes a zip reader of the UK Biobank format,
// finds the dicom of the desired name, and returns that image, with or without
// the overlay (if any is present) based on includeOverlay.
func ExtractDicomFromZipReader(rc *zip.Reader, dicomName string, includeOverlay bool) (image.Image, error) {

	for _, v := range rc.File {
		// Iterate over all of the dicoms in the zip til we find the one with
		// the desired name. This is reasonably efficient since we don't need to
		// read all of the data to find the right name.
		if v.Name != dicomName {
			continue
		}

		dicomReader, err := v.Open()
		if err != nil {
			return nil, err
		}
		defer dicomReader.Close()

		img, err := ExtractDicomFromReader(dicomReader, int64(v.UncompressedSize64), includeOverlay)

		return img, err
	}

	return nil, fmt.Errorf("Did not find the requested Dicom %s", dicomName)
}

// ExtractDicomFromReader operates on a reader that contains one DICOM.
func ExtractDicomFromReader(dicomReader io.Reader, nReaderBytes int64, includeOverlay bool) (image.Image, error) {

	p, err := dicom.NewParser(dicomReader, nReaderBytes, nil)
	if err != nil {
		return nil, err
	}

	parsedData, err := SafelyDicomParse(p, dicom.ParseOptions{
		DropPixelData: false,
	})

	if parsedData == nil || err != nil {
		return nil, fmt.Errorf("Error reading zip: %v", err)
	}

	var rescaleSlope, rescaleIntercept, windowWidth, windowCenter float64
	_, _, _, _ = rescaleSlope, rescaleIntercept, windowWidth, windowCenter

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

		if elem.Tag == dicomtag.RescaleSlope {
			rescaleSlope, err = strconv.ParseFloat(elem.Value[0].(string), 64)
			if err != nil {
				log.Println(err)
			}
		}
		if elem.Tag == dicomtag.RescaleIntercept {
			rescaleIntercept, err = strconv.ParseFloat(elem.Value[0].(string), 64)
			if err != nil {
				log.Println(err)
			}
		}
		if elem.Tag == dicomtag.WindowWidth {
			windowWidth, err = strconv.ParseFloat(elem.Value[0].(string), 64)
			if err != nil {
				log.Println(err)
			}
		}
		if elem.Tag == dicomtag.WindowCenter {
			windowCenter, err = strconv.ParseFloat(elem.Value[0].(string), 64)
			if err != nil {
				log.Println(err)
			}
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
				if frame.IsEncapsulated() {
					encImg, err := frame.GetImage()
					if err != nil {
						return nil, fmt.Errorf("Frame is encapsulated, which we did not expect. Additionally, %s", err.Error())
					}

					// We're done, since it's not clear how to add an overlay
					return encImg, nil

				}

				for j := 0; j < len(frame.NativeData.Data); j++ {
					imgPixels = append(imgPixels, frame.NativeData.Data[j][0])
				}
			}

		}

		// Extract the overlay, if it exists and we want it
		if includeOverlay && elem.Tag.Compare(dicomtag.Tag{Group: 0x6000, Element: 0x3000}) == 0 {
			// log.Println("Found the Overlay")

			// log.Println("Overlay bounds:", nOverlayCols, nOverlayRows)
			_, _ = nOverlayCols, nOverlayRows

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

				// log.Println("Created a", len(overlayPixels), "array to hold the output")

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
		if false { //j > 3000 {
			img.SetGray16(j%imgCols, j/imgCols, color.Gray16{Y: ApplyOfficialWindowScaling(leVal, rescaleSlope, rescaleIntercept, windowWidth, windowCenter, bitsAllocated)})
		} else {
			img.SetGray16(j%imgCols, j/imgCols, color.Gray16{Y: ApplyPythonicWindowScaling(leVal, maxIntensity)})
		}
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

	return img, err
}

// See 'Grayscale Image Display' under
// https://dgobbi.github.io/vtk-dicom/doc/api/image_display.html
func ApplyOfficialWindowScaling(storedValue int, rescaleSlope, rescaleIntercept, windowWidth, windowCenter float64, bitsAllocated uint16) uint16 {

	// 1: StoredValue to ModalityValue

	modalityValue := float64(storedValue)*rescaleSlope + rescaleIntercept

	// 2: ModalityValue to WindowedValue

	// The key here is that we're using bitsAllocated (e.g., 16 bits) instead of
	// bitsStored (e.g., 11 bits)
	grayLevels := math.Pow(2, float64(bitsAllocated))

	w := windowWidth - 1.0
	c := windowCenter - 0.5

	if modalityValue <= c-0.5*w {
		return 0
	}

	if modalityValue > c+0.5*w {
		return uint16(grayLevels - 1.0)
	}

	return uint16(((modalityValue-c)/w + 0.5) * (grayLevels - 1.0))

}

func ApplyPythonicWindowScaling(intensity, maxIntensity int) uint16 {
	if intensity < 0 {
		intensity = 0
	}

	return uint16(float64(math.MaxUint16) * float64(intensity) / float64(maxIntensity))
}
