package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/pfx"
)

type DicomFilename string

type AnnotationTracker struct {
	AnnotationPath string
	Nested         bool

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

		dicomName := v.Annotation.Dicom
		if dicomName == "" {
			dicomName = v.Dicom
		}

		fmt.Fprintf(f, "%s\t%s\t%s\t%s\t%s\n", v.SampleID, dicomName, v.Annotation.Value, v.Annotation.Date, v.Annotation.Path)
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

// ReadNestedManifest will attempt to parse a classical manifest. This is
// necessary when the image files are nested within a container.
func ReadNestedManifest(manifestPath, annotationPath, nestedSuffix string) (*AnnotationTracker, error) {

	mayBeNested := true

	// We assume that the Dicom filename for the merged image file has this
	// suffix. TODO: Make configurable.
	assumedMergedImageSuffix := ".png.overlay.png"

	// Fetch any annotations that already exist, or create a file to hold them
	// if none yet do.
	annotations, err := OpenOrCreateAnnotationFile(annotationPath)
	if err != nil {
		return nil, err
	}

	// Now open the full manifest of files we want to critique.
	f, _, err := bulkprocess.MaybeOpenFromGoogleStorage(manifestPath, global.storageClient)
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
	}{
		0,
		-1,
		-1,
		0,
		0,
	}

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

		if header.Dicom == -1 {
			return nil, fmt.Errorf("header column 'dicom_file' not detected")
		}

		// If Zip isn't found, this isn't an error, unless you intended to use a
		// nested structure
		if header.Zip == -1 {
			mayBeNested = false
			header.Zip = 0

			if nestedSuffix != "" {
				return nil, fmt.Errorf("header column 'zip_file' not detected")
			}
		}

		intInstance, err := strconv.Atoi(cols[header.InstanceNumber])
		if err != nil {
			// Ignore the error
			intInstance = 0
		}

		anno, _ := annotations[DicomFilename(cols[header.Dicom])]
		// Set the annotation DICOM name to be the raw dicom name
		anno.Dicom = cols[header.Dicom]

		output = append(output, ManifestEntry{
			SampleID:       cols[header.SampleID],
			Zip:            MergedNameToRawName(cols[header.Zip], nestedSuffix, ".zip"),
			Dicom:          cols[header.Dicom] + assumedMergedImageSuffix, // Total hack
			Series:         cols[header.Series],
			InstanceNumber: intInstance,
			Annotation:     anno,
		})
	}

	// For now, don't sort manifests - permits you to arrange them in a preprocessing step
	// sort.Slice(output, generateManifestSorter(output))

	return &AnnotationTracker{Entries: output, Nested: mayBeNested, AnnotationPath: annotationPath}, nil
}

// CreateManifestAndOutput lists all files in your main input directory and your
// output directory. It checks to see if the output file has already been
// created. If so, it reads the output and matches it with the input, returning
// pre-populated values.
func CreateManifestAndOutput(mergedPath, annotationPath, manifestPath, nestedSuffix string) (*AnnotationTracker, error) {

	// First, if we happen to be given a complete manifest, read it. If that
	// conforms to our old manifest format, use it. Just mark it non-nested.
	tracker, err := ReadNestedManifest(manifestPath, annotationPath, nestedSuffix)
	if err == nil {
		tracker.Nested = false
		return tracker, nil
	}

	// If we don't have an old-style manifest, process either the new simplified
	// manifest or, if none is given, process the folder contents.

	output := make([]ManifestEntry, 0)

	// Fetch any annotations that already exist, or create a file to hold them
	// if none yet do.
	annotations, err := OpenOrCreateAnnotationFile(annotationPath)
	if err != nil {
		return nil, err
	}

	// Now open the full manifest of files we want to critique.

	// If we actually passed a manifest file, assume it is valid and populate:
	if manifestPath != "" {
		rdr, _, err := bulkprocess.MaybeOpenFromGoogleStorage(manifestPath, global.storageClient)
		if err != nil {
			return nil, err
		}

		r := csv.NewReader(rdr)
		r.Comma = '\t'

		lines, err := r.ReadAll()
		if err != nil {
			return nil, err
		}

		for _, cols := range lines {
			entry := cols[0]
			anno, _ := annotations[DicomFilename(entry)]
			anno.Dicom = entry

			output = append(output, ManifestEntry{
				Dicom:      entry,
				Annotation: anno,
			})
		}

	} else {
		// If we did not pass a manifest file, list the directory and populate:
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
	}

	return &AnnotationTracker{Entries: output, AnnotationPath: annotationPath}, nil
}
