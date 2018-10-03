package main

import (
	"broad/ghgwas/ukbb/cardiacmri"
	"flag"
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	// List folder contents
	// Create or open named SQLite file
	path := ""

	flag.StringVar(&path, "path", "/Users/jamesp/go/src/broad/ghgwas/ukbb/cardiacmri/testdata/", "Path to the directory where the zipfiles sit.")

	flag.Parse()

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatalln(err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".zip") {
			continue
		}
		// Read each zip, whose name is significant
		cardiacmri.ProcessCardiacMRIZip(path+file.Name(), nil)
		// cardiacmri.SurveyZipManifests(path+file.Name(), nil)
	}
}
