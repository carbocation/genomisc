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
var nobulk bool
var bulkColName string
var fileColName string

func main() {
	var presetFlag string

	flag.StringVar(&manifestPath, "manifest", "", "Path to manifest file. May be on Google Storage.")
	flag.BoolVar(&nobulk, "nobulk", false, "If true, looks for raw image files in the --files path, skipping the bulk file step")
	flag.StringVar(&filesPath, "files", "", "Folder path where bulk files reside. May be on Google Storage.")
	flag.StringVar(&outputPath, "output", "", "Local path into which the extracted files will go.")
	flag.StringVar(&bulkSuffixDel, "bs_del", "", "Suffix to delete from the manifest->bulk column.")
	flag.StringVar(&bulkSuffixAdd, "bs_add", "", "Suffix to add to the manifest->bulk column.")
	flag.StringVar(&fileSuffixDel, "fs_del", "", "Suffix to delete from the manifest->image file column.")
	flag.StringVar(&fileSuffixAdd, "fs_add", "", "Suffix to add to the manifest->image file column.")
	flag.StringVar(&bulkColName, "bulk_col_name", "zip_file", "Column name containing bulk file name (unless --nobulk is set)")
	flag.StringVar(&fileColName, "image_col_name", "dicom_file", "Column name containing image file name")
	flag.StringVar(&presetFlag, "preset", "", "If not empty, may be 'image' or 'overlay' to override the fs_* settings with defaults.")
	flag.Parse()

	var err error
	if strings.HasPrefix(filesPath, "gs://") ||
		strings.HasPrefix(manifestPath, "gs://") {
		sclient, err = storage.NewClient(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}

	if presetFlag == "image" {
		bulkSuffixDel = ".zip"
		bulkSuffixAdd = ".png.tar.gz"
		fileSuffixAdd = ".png"
	} else if presetFlag == "overlay" {
		bulkSuffixDel = ".zip"
		bulkSuffixAdd = ".overlay.tar.gz"
		fileSuffixAdd = ".png.mask.png"
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

	lastBulk := ""
	accumulatedImageNames := make(map[string]struct{})
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

		if header.Zip == -1 && !nobulk {
			return fmt.Errorf("header column '%s' not detected", bulkColName)
		}

		var bulkFileName string
		if nobulk {
			bulkFileName = ""
		} else {
			bulkFileName = FormatSuffix(cols[header.Zip], bulkSuffixDel, bulkSuffixAdd)
		}
		fileName := FormatSuffix(cols[header.Dicom], fileSuffixDel, fileSuffixAdd)

		if i == 1 {
			lastBulk = bulkFileName
			accumulatedImageNames[fileName] = struct{}{}
		} else if bulkFileName != lastBulk {

			// Safely reference the images we have accumulated so far
			safeMap := make(map[string]struct{})
			for k, v := range accumulatedImageNames {
				safeMap[k] = v
			}

			// Track the bulk file that those came from
			safeBulk := lastBulk

			sem <- struct{}{}
			if !nobulk {
				go func(filesPath, bulkFileName string, fileNames map[string]struct{}, outputPath string) {
					if err := DownloadFiles(filesPath, bulkFileName, fileNames, outputPath); err != nil {
						log.Println(err)
					}
					<-sem
				}(filesPath, safeBulk, safeMap, outputPath)
			}

			// Now set the new image map and bulk file
			lastBulk = bulkFileName
			accumulatedImageNames = make(map[string]struct{})
			accumulatedImageNames[fileName] = struct{}{}
		} else if bulkFileName == lastBulk {
			accumulatedImageNames[fileName] = struct{}{}
		}
	}

	// We most likely have additional accumulated files that need to be processed
	if len(accumulatedImageNames) > 0 {
		if nobulk {
			for fileName, _ := range accumulatedImageNames {
				sem <- struct{}{}
				go func(filesPath, fileName string, outputPath string) {
					if err := DownloadFilesNoBulk(filesPath, fileName, outputPath); err != nil {
						log.Println(err)
					}
					<-sem
				}(filesPath, fileName, outputPath)
			}
		} else {
			sem <- struct{}{}
			go func(filesPath, bulkFileName string, fileNames map[string]struct{}, outputPath string) {
				if err := DownloadFiles(filesPath, bulkFileName, fileNames, outputPath); err != nil {
					log.Println(err)
				}
				<-sem
			}(filesPath, lastBulk, accumulatedImageNames, outputPath)
		}
	}

	// Make sure we finish all the reads before we exit, otherwise we'll lose
	// the last `concurrency` lines.
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}

	return nil
}

func DownloadFilesNoBulk(filesPath, fileName string, outputPath string) error {
	sourcePath := fmt.Sprintf("%s/%s", filesPath, fileName)

	img, err := bulkprocess.MaybeExtractImageFromGoogleStorage(sourcePath, sclient)
	if err != nil {
		return err
	}

	// Save the images to disk
	outFile := fmt.Sprintf("%s/%s", outputPath, fileName)
	// log.Println(outFile)
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}

	// Write the PNG representation of our ID-encoded image to disk
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	f.Close()

	return nil
}

// DownloadFiles pulls down files
func DownloadFiles(filesPath, bulkFileName string, fileNames map[string]struct{}, outputPath string) error {
	if !strings.HasSuffix(bulkFileName, ".tar.gz") {
		return fmt.Errorf("currently only supports .tar.gz bulk files")
	}

	sourcePath := fmt.Sprintf("%s/%s", filesPath, bulkFileName)
	// log.Printf("Fetching %s::(%d files)", sourcePath, len(fileNames))

	imgs, err := bulkprocess.ExtractImagesFromTarGzMaybeFromGoogleStorage(sourcePath, fileNames, sclient)
	if err != nil {
		return err
	}

	// Save the images to disk
	for fileName, img := range imgs {
		outFile := fmt.Sprintf("%s/%s", outputPath, fileName)
		// log.Println(outFile)
		f, err := os.Create(outFile)
		if err != nil {
			return err
		}

		// Write the PNG representation of our ID-encoded image to disk
		if err := png.Encode(f, img); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	return nil
}

func FormatSuffix(word, remove, add string) string {
	return strings.TrimSuffix(word, remove) + add
}
