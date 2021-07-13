package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"runtime"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
)

var sclient *storage.Client

var manifestPath string
var filesPath string
var outputPath string
var bulkSuffixDel string
var bulkSuffixAdd string
var fileSuffixDel string
var fileSuffixAdd string
var bulkColName = "zip_file"
var fileColName = "dicom_file"

func main() {
	flag.StringVar(&manifestPath, "manifest", "", "Path to manifest file. May be on Google Storage.")
	flag.StringVar(&filesPath, "files", "", "Folder path where bulk files reside. May be on Google Storage.")
	flag.StringVar(&outputPath, "output", "", "Local path into which the extracted files will go.")
	flag.StringVar(&bulkSuffixDel, "bs_del", "", "Suffix to delete from the manifest->bulk column.")
	flag.StringVar(&bulkSuffixAdd, "bs_add", "", "Suffix to add to the manifest->bulk column.")
	flag.StringVar(&fileSuffixDel, "fs_del", "", "Suffix to delete from the manifest->image file column.")
	flag.StringVar(&fileSuffixAdd, "fs_add", "", "Suffix to add to the manifest->image file column.")
	flag.Parse()

	var err error
	if strings.HasPrefix(filesPath, "gs://") ||
		strings.HasPrefix(manifestPath, "gs://") {
		sclient, err = storage.NewClient(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}

	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	// Now open the full manifest of files we want to critique.
	f, _, err := bulkprocess.MaybeOpenFromGoogleStorage(manifestPath, sclient)
	if err != nil {
		return err
	}
	defer f.Close()

	cr := csv.NewReader(f)
	cr.Comma = '\t'
	recs, err := cr.ReadAll()
	if err != nil {
		return err
	}

	header := struct {
		Zip   int
		Dicom int
	}{
		-1,
		-1,
	}

	sem := make(chan struct{}, runtime.NumCPU())

	for i, cols := range recs {
		if i == 0 {
			for j, col := range cols {
				if col == bulkColName {
					header.Zip = j
				} else if col == fileColName {
					header.Dicom = j
				}
			}
			continue
		}

		if header.Dicom == -1 {
			return fmt.Errorf("header column '%s' not detected", fileColName)
		}

		if header.Zip == -1 {
			return fmt.Errorf("header column '%s' not detected", bulkColName)
		}

		bulkFileName := FormatSuffix(cols[header.Zip], bulkSuffixDel, bulkSuffixAdd)
		fileName := FormatSuffix(cols[header.Dicom], fileSuffixDel, fileSuffixAdd)

		sem <- struct{}{}
		go func(filesPath, bulkFileName, fileName, outputPath string) {
			if err := DownloadFile(filesPath, bulkFileName, fileName, outputPath); err != nil {
				log.Println(err)
			}
			<-sem
		}(filesPath, bulkFileName, fileName, outputPath)
	}

	// Make sure we finish all the reads before we exit, otherwise we'll lose
	// the last `concurrency` lines.
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}

	return nil
}

func DownloadFile(filesPath, bulkFileName, fileName, outputPath string) error {
	if !strings.HasSuffix(bulkFileName, ".tar.gz") {
		return fmt.Errorf("currently only supports .tar.gz bulk files")
	}

	sourcePath := fmt.Sprintf("%s/%s", filesPath, bulkFileName)
	// log.Printf("Fetching %s::%s", sourcePath, fileName)

	img, err := bulkprocess.ExtractImageFromTarGzMaybeFromGoogleStorage(sourcePath, fileName, sclient)
	if err != nil {
		return err
	}

	// Save the BMP to disk under your project folder, using the ID-encoding.
	outFile := fmt.Sprintf("%s/%s", outputPath, fileName)
	// log.Println(outFile)
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the PNG representation of our ID-encoded image to disk
	return png.Encode(f, img)
}

func FormatSuffix(word, remove, add string) string {
	return strings.TrimSuffix(word, remove) + add
}
