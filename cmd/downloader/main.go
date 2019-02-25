package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func main() {
	// Consume a .bulk file
	// Download all data with ukbfetch

	var bulkPath, ukbKey, ukbFetch string

	flag.StringVar(&bulkPath, "bulk", "", "Path to *.bulk file, as specified by UKBB.")
	flag.StringVar(&ukbFetch, "ukbfetch", "ukbfetch", "Path to the ukbfetch utility (if not already in your PATH as ukbfetch).")
	flag.StringVar(&ukbKey, "ukbkey", ".ukbkey", "Path to the .ukbkey file with the app ID and special key.")

	flag.Parse()

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

	concurrency := 16
	if nCPU := runtime.NumCPU(); nCPU > concurrency {
		concurrency = nCPU
	}
	sem := make(chan bool, concurrency)

	for i, row := range entries {
		zipFile := fmt.Sprintf("%s_%s.zip", row[0], row[1])

		// If we already downloaded this file, skip it
		if _, err := os.Stat(zipFile); !os.IsNotExist(err) {
			log.Println(i, len(entries), "Already downloaded", zipFile)
			continue
		}

		log.Println(i, len(entries), "Downloading", zipFile)

		sem <- true
		go func(row []string) {
			defer func() { <-sem }()

			if out, err := exec.Command(ukbFetch, fmt.Sprintf("-a%s", ukbKey), fmt.Sprintf("-e%s", row[0]), fmt.Sprintf("-d%s", row[1])).CombinedOutput(); err != nil {
				log.Println(fmt.Errorf("Output: %s | Error: %s", string(out), err.Error()))
				log.Println("Sleeping 30 seconds and retrying")
				time.Sleep(30 * time.Second)
			}
		}(append([]string{}, row...))

	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}
