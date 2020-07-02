package main

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/carbocation/genomisc/overlay"
)

type maskMap struct {
	VencDicom string
	CineDicom string
}

func runFromManifest(manifest, zipPath, maskFolder, maskSuffix string, config overlay.JSONConfig) error {
	zipMap, err := getZipMap(manifest)
	if err != nil {
		return err
	}

	// Now iterate over the zips and print out the requested dicoms as images

	concurrency := 1 //runtime.NumCPU()
	sem := make(chan bool, concurrency)

	for zipFile, dicomList := range zipMap {
		sem <- true
		go func(zipFile string, dicoms []maskMap) {

			// The main purpose of this loop is to handle a specific filesystem
			// error (input/output error) that largely happens with GCSFuse, and
			// retry a few times before giving up.
			for loadAttempts, maxLoadAttempts := 1, 10; loadAttempts <= maxLoadAttempts; loadAttempts++ {

				out, err := processOneZipFile(zipPath, zipFile, dicoms, maskFolder, maskSuffix, config)

				if err != nil && loadAttempts == maxLoadAttempts {
					// We've exhausted our retries. Fail hard.
					log.Fatalln(err)
				} else if err != nil {
					log.Println("Sleeping 5s to recover from", err.Error(), ". Attempt #", loadAttempts)
					time.Sleep(5 * time.Second)
					continue
				}

				// Only now that we know we completed this file without error
				// can we print; printing before this moment will lead to
				// duplicate output if there are retries after the first line is
				// emitted.
				fmt.Fprint(STDOUT, out)

				// If no error, we don't need to retry
				break
			}
			<-sem

		}(zipFile, dicomList)
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return nil
}

func processOneZipFile(zipPath, zipFile string, dicoms []maskMap, maskFolder, maskSuffix string, config overlay.JSONConfig) (string, error) {
	f, err := os.Open(filepath.Join(zipPath, zipFile))
	if err != nil {
		return "", fmt.Errorf("ProcessOneZipFile fatal error (terminating on zip %s): %v", zipFile, err)
	}
	defer f.Close()

	// the zip reader wants to know the # of bytes in advance
	nBytes, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("ProcessOneZipFile fatal error (terminating on zip %s): %v", zipFile, err)
	}

	sb := strings.Builder{}

	// The zip is now open, so we don't have to reopen/reclose for every dicom
	for _, dicomNames := range dicoms {
		out, err := func(dicomPair maskMap) (string, error) {

			// TODO: Can consider optimizing further. This is an o(n^2)
			// operation - but if you printed the Dicom image to PNG within this
			// function, you could make it accept a map and then only iterate in
			// o(n) time. Not sure this is a bottleneck yet.
			dcmReadSeeker, err := extractDicomReaderFromZip(f, nBytes.Size(), dicomPair.VencDicom)
			if err != nil {
				return "", err
			}

			// Load the mask as an image. The mask comes from the cine dicom
			// while we will apply it to the phase (VENC) dicom.
			maskPath := filepath.Join(maskFolder, dicomPair.CineDicom+maskSuffix)
			rawOverlayImg, err := overlay.OpenImageFromLocalFile(maskPath)
			if err != nil {
				return "", err
			}

			str, err := run(dcmReadSeeker, rawOverlayImg, dicomNames.VencDicom, config)
			if err != nil {
				return "", err
			}
			return str, nil
		}(dicomNames)

		// Since we want to continue processing the zip even if one of its
		// dicoms was bad, simply log the error. However, if the underlying
		// file system is unreliable, we should then exit.
		if err != nil && strings.Contains(err.Error(), "input/output error") {

			// If we had an error due to an unreliable filesystem, we need to
			// fail the whole job or we will end up with unreliable or missing
			// data.
			return "", fmt.Errorf("ProcessOneZipFile fatal error (terminating on dicom %s in zip %s): %v", dicomNames.VencDicom, zipFile, err)

		} else if err != nil {

			// Non-i/o errors usually indicate that this is just a bad internal
			// file that will never be readable, in which case we should move on
			// rather than quitting.
			log.Printf("ProcessOneZipFile error (skipping dicom %s in zip %s): %v\n", dicomNames.VencDicom, zipFile, err)

		}

		sb.WriteString(out)
	}

	return sb.String(), nil
}

func extractDicomReaderFromZip(zipReaderAt io.ReaderAt, zipNBytes int64, dicomName string) (*bytes.Reader, error) {
	var err error

	rc, err := zip.NewReader(zipReaderAt, zipNBytes)
	if err != nil {
		return nil, err
	}

	for _, v := range rc.File {
		// Iterate over all of the dicoms in the zip til we find the one with
		// the desired name
		if v.Name != dicomName {
			continue
		}

		dcmReadCloser, err := v.Open()
		if err != nil {
			return nil, err
		}
		defer dcmReadCloser.Close()

		// Convert our readCloser to a readSeeker
		dcmBytes, err := ioutil.ReadAll(dcmReadCloser)
		if err != nil {
			return nil, err
		}

		return bytes.NewReader(dcmBytes), nil

	}

	return nil, fmt.Errorf("Did not find the requested Dicom %s", dicomName)
}

func getZipMap(manifest string) (map[string][]maskMap, error) {
	f, err := os.Open(manifest)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comma = '\t'
	entries, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	zipFileCol, vencDicomCol, cineDicomCol := -1, -1, -1

	// First, identify whether we are extracting multiple images from any zips.
	// If so, it will be more efficient to open the zip one time and extract the
	// desired images, rather than opening/closing the zip for each image
	// (especially if over gcsfuse)
	zipMap := make(map[string][]maskMap) // map[zip_file][]dicom_file
	for i, row := range entries {
		if i == 0 {
			for j, col := range row {
				if col == "zip_file" {
					zipFileCol = j
				} else if col == "dicom_cine" {
					cineDicomCol = j
				} else if col == "dicom_venc" {
					vencDicomCol = j
				}
			}

			continue
		} else if zipFileCol < 0 || vencDicomCol < 0 || cineDicomCol < 0 {
			return nil, fmt.Errorf("Did not identify zip_file, dicom_cine, or dicom_venc in the header line of %s", manifest)
		}

		// Append to this zip file's list of individual dicom images to process
		zipMap[row[zipFileCol]] = append(zipMap[row[zipFileCol]], maskMap{CineDicom: row[cineDicomCol], VencDicom: row[vencDicomCol]})
	}

	return zipMap, nil
}
