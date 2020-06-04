package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

var SiemensDelimiter = []byte{0x4d, 00, 00, 00}
var SiemensDelimiter2 = []byte{0xcd, 00, 00, 00}

func ParseSiemensHeader(v interface{}) (map[string]SiemensChunk, error) {

	sData := make(map[string]SiemensChunk)

	// The data is a byte stream that is encoded into 4-byte words.
	// Specifically, they are encoded as little-endian 32-bit words.
	bs := v.([]uint8)
	bread := bytes.NewReader(bs)

	word := make([]byte, 4)
	var offset int64

	// Confirmed the SV10 header
	if _, err := bread.ReadAt(word, offset); err != nil {
		return nil, err
	}
	offset += int64(len(word))

	if string(word) != "SV10" {
		return nil, fmt.Errorf("Siemens data didn't start with SV10")
	}

	// Check for the 04030201 constant
	bread.ReadAt(word, offset)
	offset += int64(len(word))
	if cslc := []byte{04, 03, 02, 01}; !bytes.Equal(cslc, word) {
		return nil, fmt.Errorf("Didn't find constant %v", cslc)
	}

	bread.ReadAt(word, offset)
	offset += int64(len(word))
	nElements := binary.LittleEndian.Uint32(word)
	fmt.Println("There are", nElements, "elements in the Siemens header")

	// Read the delimiter
	bread.ReadAt(word, offset)
	offset += int64(len(word))

	// Read the first named chunk
	var sc SiemensChunk

	for offset < int64(len(bs)) {
		sc, offset = ReadChunk(bread, offset)
		sData[sc.Name] = sc
		// fmt.Printf("%d | %+v\n", offset, sc)
	}

	for _, v := range sData["FlowVenc"].SubElementData {
		fmt.Println("FlowVenc", v)
	}

	return sData, nil
}

type SiemensChunk struct {
	Name           string
	VM             uint32
	VR             string
	SyngoDT        uint32
	Subelements    uint32
	SubElementData []string
}

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

	// subelements
	bread.ReadAt(word, offset)
	offset += 4
	out.Subelements = binary.LittleEndian.Uint32(word)

	// delim?
	bread.ReadAt(word, offset)
	offset += 4
	if bytes.Equal(word, SiemensDelimiter) || bytes.Equal(word, SiemensDelimiter2) {
		// log.Println("Done with element")
	}

	for i := 0; i < int(out.Subelements); i++ {
		// 16 bytes, first 4, second 4, and fourth 4 are equal to one another,
		// and third are always a delimiter(!). So, read the first 4, but
		// increment the offset by 16.
		bread.ReadAt(word, offset)
		offset += 16

		// This subelement has dataLen bytes to be read
		dataLen := binary.LittleEndian.Uint32(word)

		dword := make([]byte, dataLen)
		bread.ReadAt(dword, offset)
		offset += int64(len(dword))

		// The data fields are always padded out to a 4-byte boundary. So, if we
		// have 9 bytes of data, we will read out 3 more bytes so we are padded
		// out to 12 bytes.
		if modulus := dataLen % 4; modulus != 0 {
			fixOffset := 4 - modulus
			bread.ReadAt(make([]byte, fixOffset), offset)
			offset += int64(fixOffset)
		}

		if dataLen > 0 {
			out.SubElementData = append(out.SubElementData, string(dword))

			// fmt.Println("Length", dataLen, "Data:", string(dword))
		}
	}

	return out, offset
}

func ReadUntilDelimiter(bs []byte, offset int64) ([]byte, int64) {

	rest, _ := ioutil.ReadAll(bytes.NewReader(bs[offset:]))

	chunks := bytes.Split(rest, SiemensDelimiter)

	return chunks[0], int64(len(SiemensDelimiter) + len(chunks[0]))
}

func ReadNBytesAtOffsetDiscardingNul(bread *bytes.Reader, nBytes, offset int64) []byte {
	word := make([]byte, nBytes)

	bread.ReadAt(word, offset)

	wp := bytes.Split(word, []byte{0x00})

	return wp[0]
}
