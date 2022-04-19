package main

import (
	"log"
	"math"

	"github.com/carbocation/genomisc/cardiaccycle"
	"github.com/gonum/stat"
)

func runQC(samplesWithFlags SampleFlags, entries map[string]File, cycle []cardiaccycle.Result, nStandardDeviations float64) {

	// Flag entries that have 0 at any point in the cycle -- this should be
	// optional
	flagZeroes(entries)
	log.Println("Flagged entries with 0 pixels")

	// Flag entries that are above or below N-SD above or below the mean for
	// connected components
	mean, std := flagConnectedComponents(samplesWithFlags, entries, nStandardDeviations)
	log.Printf("connected components: mean %.5f, std %.5f. flagged entries beyond %.1f SD (below %.5f or above %.5f)", mean, std, nStandardDeviations, mean-std*nStandardDeviations, mean+std*nStandardDeviations)

	// Flag samples that are above or below N-SD above or below the mean for
	// onestep shifts in the pixel area between each timepoint
	mean, std = flagOnestepShifts(samplesWithFlags, cycle, nStandardDeviations)
	log.Printf("onestep pixel shift: mean %.5f, std %.5f. flagged entries beyond %.1f SD (below %.5f or above %.5f)", mean, std, nStandardDeviations, mean-std*nStandardDeviations, mean+std*nStandardDeviations)

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

func flagOnestepShifts(out SampleFlags, cycle []cardiaccycle.Result, nStandardDeviations float64) (mean, std float64) {

	value := make([]float64, 0, len(cycle))

	// Pass 1: populate the slice
	for _, entry := range cycle {
		// Do not include entries from flagged-bad samples in our set of
		// potentially valid values
		if len(out[entry.Identifier]) > 0 {
			continue
		}

		// Do not include entries with divide-by-zero errors (yielding NaN)
		if math.IsNaN(entry.MaxOneStepShift) {
			continue
		}

		value = append(value, entry.MaxOneStepShift)
	}

	mean, std = stat.MeanStdDev(value, nil)

	// Pass 2: flag entries that exceed the bounds:
	for _, entry := range cycle {
		if entry.MaxOneStepShift < mean-nStandardDeviations*std || entry.MaxOneStepShift > mean+nStandardDeviations*std {
			out.AddFlag(entry.Identifier, "OnestepShift")
		}
	}

	return
}

func flagConnectedComponents(samplesWithFlags SampleFlags, entries map[string]File, nStandardDeviations float64) (mean, std float64) {

	value := make([]float64, 0, len(entries))

	// Pass 1: populate the slice
	for _, entry := range entries {

		// Do not include entries from flagged-bad samples in our set of
		// potentially valid values
		if len(samplesWithFlags[entry.SampleID]) > 0 {
			continue
		}

		value = append(value, entry.ConnectedComponents)
	}

	mean, std = stat.MeanStdDev(value, nil)

	// Pass 2: flag entries that exceed the bounds:
	for k, entry := range entries {
		if entry.ConnectedComponents < mean-nStandardDeviations*std || entry.ConnectedComponents > mean+nStandardDeviations*std {
			entry.BadWhy = append(entry.BadWhy, "ConnectedComponents")
			entries[k] = entry
		}
	}

	return
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
