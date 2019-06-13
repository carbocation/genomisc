package main

import (
	"encoding/csv"
	"os"
	"sort"
	"strconv"
	"sync"
)

type DicomFilename string

type Manifest struct {
	m       sync.RWMutex
	Entries []ManifestEntry
}

func (m *Manifest) GetEntries() []ManifestEntry {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.Entries
}

func (m *Manifest) SaveAnnotations() {

}

type ManifestEntry struct {
	SampleID       string
	Zip            string
	Dicom          string
	Series         string
	InstanceNumber int
	Annotation     Annotation
}

// ReadManifestAndCreateOutput takes the path to a manifest file and extracts each line. It
// checks to see if the output file has already been created. If so, it reads
// the output and matches it with the manifest, returning pre-populated values.
func ReadManifestAndCreateOutput(manifestPath, annotationPath string) (*Manifest, error) {

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
	defer f.Close()

	cr := csv.NewReader(f)
	cr.Comma = '\t'
	recs, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}

	output := make([]ManifestEntry, 0, len(recs))

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

		anno, _ := annotations[DicomFilename(cols[header.Dicom])]

		output = append(output, ManifestEntry{
			SampleID:       cols[header.SampleID],
			Zip:            cols[header.Zip],
			Dicom:          cols[header.Dicom],
			Series:         cols[header.Series],
			InstanceNumber: intInstance,
			Annotation:     anno,
		})
	}

	sort.Slice(output, generateManifestSorter(output))

	return &Manifest{Entries: output}, nil
}

// generateManifestSorter is a convenience function to help sort a list of
// manifest objects
func generateManifestSorter(output []ManifestEntry) func(i, j int) bool {
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
