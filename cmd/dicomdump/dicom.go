package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strings"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/pfx"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/dicomtag"
	"github.com/suyashkumar/dicom/element"
)

func IterateOverFolder(path string) error {

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return pfx.Err(err)
	}

	for _, file := range files {

		func(file os.FileInfo) {

			if !strings.HasSuffix(file.Name(), ".zip") {
				return
			}

			if err := ProcessZip(path + file.Name()); err != nil {
				log.Println(err)
				return
			}
		}(file)
	}

	return nil
}

func ProcessZip(zipPath string) (err error) {

	fmt.Println(strings.Repeat("=", 30))
	fmt.Println(zipPath)
	fmt.Println(strings.Repeat("=", 30))

	rc, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	for _, v := range rc.File {
		// Looking only at the dicoms
		if strings.HasPrefix(v.Name, "manifest") {
			continue
		}

		fmt.Println(strings.Repeat("-", 30))
		fmt.Println(v.Name)
		fmt.Println(strings.Repeat("-", 30))

		unzippedFile, err := v.Open()
		if err != nil {
			return err
		}
		if err := ProcessDicom(unzippedFile); err != nil {
			log.Println("Ignoring error and continuing:", err.Error())
			continue
		}
	}

	return nil
}

// Takes in a dicom file (in bytes), emit meta-information
func ProcessDicom(dicomReader io.Reader) error {
	dcm, err := ioutil.ReadAll(dicomReader)
	if err != nil {
		return err
	}

	p, err := dicom.NewParserFromBytes(dcm, nil)
	if err != nil {
		return err
	}

	parsedData, err := bulkprocess.SafelyDicomParse(p, dicom.ParseOptions{
		DropPixelData: false,
	})
	if parsedData == nil || err != nil {
		return fmt.Errorf("Error reading dicom: %v", err)
	}

	var bitsStored float64

	for _, elem := range parsedData.Elements {
		tagName, _ := dicomtag.Find(elem.Tag)

		if tagName.Name == "" {
			tagName.Name = "____"
		}

		if elem.Tag == dicomtag.BitsStored {
			bitsStored = float64(elem.Value[0].(uint16))
		}

		// We don't want to flood the screen with pixel data. Note that as long
		// as we have set DropPixelData: true above in dicom.ParseOptions, these
		// checks are redundant.

		if elem.Tag == dicomtag.PixelData {
			data := elem.Value[0].(element.PixelDataInfo)

			imgPixels := make([]int, 0)

			for _, frame := range data.Frames {
				if frame.IsEncapsulated() {
					_, err := frame.GetImage()
					if err != nil {
						return fmt.Errorf("Frame is encapsulated, which we did not expect. Additionally, %s", err.Error())
					}

					// We're done, since it's not clear how to add an overlay
					return fmt.Errorf("Exiting encapsulated dicom")
				}

				for j := 0; j < len(frame.NativeData.Data); j++ {
					imgPixels = append(imgPixels, frame.NativeData.Data[j][0])
				}
			}

			maxIntensity := 0
			for _, v := range imgPixels {
				if v > maxIntensity {
					maxIntensity = v
				}
			}
			// Don't print the main image as text
			fmt.Println(elem.Tag, tagName.Name, "maxIntensity:", maxIntensity, "maxAllowed:", math.Pow(2., bitsStored), "~~skipping pixel data~~")
			continue
		}

		if elem.Tag.Compare(dicomtag.Tag{Group: 0x6000, Element: 0x3000}) == 0 {
			// Don't print the overlay as text
			fmt.Println(elem.Tag, tagName.Name, "~~skipping overlay pixel data~~")
			continue
		}

		if elem.Tag.Compare(dicomtag.Tag{Group: 0x0029, Element: 0x1020}) == 0 {
			// Don't print the secondary Siemens data.
			fmt.Println(elem.Tag, tagName.Name, "~~skipping secondary Siemens data~~")
			continue
		}

		// Siemens header data requires special treatment
		if elem.Tag.Compare(dicomtag.Tag{Group: 0x0029, Element: 0x1010}) == 0 {
			for _, v := range elem.Value {
				sc, err := bulkprocess.ParseSiemensHeader(v)
				if err != nil {
					return err
				}
				fmt.Printf("%v %v %s NElements: %+v\n", elem.Tag, tagName.Name, "SiemensHeader", sc.NElements)
				for _, v := range sc.Slice() {
					fmt.Printf("%v %v %s %+v\n", elem.Tag, tagName.Name, "SiemensHeader", v)
				}
			}

			continue
		}

		fmt.Println(elem.Tag, tagName.Name, elem.Value)
	}

	return nil
}
