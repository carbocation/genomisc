package main

import (
	"encoding/json"
	"io"
)

type ReplacementString struct {
	Old uint32 `json:"old"`
	New uint32 `json:"new"`
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
		out[rep.Old] = rep.New
	}

	return out, err
}
