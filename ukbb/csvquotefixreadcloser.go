package ukbb

import (
	"bufio"
	"io"
	"strings"
)

// CSVQuoteFixReadCloser transparently replaces the invalid \" delimiter with ""
type CSVQuoteFixReadCloser struct {
	r        *bufio.Reader
	leftover *strings.Reader
	Close    func() error
}

func NewCSVQuoteFixReadCloser(r io.ReadCloser) *CSVQuoteFixReadCloser {
	return &CSVQuoteFixReadCloser{r: bufio.NewReader(r), leftover: &strings.Reader{}, Close: r.Close}
}

func (m *CSVQuoteFixReadCloser) Read(p []byte) (n int, err error) {
	if m.leftover.Len() == 0 {
		line, err := m.r.ReadString('\n')
		line = strings.ReplaceAll(line, "\\\"", "\"\"")
		if err != nil {
			return len(line), err
		}
		m.leftover = strings.NewReader(line)
	}

	n, err = m.leftover.Read(p)

	return
}
