package main

import (
	"broad/ghgwas/ukbb/bulkmanifest"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	// Consume a .bulk file
	// Need a temp file
	// Optionally consume a file that creates two or more groups
	// Download and extract each bulk file in parallel, on-the-fly
	// When enough members of a class have been downloaded, extract the data and combine into a bzip2

	// For both active .tar files (cases and controls), append to a tar by using
	// this trick: https://stackoverflow.com/a/18330903/199475

	var batchSize int
	var bulkPath, tmpPath, groupPath, ukbFetch string

	flag.IntVar(&batchSize, "batch_size", 1000, "Number of samples to place in each batched zip file [max].")
	flag.StringVar(&bulkPath, "bulk", "", "Path to *.bulk file, as specified by UKBB.")
	flag.StringVar(&tmpPath, "tmp", os.TempDir(), "Path to a temporary directory where dicom and zip files can be stored.")
	flag.StringVar(&groupPath, "group", "", "Optional. File that assigns each sample to a group.")
	flag.StringVar(&ukbFetch, "ukbfetch", "ukbfetch", "Path to the ukbfetch utility (if not already in your PATH as ukbfetch).")

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
		dicoms, err := bulkmanifest.ProcessCardiacMRIZip(path+file.Name(), nil)
		if err != nil {
			log.Fatalln(err)
		}

		for _, dcm := range dicoms {
			fmt.Printf("%+v\n", dcm)
		}
	}
}
