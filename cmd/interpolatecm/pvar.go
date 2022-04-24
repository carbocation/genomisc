package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func processPVAR(f io.Reader, w io.Writer) error {
	var header []string
	sawHeader := false
	pvarCMColumn := -1

	currentMapSliceKey := 0

	// Read each line from the reader f and print it to the writer w
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "##") {
			fmt.Fprintln(w, line)
			continue
		}

		fields := strings.Split(line, "\t")

		// Setup the header
		if !sawHeader {
			sawHeader = true
			header = fields

			for j, v := range header {
				if v == "CM" {
					pvarCMColumn = j
					break
				}
			}
			if pvarCMColumn < 0 {
				return fmt.Errorf("CM column not found in header. Saw: %v", header)
			}

			fmt.Fprintln(w, line)
			continue
		}

		pos, err := strconv.Atoi(fields[PVARPosColumn])
		if err != nil {
			return err
		}

		currentMapSliceKey = lookupMapPosition(pos, currentMapSliceKey)
		leftMap := centiSlice[currentMapSliceKey]
		var rightMap centiMapValue
		if currentMapSliceKey == len(centiSlice)-1 {
			rightMap = centiMapValue{}
		} else {
			rightMap = centiSlice[currentMapSliceKey+1]
		}

		value := interpolate(pos, leftMap, rightMap)
		fields[pvarCMColumn] = fmt.Sprintf("%.6f", value)

		fmt.Fprintln(w, strings.Join(fields, "\t"))
	}

	return scanner.Err()
}

// Y3 = Y1 + (Y2 - Y1) / (X2 - X1) * (X3 - X1)
func interpolate(position int, behind, ahead centiMapValue) float64 {
	if ahead.key.pos == 0 {
		// If ahead is beyond the last entry, then we can't interpolate
		return behind.value
	}

	if behind.key.pos == 0 {
		// If behind is before the first entry, then we can't interpolate
		return ahead.value
	}

	return behind.value + (float64(ahead.value-behind.value) / float64(ahead.key.pos-behind.key.pos) * float64(position-behind.key.pos))
}
