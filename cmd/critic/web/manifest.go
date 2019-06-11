package main

import (
	"encoding/csv"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type Manifest struct {
	Zip                   string
	Dicom                 string
	Series                string
	InstanceNumber        int
	HasOverlayFromProject bool
}

func UpdateManifest() error {
	global.m.Lock()
	defer global.m.Unlock()

	// First, look in the project directory to get updates to annotations.
	suffix := ".png"
	files, err := ioutil.ReadDir(filepath.Join(".", global.Project))
	if os.IsNotExist(err) {
		// Not a problem
	} else if err != nil {
		return err
	}
	overlaysExist := make(map[string]struct{})
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		overlaysExist[f.Name()] = struct{}{}
	}

	updatedManifest := global.manifest

	// Toggle to the latest knowledge about the manifest
	for i, v := range updatedManifest {
		_, hasOverlay := overlaysExist[v.Zip+"_"+v.Dicom+suffix]
		updatedManifest[i].HasOverlayFromProject = hasOverlay
	}

	return nil
}

// ReadManifest takes the path to a manifest file and extracts each line.
func ReadManifest(manifestPath, projectPath string) ([]Manifest, error) {
	// First, look in the project directory to see if there is any annotation.
	suffix := ".png"
	files, err := ioutil.ReadDir(filepath.Join(".", projectPath))
	if os.IsNotExist(err) {
		// Not a problem
	} else if err != nil {
		return nil, err
	}
	overlaysExist := make(map[string]struct{})
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		overlaysExist[f.Name()] = struct{}{}
	}

	f, err := os.Open(manifestPath)
	if err != nil {
		return nil, err
	}

	cr := csv.NewReader(f)
	cr.Comma = '\t'
	recs, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}

	output := make([]Manifest, 0, len(recs))

	header := struct {
		Zip            int
		Dicom          int
		Series         int
		InstanceNumber int
	}{}

	for i, cols := range recs {
		if i == 0 {
			for j, col := range cols {
				if col == "zip_file" {
					header.Zip = j
				} else if col == "dicom_file" {
					header.Dicom = j
				} else if col == "series" {
					header.Series = j
				} else if col == "instance_number" {
					header.InstanceNumber = j
				}
			}
			continue
		}

		intInstance, err := strconv.Atoi(cols[header.InstanceNumber])
		if err != nil {
			// Ignore the error
			intInstance = 0
		}

		_, hasOverlay := overlaysExist[cols[header.Zip]+"_"+cols[header.Dicom]+suffix]

		output = append(output, Manifest{
			Zip:                   cols[header.Zip],
			Dicom:                 cols[header.Dicom],
			Series:                cols[header.Series],
			InstanceNumber:        intInstance,
			HasOverlayFromProject: hasOverlay,
		})
	}

	sort.Slice(output, generateManifestSorter(output))

	return output, nil
}

// generateManifestSorter is a convenience function to help sort a list of
// manifest objects
func generateManifestSorter(output []Manifest) func(i, j int) bool {
	return func(i, j int) bool {
		// Need both conditions so that ties will proceed to the next check
		if output[i].Zip < output[j].Zip {
			return true
		} else if output[i].Zip > output[j].Zip {
			return false
		}

		if output[i].Series < output[j].Series {
			return true
		} else if output[i].Series > output[j].Series {
			return false
		}

		if output[i].InstanceNumber < output[j].InstanceNumber {
			return true
		} else if output[i].InstanceNumber > output[j].InstanceNumber {
			return false
		}

		if output[i].Dicom < output[j].Dicom {
			return true
		} else if output[i].Dicom > output[j].Dicom {
			return false
		}

		return false
	}
}
