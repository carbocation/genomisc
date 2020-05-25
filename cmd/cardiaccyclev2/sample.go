package main

// SeriesSample represents a sample-instance-series pair. Most samples have just
// one series (set of images) occurring at once instance (imaging visit).
// However, thousands of samples will have at least a second instance.
// Additionally, many samples have multiple series within one instance. This
// appears to be due to problems with image acquisition in the first series. You
// don't want to accidentally model them jointly.
type SeriesSample struct {
	Identifier string
}
