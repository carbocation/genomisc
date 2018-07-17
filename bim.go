package genomisc

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type BIM struct {
	path    string
	file    *os.File
	scanner *bufio.Scanner
	err     error
}

func OpenBIM(path string) (*BIM, error) {
	bim := &BIM{
		path: path,
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	bim.file = file
	bim.scanner = bufio.NewScanner(file)

	return bim, nil
}

func (b *BIM) Close() error {
	return b.file.Close()
}

func (b *BIM) Err() error {
	if b.err != nil {
		return b.err
	}

	return b.scanner.Err()
}

func (b *BIM) Read() *BIMRow {
	b.scanner.Scan()
	if b.scanner.Err() != nil {
		return nil
	}

	data := b.scanner.Text()
	cols := strings.Fields(data)

	if len(cols) < Allele2+1 {
		return nil
	}

	row := &BIMRow{
		Chromosome: cols[Chromosome],
		VariantID:  cols[VariantID],
		Allele1:    cols[Allele1],
		Allele2:    cols[Allele2],
	}

	coord64, err := strconv.ParseUint(cols[Coordinate], 10, 32)
	if err != nil {
		b.err = err
		return nil
	}
	row.Coordinate = uint32(coord64)

	return row
}
