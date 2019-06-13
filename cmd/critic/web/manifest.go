package main

import (
	"encoding/csv"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type DicomFilename string

type Manifest struct {
	SampleID              string
	Zip                   string
	Dicom                 string
	Series                string
	InstanceNumber        int
	HasOverlayFromProject bool
	Annotation            Annotation
}

type Annotation struct {
	Dicom    string
	SampleID string
	Value    string
}

func UpdateManifest() error {
	global.m.Lock()
	defer global.m.Unlock()

	// First, look in the project directory to get updates to annotations.
	suffix := ".png"
	files, err := ioutil.ReadDir(filepath.Join(".", global.Project))
	if os.IsNotExist(err) {
		// Not a problem
	} else if err != nil {
		return err
	}
	overlaysExist := make(map[string]struct{})
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		overlaysExist[f.Name()] = struct{}{}
	}

	updatedManifest := global.manifest

	// Toggle to the latest knowledge about the manifest
	for i, v := range updatedManifest {
		_, hasOverlay := overlaysExist[v.Zip+"_"+v.Dicom+suffix]
		updatedManifest[i].HasOverlayFromProject = hasOverlay
	}

	// TODO: ???

	return nil
}

// ReadManifestAndCreateOutput takes the path to a manifest file and extracts each line. It
// checks to see if the output file has already been created. If so, it reads
// the output and matches it with the manifest, returning pre-populated values.
func ReadManifestAndCreateOutput(manifestPath, annotationPath string) ([]Manifest, error) {

	// Fetch any annotations that already exist, or create a file to hold them
	// if none yet do.
	annotations, err := OpenOrCreateAnnotationFile(annotationPath)
	if err != nil {
		return nil, err
	}

	// Now open the full manifest of files we want to critique.
	f, err := os.Open(manifestPath)
	if err != nil {
		return nil, err
	}

	cr := csv.NewReader(f)
	cr.Comma = '\t'
	recs, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}

	output := make([]Manifest, 0, len(recs))

	header := struct {
		SampleID       int
		Zip            int
		Dicom          int
		Series         int
		InstanceNumber int
	}{}

	for i, cols := range recs {
		if i == 0 {
			for j, col := range cols {
				if col == "zip_file" {
					header.Zip = j
				} else if col == "dicom_file" {
					header.Dicom = j
				} else if col == "series" {
					header.Series = j
				} else if col == "instance_number" {
					header.InstanceNumber = j
				} else if col == "sample_id" {
					header.SampleID = j
				}
			}
			continue
		}

		intInstance, err := strconv.Atoi(cols[header.InstanceNumber])
		if err != nil {
			// Ignore the error
			intInstance = 0
		}

		anno, hasAnno := annotations[DicomFilename(cols[header.Dicom])]

		output = append(output, Manifest{
			SampleID:              cols[header.SampleID],
			Zip:                   cols[header.Zip],
			Dicom:                 cols[header.Dicom],
			Series:                cols[header.Series],
			InstanceNumber:        intInstance,
			HasOverlayFromProject: hasAnno,
			Annotation:            anno,
		})
	}

	sort.Slice(output, generateManifestSorter(output))

	return output, nil
}

func OpenOrCreateAnnotationFile(annotationPath string) (map[DicomFilename]Annotation, error) {
	// Create the annotation file, if it does not yet exist
	log.Printf("Creating %s if it does not yet exist\n", annotationPath)
	if err := CreateFileAndPath(annotationPath); err != nil {
		return nil, err
	}
	log.Println("Output will be stored at", annotationPath)

	// First, look in the annotation file to see if there is any annotation.
	annoFile, err := os.Open(annotationPath)
	if err != nil {
		return nil, err
	}
	defer annoFile.Close()

	// File format: tab-delimited, 3 columns: dicom_filename, sample_id, annotation
	cread := csv.NewReader(annoFile)
	cread.Comma = '\t'
	priorAnnotationCSV, err := cread.ReadAll()
	if err != nil {
		return nil, err
	}

	extantAnnotations := make(map[DicomFilename]Annotation)
	for i, row := range priorAnnotationCSV {
		if i == 0 {
			continue
		}
		if len(row) != 3 {
			continue
		}

		// DICOM filenames are UUIDs and so are unique.
		extantAnnotations[DicomFilename(row[0])] = Annotation{
			Dicom:    row[0],
			SampleID: row[1],
			Value:    row[2],
		}
	}

	return extantAnnotations, nil
}

// generateManifestSorter is a convenience function to help sort a list of
// manifest objects
func generateManifestSorter(output []Manifest) func(i, j int) bool {
	return func(i, j int) bool {
		// Need both conditions so that it will proceed to the next check only
		// if there is a tie at this one
		if output[i].Zip < output[j].Zip {
			return true
		} else if output[i].Zip > output[j].Zip {
			return false
		}

		if output[i].Series < output[j].Series {
			return true
		} else if output[i].Series > output[j].Series {
			return false
		}

		if output[i].InstanceNumber < output[j].InstanceNumber {
			return true
		} else if output[i].InstanceNumber > output[j].InstanceNumber {
			return false
		}

		if output[i].Dicom < output[j].Dicom {
			return true
		} else if output[i].Dicom > output[j].Dicom {
			return false
		}

		return false
	}
}
