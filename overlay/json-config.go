package overlay

import (
	"encoding/json"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/carbocation/pfx"
)

type JSONConfig struct {
	ConfigPath   string
	ManifestPath string   `json:"manifest"`
	Project      string   `json:"project"`
	Port         int      `json:"port"`
	Labels       LabelMap `json:"labels"`
	ImagePath    string   `json:"image_path"`
	ImageSuffix  string   `json:"image_suffix"`
	DefaultBrush string   `json:"default_brush"`

	// Determined by whether ImagePath is set
	PreParsed bool `json:"-"`
}

func ParseJSONConfigFromPath(path string) (JSONConfig, error) {
	out := JSONConfig{ConfigPath: path}

	f, err := os.Open(path)
	if err != nil {
		return out, pfx.Err(err)
	}

	err = json.NewDecoder(f).Decode(&out)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
			return out, pfx.Err(err)
		}

		return out, pfx.Err(err)
	}

	if out.ImagePath != "" {
		out.PreParsed = true
	}

	// Internally, go uses lower case for all colors, so we will too (while
	// permitting the user to use mixed case)
	for k, v := range out.Labels {
		v.Color = strings.ToLower(out.Labels[k].Color)
		out.Labels[k] = v
	}

	// Interpret ~ if present
	out.ConfigPath = expandHomeDir(out.ConfigPath)
	out.ImagePath = expandHomeDir(out.ImagePath)
	out.ManifestPath = expandHomeDir(out.ManifestPath)
	out.Project = expandHomeDir(out.Project)

	return out, pfx.Err(err)
}

// Via https://stackoverflow.com/a/17617721/199475
func expandHomeDir(path string) string {

	usr, err := user.Current()
	if err != nil {
		return path
	}

	dir := usr.HomeDir

	if path == "~" {
		// In case of "~", which won't be caught by the "else if"
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		path = filepath.Join(dir, path[2:])
	}

	return path
}
