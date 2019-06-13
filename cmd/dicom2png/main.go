package main

import (
	"encoding/csv"
	"flag"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
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

	concurrency := runtime.NumCPU()
	sem := make(chan bool, concurrency)

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

		sem <- true
		go func(zipStr, dicomStr string) {
			if err := ProcessOneFile(inputPath, outputPath, zipStr, dicomStr); err != nil {
				log.Println(err.Error(), "Skipping file...")
			}
			<-sem
		}(row[zipFileCol], row[dicomFileCol])
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
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
