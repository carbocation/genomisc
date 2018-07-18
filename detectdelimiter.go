package genomisc

import (
	"io"

	"github.com/csimplestring/go-csv/detector"
)

var (
	sampleLines             = 15
	nonDelimiterRegexString = `[[:alnum:]\n\r@\.]`
)

// DetermineDelimiter returns the single most likely rune that would delimit the
// values in the reader, assuming a CSV-like file.
func DetermineDelimiter(r io.Reader) rune {
	d := detector.New()
	d.Configure(&sampleLines, &nonDelimiterRegexString)

	delimiters := d.DetectDelimiter(r, '"')

	if len(delimiters) > 0 {
		return rune(delimiters[0][0])
	}

	return ','
}
