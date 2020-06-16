// colqc performs qc on columnar data (the output of pixelcounter)
package main

import (
	"flag"
	"log"

	"github.com/carbocation/genomisc/cardiaccycle"
)

type File struct {
	Pixels              float64
	ConnectedComponents float64
	BadWhy              []string
	SampleID            string
	TimeID              float64
	PxHeight            float64
	PxWidth             float64
}

func (f File) CM2() float64 {
	return f.Pixels * f.PxHeight * f.PxWidth / 100.0
}

func main() {
	var pixels, connectedComponents string
	var pixelcountFile, covarFile string
	var sampleID string
	var imageID string
	var timeID string
	var pxHeight, pxWidth string

	flag.StringVar(&pixelcountFile, "pixelcountfile", "", "Path to file with pixelcount output (tab-delimited; containing imageid, value, connectedComponents)")
	flag.StringVar(&covarFile, "covarfile", "", "Path to file with covariates output (comma delimited; containing sampleid, imageid, timeid, pxheight, pxwidth)")
	flag.StringVar(&pixels, "pixels", "", "Column name that identifies the number of pixels belonging to your value of interest")
	flag.StringVar(&connectedComponents, "cc", "", "Column name that identifies the number of connected components belonging to your value of interest")
	flag.StringVar(&sampleID, "sampleid", "", "Column name that uniquely identifies a sample at a visit.")
	flag.StringVar(&imageID, "imageid", "", "Column name that identifies the image identifier.")
	flag.StringVar(&timeID, "timeid", "", "Column name that contains a time-ordered value for sorting the images for a given sampleID.")
	flag.StringVar(&pxHeight, "pxheight", "", "Column name that identifies the column with data converting pixel height to mm. (Optional.)")
	flag.StringVar(&pxWidth, "pxwidth", "", "Column name that identifies the column with data converting pixel width to mm. (Optional.)")

	flag.Parse()

	if pixelcountFile == "" {
		log.Fatalln("Please provide -pixelcountfile")
	}

	if covarFile == "" {
		log.Fatalln("Please provide -covarfile")
	}

	if sampleID == "" {
		log.Fatalln("Please provide -sampleid")
	}

	if pixels == "" {
		log.Fatalln("Please provide -pixels")
	}

	if connectedComponents == "" {
		log.Fatalln("Please provide -cc")
	}

	if imageID == "" {
		log.Fatalln("Please provide -imageid")
	}

	if timeID == "" {
		log.Fatalln("Please provide -timeid")
	}

	if pxHeight == "" {
		log.Fatalln("Please provide -pxheight")
	}

	if pxWidth == "" {
		log.Fatalln("Please provide -pxwidth")
	}

	log.Println("Launched pixelqc")

	if err := runAll(pixelcountFile, covarFile, pixels, connectedComponents, sampleID, imageID, timeID, pxHeight, pxWidth); err != nil {
		log.Fatalln(err)
	}
}

func runAll(pixelcountFile, covarFile, pixels, connectedComponents, sampleID, imageID, timeID, pxHeight, pxWidth string) error {

	// The primary output is one row per sample, with the greatest and smallest
	// pixel value encountered for that sample during the series, with
	// information on which samples seem to have bad data.
	samplesWithFlags := SampleFlags{}

	// Start by populating the entries from the pixelcount file; i.e., the file
	// that actually has computed pixel values
	entries, err := parsePixelcountFile(pixelcountFile, imageID, pixels, connectedComponents)
	if err != nil {
		return err
	}
	log.Println("Loaded", pixelcountFile)

	// Next, add in covariate metadata.
	err = parseCovarFile(entries, covarFile, sampleID, imageID, timeID, pxHeight, pxWidth)
	if err != nil {
		return err
	}
	log.Println("Loaded", covarFile)

	// Find the timepoint and values for the min and max in the cardiac cycle.
	// When doing so, look at the adjacent 2 values and discard 0 extremes.
	cycle, err := cardiacCycle(entries, 2, 0)
	if err != nil {
		return err
	}
	log.Println("Processed cyclic data across", len(cycle), "samples")

	runQC(samplesWithFlags, entries, cycle)

	return nil
}

func cardiacCycle(entries map[string]File, adjacentN, discardN int) ([]cardiaccycle.Result, error) {

	// Prep the inputs
	var sampleIDs, instances, metrics = []string{}, []uint16{}, []float64{}

	for _, v := range entries {
		sampleIDs = append(sampleIDs, v.SampleID)
		instances = append(instances, uint16(v.TimeID))
		metrics = append(metrics, v.CM2())
	}

	return cardiaccycle.RunFromSlices(sampleIDs, instances, metrics, adjacentN, discardN)
}
