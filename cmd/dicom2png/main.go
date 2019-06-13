package main

import (
	"encoding/csv"
	"flag"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var inputPath, outputPath, manifest string
	flag.StringVar(&inputPath, "raw", "", "Path to the local folder containing the raw zip files")
	flag.StringVar(&outputPath, "out", "", "Path to the local folder where the extracted PNGs will go")
	flag.StringVar(&manifest, "manifest", "", "Manifest file containing Zip names and Dicom names.")

	flag.Parse()
	if inputPath == "" || outputPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	f, err := os.Open(manifest)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comma = '\t'
	entries, err := csvReader.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}

	zipFileCol, dicomFileCol := 0, 0

	for i, row := range entries {
		if i == 0 {
			for j, col := range row {
				if col == "zip_file" {
					zipFileCol = j
				} else if col == "dicom_file" {
					dicomFileCol = j
				}
			}

			continue
		}

		if err := ProcessOneFile(inputPath, outputPath, row[zipFileCol], row[dicomFileCol]); err != nil {
			log.Fatalln(err)
		}
	}

}

func ProcessOneFile(inputPath, outputPath, zipName, dicomName string) error {
	img, err := ExtractDicomFromLocalFile(filepath.Join(inputPath, zipName), dicomName, true)
	if err != nil {
		return err
	}

	f, err := os.Create(outputPath + "/" + dicomName + ".png")
	if err != nil {
		return err
	}

	return png.Encode(f, img)
}
