package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func main() {
	// Consume a .bulk file
	// Download all data with ukbfetch

	var bulkPath, ukbKey, ukbFetch string
	var concurrency int

	flag.StringVar(&bulkPath, "bulk", "", "Path to *.bulk file, as specified by UKBB.")
	flag.StringVar(&ukbFetch, "ukbfetch", "ukbfetch", "Path to the ukbfetch utility (if not already in your PATH as ukbfetch).")
	flag.StringVar(&ukbKey, "ukbkey", ".ukbkey", "Path to the .ukbkey file with the app ID and special key.")
	flag.IntVar(&concurrency, "concurrency", 20, "Number of simultaneous connections to UK Biobank servers.")

	flag.Parse()

	log.Println("Note: This tool only checks for pre-existing files in the order specified by the bulk file.")

	if bulkPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	f, err := os.Open(bulkPath)
	if err != nil {
		log.Fatalln(err)
	}

	c := csv.NewReader(f)
	c.Comma = ' '

	entries, err := c.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}

	// Note: The UK Biobank updated their rules to permit only 10 simultaneous
	// downloads per application in 3/2019.
	log.Println("Using up to", concurrency, "simultaneous downloads")

	// Make it 1-based
	concurrency = concurrency - 1

	sem := make(chan bool, concurrency)

	finishedCheckingExisting := false
	for i, row := range entries {
		exists := false
		zipFile := fmt.Sprintf("%s_%s", row[0], row[1])

		if !finishedCheckingExisting {
			// Since statting on a GCSFuse filesystem is slow, we assume sorted
			// order. If that is true, then once we stop finding files we have
			// already downloaded, we can stop checking.
			for _, suffix := range []string{"zip", "cram", "cram.crai", "xml"} {
				downloadFileName := fmt.Sprintf("%s_%s.%s", row[0], row[1], suffix)

				// If we already downloaded this file, skip it
				if _, err := os.Stat(downloadFileName); !os.IsNotExist(err) {
					log.Println(i, len(entries), "Already downloaded", downloadFileName)
					exists = true
					break
				}
			}
		}

		if exists {
			continue
		}

		finishedCheckingExisting = true

		log.Println(i, len(entries), "Downloading", zipFile)

		sem <- true
		go func(row []string) {
			defer func() { <-sem }()

			if out, err := exec.Command(ukbFetch, fmt.Sprintf("-a%s", ukbKey), fmt.Sprintf("-e%s", row[0]), fmt.Sprintf("-d%s", row[1])).CombinedOutput(); err != nil {
				log.Println(fmt.Errorf("Output: %s | Error: %s | Row %v", string(out), err.Error(), row))
				log.Println("Skipping")
				time.Sleep(2 * time.Second)
			}
		}(append([]string{}, row...))

	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}
