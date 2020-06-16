// colqc performs qc on columnar data (the output of pixelcounter)
package main

import (
	"flag"
	"log"
	"math"

	"github.com/carbocation/genomisc/cardiaccycle"
	"github.com/gonum/stat"
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

	// Flag entries that have 0 at any point in the cycle -- this should be
	// optional
	flagZeroes(entries)
	log.Println("Flagged entries with 0 pixels")

	// Flag entries that are above or below N-SD above or below the mean for
	// connected components
	SD := 5.0
	flagConnectedComponents(entries, SD)
	log.Println("Flagged entries beyond", SD, "standard deviations above or below the mean connected components")

	// Flag samples that are above or below N-SD above or below the mean for
	// onestep shifts in the pixel area between each timepoint
	SD = 5.0
	flagOnestepShifts(samplesWithFlags, cycle, SD)
	log.Println("Flagged entries beyond", SD, "standard deviations above or below the mean onstep pixel shift")

	// Flag samples that don't have the modal number of images
	flagAbnormalImageCounts(samplesWithFlags, entries)
	log.Println("Flagged entries that didn't have the modal number of images")

	// Consolidate counts
	seenSamples := make(map[string]struct{})
	for _, v := range entries {
		seenSamples[v.SampleID] = struct{}{}
		for _, bad := range v.BadWhy {
			samplesWithFlags.AddFlag(v.SampleID, bad)
		}
	}

	// Number of samples with each flag:
	flagCounts := make(map[string]int)
	for _, flags := range samplesWithFlags {
		for v := range flags {
			flagCounts[v]++
		}
	}

	log.Println(len(samplesWithFlags), "samples out of", len(seenSamples), "have been flagged as potentially having invalid data")
	log.Printf("Number of samples with each flag: %+v\n", flagCounts)

	return nil
}

func flagAbnormalImageCounts(out SampleFlags, entries map[string]File) {
	sampleCounts := make(map[string]int)
	countCounts := make(map[int]int)

	// Tally up the number of images for each sample
	for _, entry := range entries {
		sampleCounts[entry.SampleID]++
	}

	// Count the number of samples with each discrete number of images
	for _, sampleCount := range sampleCounts {
		countCounts[sampleCount]++
	}

	// Find the modal number of images per sample
	var modalCount, maxCount = -1, -1
	for imagesPerSample, samplesWithThisImageCount := range countCounts {
		if samplesWithThisImageCount > maxCount {
			modalCount = imagesPerSample
			maxCount = samplesWithThisImageCount
		}
	}

	// Flag samples that don't have the modal image count
	for k, count := range sampleCounts {
		if count != modalCount {
			out.AddFlag(k, "AbnormalImageCount")
		}
	}
}

func flagOnestepShifts(out SampleFlags, cycle []cardiaccycle.Result, nStandardDeviations float64) {

	value := make([]float64, 0, len(cycle))

	// Pass 1: populate the slice
	for _, entry := range cycle {
		value = append(value, entry.MaxOneStepShift)
	}

	m, s := stat.MeanStdDev(value, nil)

	// Pass 2: flag entries that exceed the bounds:
	for _, entry := range cycle {
		if entry.MaxOneStepShift < m-nStandardDeviations*s || entry.MaxOneStepShift > m+nStandardDeviations*s {
			out.AddFlag(entry.Identifier, "OnestepShift")
		}
	}
}

func flagConnectedComponents(entries map[string]File, nStandardDeviations float64) {

	value := make([]float64, 0, len(entries))

	// Pass 1: populate the slice
	for _, entry := range entries {
		value = append(value, entry.ConnectedComponents)
	}

	m, s := stat.MeanStdDev(value, nil)

	// Pass 2: flag entries that exceed the bounds:
	for k, entry := range entries {
		if entry.ConnectedComponents < m-nStandardDeviations*s || entry.ConnectedComponents > m+nStandardDeviations*s {
			entry.BadWhy = append(entry.BadWhy, "ConnectedComponents")
			entries[k] = entry
		}
	}
}

func flagZeroes(entries map[string]File) {

	// Identify the samples which have *any* frames with 0 pixels
	for k, entry := range entries {
		if entry.CM2() < math.SmallestNonzeroFloat64 {
			entry.BadWhy = append(entry.BadWhy, "ZeroPixels")
			entries[k] = entry
		}
	}

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
