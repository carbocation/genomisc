package main

import (
	"encoding/csv"
	"log"
	"os"
)

const (
	AnnoSampleIDCol = iota
	AnnoDicomCol
	AnnoValueCol
	AnnoDateCol
	AnnoPathCol
	DummyLastConst
)

type Annotation struct {
	Dicom    string
	SampleID string
	Value    string
	Date     string
	Path     string
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
		if len(row) != DummyLastConst {
			continue
		}

		// DICOM filenames are UUIDs and so are unique.
		extantAnnotations[DicomFilename(row[AnnoDicomCol])] = Annotation{
			Dicom:    row[AnnoDicomCol],
			SampleID: row[AnnoSampleIDCol],
			Value:    row[AnnoValueCol],
			Date:     row[AnnoDateCol],
			Path:     row[AnnoPathCol],
		}
	}

	return extantAnnotations, nil
}
