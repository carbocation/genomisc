package main

import (
	"encoding/json"
	"io"
)

// Labels are turned into buttons in the UI, allowing each image to be labeled
// with one class
type Label struct {
	Value       string `json:"value"`
	DisplayName string `json:"name"`
}

func ParseLabelFile(input io.Reader) ([]Label, error) {
	z := struct {
		Labels []Label
	}{}

	err := json.NewDecoder(input).Decode(&z)

	return z.Labels, err
}
