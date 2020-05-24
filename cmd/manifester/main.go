package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/carbocation/genomisc"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
)

var (
	BufferSize = 4096
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
)

// Makes one big combined manifest
// Emits to stdout
func main() {
	defer STDOUT.Flush()

	var path, output string
	var filetypes string

	flag.StringVar(&path, "path", "", "Path where the UKBB bulk .zip files are being held.")
	flag.StringVar(&filetypes, "type", "dicomzip", "File type. Options include 'dicomzip' (default), '12leadekg', and 'exerciseekg' (for EKG data).")
	flag.StringVar(&output, "output", "", "Output file. If blank, output will go to STDOUT.")
	flag.Parse()

	if path == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Replace the output target with a file, if desired
	if output != "" {
		outF, err := os.Create(genomisc.ExpandHome(output))
		if err != nil {
			log.Fatalln(err)
		}
		defer outF.Close()

		STDOUT = bufio.NewWriterSize(outF, BufferSize)
		defer STDOUT.Flush()
	}

	if filetypes == "exerciseekg" {
		if err := ManifestForExerciseEKG(path); err != nil {
			log.Fatalln(err)
		}

		return
	}

	if filetypes == "12leadekg" {
		if err := ManifestFor12LeadEKG(path); err != nil {
			log.Fatalln(err)
		}

		return
	}

	if filetypes != "dicomzip" {
		log.Println("Requested filetype '' not recognized")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := ManifestForDicom(path); err != nil {
		log.Fatalln(err)
	}
}

func PrintCSVRow(row bulkprocess.DicomOutput, results chan<- string) error {
	studyDate, err := row.ParsedDate()
	if err != nil {
		log.Println("Ignoring date parsing error:", err)
		// return err
		studyDate = time.Time{}
	}

	overlayText := "NoOverlay"
	if row.DicomMeta.HasOverlay {
		overlayText = "HasOverlay"
	}

	results <- fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%.8f\t%d\t%d\t%d\t%d\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%s\t%s\t%s\t%s\t%s\t%v\t%s\t%s\t%s",
		row.SampleID, row.FieldID, row.Instance, row.Index, row.ZipFile,
		row.Filename, row.DicomMeta.SeriesDescription, studyDate.Format("2006-01-02"),
		row.DicomMeta.InstanceNumber, overlayText, row.DicomMeta.OverlayFraction, row.DicomMeta.OverlayRows, row.DicomMeta.OverlayCols,
		row.DicomMeta.Rows, row.DicomMeta.Cols,
		row.DicomMeta.PatientX, row.DicomMeta.PatientY, row.DicomMeta.PatientZ, row.DicomMeta.PixelHeightMM, row.DicomMeta.PixelWidthMM,
		row.DicomMeta.SliceThicknessMM,
		row.DicomMeta.SeriesNumber, row.DicomMeta.AcquisitionNumber, row.DicomMeta.DeviceSerialNumber, row.DicomMeta.StationName, row.DicomMeta.SoftwareVersions,
		row.DicomMeta.EchoTime, row.DicomMeta.NominalInterval, row.DicomMeta.SliceLocation, row.DicomMeta.TriggerTime)
	return nil
}
