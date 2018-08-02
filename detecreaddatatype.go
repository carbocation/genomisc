package genomisc

import (
	"compress/bzip2"
	"compress/gzip"
	"compress/zlib"
	"io"
	"os"

	"github.com/krolaw/zipstream"
	"github.com/xi2/xz"
)

type DataType byte

const (
	DataTypeInvalid DataType = iota
	DataTypeNoCompression
	DataTypeGzip
	DataTypeZip
	DataTypeXZ
	DataTypeZ
	DataTypeBZip2
)

var byteCodeSigs = map[DataType][]byte{
	DataTypeGzip:  {0x1f, 0x8b, 0x08},
	DataTypeZip:   {0x50, 0x4b, 0x03, 0x04},
	DataTypeXZ:    {0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00},
	DataTypeZ:     {0x1f, 0x9d},
	DataTypeBZip2: {0x42, 0x5a, 0x68},
}

// DetectDataType attempts to detect the data type of a stream by checking
// against a set of known data types.  Byte code signatures from
// https://stackoverflow.com/a/19127748/199475
func DetectDataType(r io.Reader) (DataType, error) {
	buff := make([]byte, 6)
	if _, err := r.Read(buff); err != nil {
		return DataTypeInvalid, err
	}

	// Match known signatures
Outer:
	for dt, sig := range byteCodeSigs {
		for position := range sig {
			if buff[position] != sig[position] {
				continue Outer
			}
		}
		return dt, nil
	}

	return DataTypeNoCompression, nil
}

// If the file contains
func MaybeDecompressReadCloserFromFile(f *os.File) (io.ReadCloser, error) {
	dt, err := DetectDataType(f)
	if err != nil {
		return nil, err
	}
	// Reset your original reader
	defer f.Seek(0, 0)

	switch dt {
	case DataTypeGzip:
		return gzip.NewReader(f)
	case DataTypeZip:
		return &readCloserFaker{zipstream.NewReader(f)}, nil
	case DataTypeBZip2:
		return &readCloserFaker{bzip2.NewReader(f)}, nil
	case DataTypeXZ:
		reader, err := xz.NewReader(f, 0)
		if err != nil {
			return nil, err
		}
		return &readCloserFaker{reader}, nil
	case DataTypeZ:
		return zlib.NewReader(f)
	}

	// No data type detected. For now, we assume this is uncompressed.
	return f, nil
}

// readCloserFaker "upgrades" readers that don't need to be closed
type readCloserFaker struct {
	io.Reader
}

func (c *readCloserFaker) Close() error {
	return nil
}
