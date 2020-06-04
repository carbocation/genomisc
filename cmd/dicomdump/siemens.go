package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
)

// The instructions for parsing the Siemens header are given in pseudocode here:
// https://scion.duhs.duke.edu/vespa/project/raw-attachment/wiki/SiemensCsaHeaderParsing/csa.html
// or cached version
// https://webcache.googleusercontent.com/search?q=cache:EbsOluxWNWIJ:https://scion.duhs.duke.edu/vespa/project/raw-attachment/wiki/SiemensCsaHeaderParsing/csa.html+&cd=4&hl=en&ct=clnk&gl=us

type SiemensHeader struct {
	NElements int
	Elements  map[string]SiemensChunk
}

func (s SiemensHeader) Slice() []SiemensChunk {
	orderedSlice := make([]SiemensChunk, 0, len(s.Elements))

	for _, v := range s.Elements {
		orderedSlice = append(orderedSlice, v)
	}

	sort.Slice(orderedSlice, func(i, j int) bool {
		return orderedSlice[i].Order < orderedSlice[j].Order
	})

	return orderedSlice
}

func (s SiemensHeader) String() string {
	orderedSlice := s.Slice()

	return fmt.Sprintf("NElements: %d Elements: %+v", s.NElements, orderedSlice)
}

type SiemensChunk struct {
	Name           string
	VM             uint32
	VR             string
	SyngoDT        uint32
	Subelements    uint32
	SubElementData []string
	Order          int
}

// SiemensDelimiter is defined and used, but we don't actually need to be aware
// of it to parse correctly
var SiemensDelimiter = []byte{0x4d, 00, 00, 00}

// SiemensDelimiter2 is defined and used, but we don't actually need to be aware
// of it to parse correctly
var SiemensDelimiter2 = []byte{0xcd, 00, 00, 00}

// ParseSiemensHeader consumes the element corresponding to the Siemens header
// and parses it fully.
func ParseSiemensHeader(v interface{}) (SiemensHeader, error) {
	sData := SiemensHeader{
		Elements: make(map[string]SiemensChunk),
	}

	// The data is a byte stream that is encoded into 4-byte words.
	// Specifically, they are encoded as little-endian 32-bit words.
	bs := v.([]uint8)
	bread := bytes.NewReader(bs)

	word := make([]byte, 4)
	var offset int64

	// Confirmed the SV10 header
	if _, err := bread.ReadAt(word, offset); err != nil {
		return sData, err
	}
	offset += int64(len(word))

	if string(word) != "SV10" {
		return sData, fmt.Errorf("Siemens data didn't start with SV10")
	}

	// Check for the 04030201 constant
	bread.ReadAt(word, offset)
	offset += int64(len(word))
	if cslc := []byte{04, 03, 02, 01}; !bytes.Equal(cslc, word) {
		return sData, fmt.Errorf("Didn't find constant %v", cslc)
	}

	bread.ReadAt(word, offset)
	offset += int64(len(word))
	nElements := binary.LittleEndian.Uint32(word)
	sData.NElements = int(nElements)

	// Read the delimiter
	bread.ReadAt(word, offset)
	offset += int64(len(word))

	// Read the first named chunk
	var sc SiemensChunk

	i := 0

	// Since each Name field is 64 bytes, if the remaining data wouldn't be
	// enough for a Name field, we know we are done.
	for offset < (int64(len(bs)) - 64) {
		sc, offset = ReadChunk(bread, offset)
		sc.Order = i
		sData.Elements[sc.Name] = sc

		i++
	}

	return sData, nil
}

// ReadChunk reads a named element and any subelements from the Siemens header.
func ReadChunk(bread *bytes.Reader, offset int64) (SiemensChunk, int64) {
	out := SiemensChunk{}

	// Most words are 4 bytes
	word := make([]byte, 4)

	// Read the element name. 64 bytes, ignore after the first 0x00
	eword := make([]byte, 64)
	bread.ReadAt(eword, offset)
	offset += int64(len(eword))

	// Eliminate data from the name that comes after the first 0x00

	// fmt.Println(string(eword))
	wp := bytes.Split(eword, []byte{0x00})
	// fmt.Println(string(wp[0]))

	out.Name = string(wp[0])

	// Next, read the info about this name:

	// VM
	bread.ReadAt(word, offset)
	offset += 4
	out.VM = binary.LittleEndian.Uint32(word)

	// VR
	bread.ReadAt(word, offset)
	offset += 4
	out.VR = string(word)

	// syngo dt
	bread.ReadAt(word, offset)
	offset += 4
	out.SyngoDT = binary.LittleEndian.Uint32(word)

	// subelement count
	bread.ReadAt(word, offset)
	offset += 4
	out.Subelements = binary.LittleEndian.Uint32(word)

	// Next is another delim. No need to read since we don't use it, but do need
	// to increment the offset.

	// bread.ReadAt(word, offset)
	// if bytes.Equal(word, SiemensDelimiter) || bytes.Equal(word, SiemensDelimiter2) {
	// }
	offset += 4

	// Iterate over the subelements
	for i := 0; i < int(out.Subelements); i++ {
		// Always 16 bytes. The first 4, second 4, and fourth 4 are equal to one
		// another, and signify the number of bytes that correspond to the data
		// in this subelement. The third 4 bytes are always a delimiter(!). So,
		// read the first 4 bytes, treat them as a 32-bit little-endian integer,
		// and increment the offset by 16.
		bread.ReadAt(word, offset)
		offset += 16

		// This subelement has dataLen bytes to be read
		dataLen := binary.LittleEndian.Uint32(word)

		dword := make([]byte, dataLen)
		bread.ReadAt(dword, offset)
		offset += int64(len(dword))

		// The data fields are always padded out to the nearest 4-byte boundary.
		// So, e.g., if we have 9 bytes of data, we will read out 3 more bytes
		// so we are padded out to 12 bytes. (Don't actually need to read out
		// the bytes - just need to advance the offset)
		if modulus := dataLen % 4; modulus != 0 {
			fixOffset := 4 - modulus
			// bread.ReadAt(make([]byte, fixOffset), offset)
			offset += int64(fixOffset)
		}

		if dataLen > 0 {
			out.SubElementData = append(out.SubElementData, string(dword))
			// noNullDword := bytes.Split(dword, []byte{0x00})
			// out.SubElementData = append(out.SubElementData, string(noNullDword[0]))
		}
	}

	return out, offset
}
