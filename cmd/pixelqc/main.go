// colqc performs qc on columnar data (the output of pixelcounter)
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

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
	return f.Pixels * f.PxHeight * f.PxWidth / divisor
}

var divisor float64

func main() {

	fmt.Fprintf(os.Stderr, "%q\n", os.Args)

	var pixels, connectedComponents string
	var pixelcountFile, covarFile string
	var sampleID string
	var imageID string
	var timeID string
	var pxHeight, pxWidth string
	var nStandardDeviations float64

	flag.StringVar(&pixelcountFile, "pixelcountfile", "", "Path to file with pixelcount output (tab-delimited; containing imageid, value, connectedComponents)")
	flag.StringVar(&covarFile, "covarfile", "", "Path to file with covariates output (comma delimited; containing sampleid, imageid, timeid, pxheight, pxwidth)")
	flag.StringVar(&pixels, "pixels", "", "Column name that identifies the number of pixels belonging to your value of interest")
	flag.StringVar(&connectedComponents, "cc", "", "Column name that identifies the number of connected components belonging to your value of interest")
	flag.StringVar(&sampleID, "sampleid", "", "Column name that uniquely identifies a sample at a visit.")
	flag.StringVar(&imageID, "imageid", "", "Column name that identifies the image identifier.")
	flag.StringVar(&timeID, "timeid", "", "Column name that contains a time-ordered value for sorting the images for a given sampleID.")
	flag.StringVar(&pxHeight, "pxheight", "", "Column name that identifies the column with data converting pixel height to mm. (Optional.)")
	flag.StringVar(&pxWidth, "pxwidth", "", "Column name that identifies the column with data converting pixel width to mm. (Optional.)")
	flag.Float64Var(&nStandardDeviations, "sd", 5.0, "Number of standard deviations beyond which to consider our metrics to have failed QC.")
	flag.Float64Var(&divisor, "divisor", 100.0, "Divide output by this value (e.g., divide by 100.0 if input was mm^2 and goal output is cm^2, or 1.0 if input was cm and no adjustment is desired).")

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
		log.Println("pxHeight not provided; assuming that pxHeight == 1")
	}

	if pxWidth == "" {
		log.Println("pxWidth not provided; assuming that pxWidth == 1")
	}

	log.Println("Launched pixelqc for", pixels)

	if err := runAll(pixelcountFile, covarFile, pixels, connectedComponents, sampleID, imageID, timeID, pxHeight, pxWidth, nStandardDeviations); err != nil {
		log.Fatalln(err)
	}
}

func runAll(pixelcountFile, covarFile, pixels, connectedComponents, sampleID, imageID, timeID, pxHeight, pxWidth string, nStandardDeviations float64) error {

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
	log.Println("Loaded pixel data on ", len(entries), "images from", pixelcountFile)

	// Next, add in covariate metadata.
	if covarFile != "" {
		err = parseCovarFile(entries, covarFile, sampleID, imageID, timeID, pxHeight, pxWidth)
		if err != nil {
			return err
		}
		log.Println("Loaded covariate data from", covarFile)
	}

	// Find the timepoint and values for the min and max in the cardiac cycle.
	// When doing so, look at the adjacent 2 values and discard 0 extremes.
	cycle, err := cardiacCycle(entries, 2, 0)
	if err != nil {
		return err
	}
	log.Println("Processed cyclic data across", len(cycle), "samples")

	// Run all QC

	log.Println("Will run QC iteratively until no additional samples are flagged in an iteration.")

	flagged := samplesWithFlags.CountFlagged()
	for i := 1; ; i++ {

		log.Println("QC iteration", i)

		runQC(samplesWithFlags, entries, cycle, nStandardDeviations)

		newFlagged := samplesWithFlags.CountFlagged()
		if newFlagged == flagged {
			break
		}

		log.Printf("Flagged %d new bad samples this QC iteration; iterating again.\n", newFlagged-flagged)
		flagged = newFlagged
	}

	// Finally, print the results

	fmt.Println(strings.Join([]string{
		"sampleid",
		"metric",
		"connectedcomponents",
		"min",
		"max",
		"onestepshift",
		"timeid_min",
		"timeid_max",
		"bad",
	}, "\t"))

	sort.Slice(cycle, func(i, j int) bool { return cycle[i].Identifier < cycle[j].Identifier })

	for _, v := range cycle {
		out := []string{
			v.Identifier,
			pixels,
			connectedComponents,
			strconv.FormatFloat(v.Min, 'g', 4, 64),
			strconv.FormatFloat(v.Max, 'g', 4, 64),
			strconv.FormatFloat(v.MaxOneStepShift, 'g', 4, 64),
			strconv.FormatUint(uint64(v.InstanceNumberAtMin), 10),
			strconv.FormatUint(uint64(v.InstanceNumberAtMax), 10),
			samplesWithFlags[v.Identifier].String(),
		}

		fmt.Println(strings.Join(out, "\t"))
	}

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
