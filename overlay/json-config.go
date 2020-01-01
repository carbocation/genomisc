package overlay

import (
	"encoding/json"
	"os"
	"strings"
)

type JSONConfig struct {
	ConfigPath   string
	ManifestPath string   `json:"manifest"`
	Project      string   `json:"project"`
	Port         int      `json:"port"`
	Labels       LabelMap `json:"labels"`
	ImagePath    string   `json:"image_path"`
	ImageSuffix  string   `json:"image_suffix"`

	// Determined by whether ImagePath is set
	PreParsed bool `json:"-"`
}

func ParseJSONConfigFromPath(path string) (JSONConfig, error) {
	out := JSONConfig{ConfigPath: path}

	f, err := os.Open(path)
	if err != nil {
		return out, err
	}

	err = json.NewDecoder(f).Decode(&out)

	if out.ImagePath != "" {
		out.PreParsed = true
	}

	// Internally, go uses lower case for all colors, so we will too (while
	// permitting the user to use mixed case)
	for k, v := range out.Labels {
		v.Color = strings.ToLower(out.Labels[k].Color)
		out.Labels[k] = v
	}

	return out, err
}
