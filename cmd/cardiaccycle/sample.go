package main

// SeriesSample represents a sample-series pair. Most samples have just one series,
// but for the samples with more than one, one of the series is usually bad -
// typically the first. But at least, you don't want to accidentally model them
// jointly.
type SeriesSample struct {
	SampleID     string
	SeriesNumber string
}
