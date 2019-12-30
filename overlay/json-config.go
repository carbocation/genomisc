package overlay

import (
	"encoding/json"
	"os"
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

	return out, err
}
