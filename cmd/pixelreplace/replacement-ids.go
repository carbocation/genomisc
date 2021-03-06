package main

import (
	"encoding/json"
	"io"
	"strconv"
)

// JSON only permits strings as keys
type ReplacementString struct {
	Old string `json:"old"`
	New string `json:"new"`
}

type ReplacementMap map[uint32]uint32

func (rm ReplacementMap) Replace(old uint32) uint32 {
	if v, exists := rm[old]; exists {
		return v
	}

	return old
}

func ParseReplacementFile(input io.Reader) (ReplacementMap, error) {
	z := struct {
		Replacements []ReplacementString
	}{}

	err := json.NewDecoder(input).Decode(&z)

	// Convert string numbers to uint32 as expected by the label utilities
	out := make(ReplacementMap)
	for _, rep := range z.Replacements {
		k, err := strconv.Atoi(rep.Old)
		if err != nil {
			return nil, err
		}

		v, err := strconv.Atoi(rep.New)
		if err != nil {
			return nil, err
		}

		out[uint32(k)] = uint32(v)
	}

	return out, err
}
