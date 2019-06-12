package bulkprocess

import (
	"archive/zip"
	"log"
	"path"
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

type DicomOutput struct {
	SampleID  string
	ZipFile   string
	FieldID   string
	Instance  string
	Index     string
	Filename  string
	DicomMeta DicomMeta
}

func (d DicomOutput) ParsedDate() (time.Time, error) {
	res, err := dateparse.ParseAny(d.DicomMeta.Date)
	if err == nil {
		return res, nil
	}

	// Try some known values that dateparse fails to understand
	return time.Parse("02-Jan-2006 15:04:05", d.DicomMeta.Date)
}

func CardiacMRIZipIterator(zipPath string, processOne func(DicomOutput) error) (err error) {
	metadata, err := zipPathToMetadata(zipPath)
	if err != nil {
		return err
	}

	if metadata.SampleID == "" {
		return nil
	}

	zipName := path.Base(zipPath)

	rc, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	for _, v := range rc.File {
		// Looking only at the dicoms
		if strings.HasPrefix(v.Name, "manifest") {
			continue
		}

		// Some information comes from the zipfile itself
		dcm := DicomOutput{}
		dcm.SampleID = metadata.SampleID
		dcm.ZipFile = zipName
		dcm.FieldID = metadata.FieldID
		dcm.Instance = metadata.Instance
		dcm.Index = metadata.Index
		dcm.Filename = v.Name

		unzippedFile, err := v.Open()
		if err != nil {
			return err
		}
		meta, err := DicomToMetadata(unzippedFile)
		if err != nil {
			log.Println("Ignoring error and continuing:", err.Error())
			continue
		}

		dcm.DicomMeta = *meta

		if err := processOne(dcm); err != nil {
			return err
		}
	}

	return nil
}
