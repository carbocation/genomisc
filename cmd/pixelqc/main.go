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

	for file, entry := range entries {
		log.Printf("%s | %+v\n", file, entry)
		break
	}

	// Find the timepoint and values for the min and max in the cardiac cycle.
	// When doing so, look at the adjacent 2 values and discard 0 extremes.
	cycle, err := cardiacCycle(entries, 2, 0)

	for _, v := range cycle {
		if v.Identifier != "5690295_2_23" {
			continue
		}

		log.Printf("%+v\n", v)
		break
	}

	// Flag values that have 0 at any point in the cycle -- this should be optional
	flagZeroes(entries)

	for file, entry := range entries {
		if entry.Pixels >= math.SmallestNonzeroFloat64 {
			continue
		}
		log.Printf("%s | %+v\n", file, entry)
		break
	}

	// Flag entries that are above or below N-SD above or below the mean for
	// connected components
	flagConnectedComponents(entries, 5)

	// Flag samples that are above or below N-SD above or below the mean for
	// onestep shifts in the pixel area between each timepoint
	samplesWithFlags := make(map[string]struct{})
	flagOnestepShifts(samplesWithFlags, cycle, 5)

	// Flag samples that don't have the modal number of images
	flagAbnormalImageCounts(samplesWithFlags, entries)

	// Consolidate counts
	seenSamples := make(map[string]struct{})
	for _, v := range entries {
		seenSamples[v.SampleID] = struct{}{}
		if len(v.BadWhy) > 0 {
			samplesWithFlags[v.SampleID] = struct{}{}
		}
	}

	log.Println(len(samplesWithFlags), "samples out of", len(seenSamples), "have been flagged as potentially having invalid data")

	return nil
}

func flagAbnormalImageCounts(out map[string]struct{}, entries map[string]File) {
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
			out[k] = struct{}{}
		}
	}
}

func flagOnestepShifts(out map[string]struct{}, cycle []cardiaccycle.Result, nStandardDeviations float64) {

	value := make([]float64, 0, len(cycle))

	// Pass 1: populate the slice
	for _, entry := range cycle {
		value = append(value, entry.MaxOneStepShift)
	}

	m, s := stat.MeanStdDev(value, nil)

	// Pass 2: flag entries that exceed the bounds:
	for _, entry := range cycle {
		if entry.MaxOneStepShift < m-nStandardDeviations*s || entry.MaxOneStepShift > m+nStandardDeviations*s {
			out[entry.Identifier] = struct{}{}
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
			entry.BadWhy = append(entry.BadWhy, "Zero_pixels")
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
