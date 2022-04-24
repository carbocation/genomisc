package main

import (
	"encoding/csv"
	"io"
	"strconv"
)

type centiMapKey struct {
	chr string
	pos int
}

type centiMapValue struct {
	value   float64
	key     centiMapKey
	prevKey centiMapKey
	nextKey centiMapKey
}

var centiSlice = make([]centiMapValue, 0)
var centiMap = make(map[centiMapKey]centiMapValue)

func loadMap(f io.Reader) error {
	r := csv.NewReader(f)
	r.Comma = []rune(MapDelim)[0]
	r.Comment = '#'

	lines, err := r.ReadAll()
	if err != nil {
		return err
	}

	lastKey := centiMapKey{}
	for i, v := range lines {
		chr := v[MapCHRColumn]
		bp, err := strconv.Atoi(v[MapBPColumn])
		if err != nil {
			return err
		}
		cm, err := strconv.ParseFloat(v[MapCMColumn], 64)
		if err != nil {
			return err
		}

		key := centiMapKey{chr, bp}
		value := centiMapValue{cm, key, lastKey, centiMapKey{}}
		centiMap[key] = value
		centiSlice = append(centiSlice, value)
		if i != 0 {
			// Update the prior entry's next key to point to this entry
			lastValue := centiMap[lastKey]
			lastValue.nextKey = key
			centiMap[lastKey] = lastValue
			centiSlice[i-1] = lastValue
		}
		lastKey = key
	}

	// Special cases: lastKey=0,0 (first entry) and nextKey=0,0 (last entry)

	return nil
}

func lookupMapPosition(position, precedingMapSliceKey int) int {
	// If we are already at the end of the map, stop
	if precedingMapSliceKey == len(centiSlice)-1 {
		return precedingMapSliceKey
	}

	// If the lookup position is before the current map position, stay at the
	// current map position
	if position < centiSlice[precedingMapSliceKey].key.pos {
		return precedingMapSliceKey
	}

	// The lookup position is to the right of the current map position. We could
	// still be in the correct place, but we need to travel forward to see if we
	// need to advance—and, if so, by how much.
	advanceBy := 0
	for i := precedingMapSliceKey + 1; i < len(centiSlice); i++ {
		if position > centiSlice[i].key.pos {
			// If the position is *still* to the right of our new position, then
			// we need to keep advancing.
			advanceBy++
		} else {
			// The position is finally to the left of our new position, so we
			// don't want to advance to here and we can stop searching.
			break
		}
	}

	return advanceBy + precedingMapSliceKey
}
