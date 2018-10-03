package cardiacmri

import (
	"fmt"
)

type DicomOutput struct {
	SampleID string
	ZipFile  string
	Dicom    DicomRow
}

type DicomRow struct {
	Filename          string
	PatientID         string
	StudyID           string
	StudyDescription  string
	Date              string
	SeriesID          string
	SeriesDescription string
	Modality          string // Not always present
	AET               string
	Host              string
}

func stringSliceToDicomStruct(input []string) (out DicomOutput, err error) {
	if l := len(input); l < 9 || l > 10 {
		return out, fmt.Errorf("Expected 9 or 10 fields, found %d", l)
	}

	out.Dicom.Filename = input[0]
	out.Dicom.PatientID = input[1]
	out.Dicom.StudyID = input[2]
	out.Dicom.StudyDescription = input[3]
	out.Dicom.Date = input[4]
	out.Dicom.SeriesID = input[5]
	out.Dicom.SeriesDescription = input[6]

	if len(input) == 10 {
		out.Dicom.Modality = input[7]
		out.Dicom.AET = input[8]
		out.Dicom.Host = input[9]
	} else {
		out.Dicom.AET = input[7]
		out.Dicom.Host = input[8]
	}

	return
}
