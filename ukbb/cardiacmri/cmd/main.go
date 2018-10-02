package main

import (
	"broad/ghgwas/ukbb/cardiacmri"
)

func main() {
	// List folder contents
	// Create or open named SQLite file
	// Read each zip, whose name is significant
	files := []string{"/Users/jamesp/go/src/broad/ghgwas/ukbb/cardiacmri/testdata/2780231_20208_2_0.zip"}

	for _, file := range files {
		cardiacmri.ProcessCardiacMRIZip(file, nil)
	}
}
