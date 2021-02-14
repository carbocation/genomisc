package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/carbocation/pfx"
)

type DicomFilename string

type AnnotationTracker struct {
	AnnotationPath string

	m       sync.RWMutex
	Entries []ManifestEntry
}

func (m *AnnotationTracker) GetEntries() []ManifestEntry {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.Entries
}

func (m *AnnotationTracker) SetAnnotation(manifestIdx int, value string) error {
	m.m.Lock()
	defer m.m.Unlock()

	err := fmt.Errorf("manifest entry #%d was not found", manifestIdx)

	if len(m.Entries)-1 < manifestIdx {
		return err
	}

	chosen := m.Entries[manifestIdx]
	chosen.Annotation.Date = time.Now().Format(time.RFC3339)
	chosen.Annotation.Value = value

	// Make sure the main array gets the values
	m.Entries[manifestIdx] = chosen

	return nil
}

func (m *AnnotationTracker) WriteAnnotationsToDisk(imagePath string) error {
	m.m.Lock()
	defer m.m.Unlock()

	// f, err := os.Open(m.AnnotationPath)
	f, err := os.OpenFile(m.AnnotationPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "sample_id\tdicom\tvalue\tdate\tpath\n")
	for _, v := range m.Entries {
		if v.Annotation.Value == "" {
			continue
		}

		if v.Annotation.Path == "" {
			v.Annotation.Path = imagePath
		}

		fmt.Fprintf(f, "%s\t%s\t%s\t%s\t%s\n", v.SampleID, v.Dicom, v.Annotation.Value, v.Annotation.Date, v.Annotation.Path)
	}

	return nil
}

type ManifestEntry struct {
	SampleID       string
	Zip            string
	Dicom          string
	Series         string
	InstanceNumber int
	Annotation     Annotation
}

// CreateManifestAndOutput lists all files in your main input directory and your
// output directory. It checks to see if the output file has already been
// created. If so, it reads the output and matches it with the input, returning
// pre-populated values.
func CreateManifestAndOutput(mergedPath, annotationPath string) (*AnnotationTracker, error) {

	output := make([]ManifestEntry, 0)

	// Fetch any annotations that already exist, or create a file to hold them
	// if none yet do.
	annotations, err := OpenOrCreateAnnotationFile(annotationPath)
	if err != nil {
		return nil, err
	}

	// Now open the full manifest of files we want to critique.
	files, err := ioutil.ReadDir(mergedPath)
	if err != nil {
		return nil, pfx.Err(err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		anno, _ := annotations[DicomFilename(f.Name())]

		output = append(output, ManifestEntry{
			Dicom:      f.Name(),
			Annotation: anno,
		})
	}

	return &AnnotationTracker{Entries: output, AnnotationPath: annotationPath}, nil
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