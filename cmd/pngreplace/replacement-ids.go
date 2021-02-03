package main

import (
	"encoding/json"
	"image/color"
	"io"

	"github.com/icza/gox/imagex/colorx"
)

// JSON only permits strings as keys
type ReplacementString struct {
	Old string `json:"old"`
	New string `json:"new"`
}

type ReplacementMap map[color.Color]color.Color

func (rm ReplacementMap) Replace(old color.Color) color.Color {
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
		old, err := colorx.ParseHexColor(rep.Old)
		if err != nil {
			return nil, err
		}

		v, err := colorx.ParseHexColor(rep.New)
		if err != nil {
			return nil, err
		}

		out[old] = v
	}

	return out, err
}
