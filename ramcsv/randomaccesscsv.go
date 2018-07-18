package ramcsv

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
)

type locator struct {
	Offset int64
	Length int
}

type RAMCSV struct {
	offset int64       // the current offset
	m      []locator   // maps line numbers (key) to Offset and Length (value)
	rdr    *csv.Reader // to store settings
	file   *os.File
}

func NewRAMCSV(file *os.File, rdr *csv.Reader) *RAMCSV {
	file.Seek(0, 0) // Make sure our current offset is at the start of the file

	ram := RAMCSV{
		m:    make([]locator, 0),
		rdr:  rdr,
		file: file,
	}

	// To initialize, scan through the entire file once to identify the offsets
	// at each line.
	scanner := bufio.NewScanner(ram.file)
	scanner.Split(scanLinesNondestructive)
	var b []byte
	for scanner.Scan() {
		b = scanner.Bytes()

		ram.m = append(ram.m, locator{
			Offset: ram.offset,
			Length: len(b),
		})

		ram.offset += int64(len(b))
	}

	return &ram
}

func (ram *RAMCSV) Read(line int) ([]string, error) {
	if len(ram.m)-1 < line {
		return nil, fmt.Errorf("Line %d is greater than the length of the file (%d)", line, len(ram.m))
	}

	val := make([]byte, ram.m[line].Length)
	if _, err := ram.file.ReadAt(val, ram.m[line].Offset); err != nil {
		return nil, err
	}

	csvr := csv.NewReader(bytes.NewBuffer(val))
	csvr.Comma = ram.rdr.Comma
	csvr.Comment = ram.rdr.Comment
	csvr.FieldsPerRecord = ram.rdr.FieldsPerRecord
	csvr.LazyQuotes = ram.rdr.LazyQuotes
	csvr.ReuseRecord = ram.rdr.ReuseRecord
	csvr.TrailingComma = ram.rdr.TrailingComma
	csvr.TrimLeadingSpace = ram.rdr.TrimLeadingSpace

	return csvr.Read()
}

// scanLinesNondestructive does not destroy the \n or the possible \r\n from a
// line. Otherwise it is like
// https://golang.org/src/bufio/scan.go?s=11522:11600#L330
func scanLinesNondestructive(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
