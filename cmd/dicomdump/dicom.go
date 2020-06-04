package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/carbocation/pfx"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/dicomtag"
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

	parsedData, err := p.Parse(dicom.ParseOptions{
		DropPixelData: true,
	})
	if parsedData == nil || err != nil {
		return fmt.Errorf("Error reading dicom: %v", err)
	}

	for _, elem := range parsedData.Elements {
		tagName, _ := dicomtag.Find(elem.Tag)

		if tagName.Name == "" {
			tagName.Name = "____"
		}

		if elem.Tag.Compare(dicomtag.Tag{Group: 0x0029, Element: 0x1010}) == 0 {
			for _, v := range elem.Value {
				sc, err := ParseSiemensHeader(v)
				if err != nil {
					return err
				}
				for _, v := range sc {
					fmt.Printf("%v %v %s %+v\n", elem.Tag, tagName.Name, "SiemensHeader", v)
				}
			}

			continue
		}

		if elem.Tag.Compare(dicomtag.Tag{Group: 0x0029, Element: 0x1020}) == 0 {
			fmt.Println(elem.Tag, tagName.Name, "~~skipping value~~")
			continue
		}

		fmt.Println(elem.Tag, tagName.Name, elem.Value)
	}

	return nil
}
