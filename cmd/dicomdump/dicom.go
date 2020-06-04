package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/carbocation/pfx"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/dicomtag"
	"github.com/suyashkumar/dicom/element"
)

func IterateOverFolder(path string) error {

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return pfx.Err(err)
	}

	for _, file := range files {

		func(file os.FileInfo) {

			if !strings.HasSuffix(file.Name(), ".zip") {
				return
			}

			if err := ProcessZip(path + file.Name()); err != nil {
				log.Println(err)
				return
			}
		}(file)
	}

	return nil
}

func ProcessZip(zipPath string) (err error) {

	fmt.Println(strings.Repeat("=", 30))
	fmt.Println(zipPath)
	fmt.Println(strings.Repeat("=", 30))

	rc, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	for _, v := range rc.File {
		// Looking only at the dicoms
		if strings.HasPrefix(v.Name, "manifest") {
			continue
		}

		fmt.Println(strings.Repeat("-", 30))
		fmt.Println(v.Name)
		fmt.Println(strings.Repeat("-", 30))

		unzippedFile, err := v.Open()
		if err != nil {
			return err
		}
		if err := ProcessDicom(unzippedFile); err != nil {
			log.Println("Ignoring error and continuing:", err.Error())
			continue
		}
	}

	return nil
}

// Takes in a dicom file (in bytes), emit meta-information
func ProcessDicom(dicomReader io.Reader) error {
	dcm, err := ioutil.ReadAll(dicomReader)
	if err != nil {
		return err
	}

	p, err := dicom.NewParserFromBytes(dcm, nil)
	if err != nil {
		return err
	}

	parsedData, err := p.Parse(dicom.ParseOptions{
		DropPixelData: true,
	})
	if parsedData == nil || err != nil {
		return fmt.Errorf("Error reading dicom: %v", err)
	}

	for _, elem := range parsedData.Elements {
		tagName, _ := dicomtag.Find(elem.Tag)

		if tagName.Name == "" {
			tagName.Name = "____"
		}

		if elem.Tag.Compare(dicomtag.Tag{Group: 0x0029, Element: 0x1010}) == 0 {
			if err := ParseSiemensHeader(elem); err != nil {
				return err
			}

			continue
		} else {
			continue
		}

		if elem.Tag.Compare(dicomtag.Tag{Group: 0x0029, Element: 0x1020}) == 0 {
			fmt.Println(elem.Tag, tagName.Name, "~~skipping value~~")
			continue
		}

		fmt.Println(elem.Tag, tagName.Name, elem.Value)
	}

	return nil
}

var SiemensDelimiter = []byte{0x4d, 00, 00, 00}
var SiemensDelimiter2 = []byte{0xcd, 00, 00, 00}

func ParseSiemensHeader(elem *element.Element) error {
	for _, v := range elem.Value {

		sData := make(map[string]SiemensChunk)

		// The data is a byte stream that is encoded into 4-byte words.
		// Specifically, they are encoded as little-endian 32-bit words.
		bs := v.([]uint8)
		bread := bytes.NewReader(bs)

		word := make([]byte, 4)
		var offset int64

		// Confirmed the SV10 header
		if _, err := bread.ReadAt(word, offset); err != nil {
			return err
		}
		offset += int64(len(word))

		if string(word) != "SV10" {
			return fmt.Errorf("Siemens data didn't start with SV10")
		}

		// Check for the 04030201 constant
		bread.ReadAt(word, offset)
		offset += int64(len(word))
		if cslc := []byte{04, 03, 02, 01}; !bytes.Equal(cslc, word) {
			return fmt.Errorf("Didn't find constant %v", cslc)
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

		// sc, offset = ReadChunk(bread, offset)
		// fmt.Printf("%d | %+v\n", offset, sc)

		// if offset < int64(len(bs)) {
		// 	sc, offset = ReadChunk(bread, offset)
		// 	fmt.Printf("%d | %+v\n", offset, sc)
		// }
		// if offset < int64(len(bs)) {
		// 	sc, offset = ReadChunk(bread, offset)
		// 	fmt.Printf("%d | %+v\n", offset, sc)
		// }
		// if offset < int64(len(bs)) {
		// 	sc, offset = ReadChunk(bread, offset)
		// 	fmt.Printf("%d | %+v\n", offset, sc)
		// }

		// // VM
		// info := ReadNBytesAtOffsetDiscardingNul(bread, 4, offset)
		// offset += 4
		// fmt.Println(info)

		// // VR
		// info = ReadNBytesAtOffsetDiscardingNul(bread, 4, offset)
		// offset += 4
		// fmt.Println(string(info))

		// // syngo dt
		// info = ReadNBytesAtOffsetDiscardingNul(bread, 4, offset)
		// offset += 4
		// fmt.Println(info)

		// // subelements
		// info = ReadNBytesAtOffsetDiscardingNul(bread, 4, offset)
		// offset += 4
		// fmt.Println(info)

		// cv, newOff := ReadUntilDelimiter(bs, offset)
		// offset += newOff
		// fmt.Println(cv)

		// cv, newOff = ReadUntilDelimiter(bs, offset)
		// offset += newOff
		// fmt.Println(cv)

		// ??
		// info = ReadNBytesAtOffsetDiscardingNul(bread, 4, offset)
		// offset += 4
		// fmt.Println(info)

		// bread.ReadAt(word, offset)
		// offset += int64(len(word))
		// fmt.Println(word)

		// fmt.Println("---")

		// binary.LittleEndian.PutUint32(word, 0x4d)
		// fmt.Println([]byte{0x4d, 00, 00, 00})
		// fmt.Println(word)

		// // How many delimiters remain?
		// rest, _ := ioutil.ReadAll(bytes.NewReader(bs[offset:]))
		// fmt.Println(len(rest))
		// fmt.Println(len(rest) / 64)
		// fmt.Println(bytes.Count(rest, []byte{0x4d}))
		// fmt.Println(bytes.Count(rest, SiemensDelimiter))

		// chunks := bytes.Split(rest, SiemensDelimiter)
		// for i, chunk := range chunks {
		// 	if len(chunk) < 10 {
		// 		fmt.Println(chunk)
		// 		continue
		// 	}

		// 	fmt.Printf("%02d| %s\n", i, string(chunk))
		// }

		// bitreader.NewReader(bytes.NewReader(bs[offset:]))
		// fmt.Println([]byte(uint(0x4d0000))

		// fmt.Println(string(bs))
		// rest, _ := ioutil.ReadAll(bread)
		// fmt.Println(string(rest))
		// fmt.Println(string(bytes.ReplaceAll(rest, []byte{00, 00}, []byte{00})))

		// r := bitreader.NewReader(bytes.NewReader(bs[8:]))
		// bitVal, _ := r.Read1()
		// fmt.Println(bs[8])
		// fmt.Println(0x11223344)
		// z := make([]byte, 4)
		// binary.BigEndian.PutUint32(z, 0x11223344)
		// fmt.Println(z)
		// fmt.Println(int(binary.BigEndian.Uint32(z)))

		// claim := []byte{11, 22, 33, 44}
		// fmt.Println(claim)
		// fmt.Println(binary.BigEndian.Uint32(claim))
		// fmt.Println(bitVal)

		// fmt.Print(fmt.Sprintf("%s\n", string(bytes.Runes(bs))))

	}

	return nil
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
