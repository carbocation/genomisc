package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

const (
	SampleIDColumnName = "sample_id"
	InstanceColumnName = "instance"
)

var (
	TimepointColumnName = "trigger_time"
	DicomColumnName     = "dicom_file"
)

// Safe for concurrent use by multiple goroutines so we'll make this a global
var client *storage.Client

func main() {
	defer func() { log.Println("Quitting") }()

	var manifest, folder, suffix, sampleList string
	var delay int
	var grid bool
	flag.StringVar(&manifest, "manifest", "", "Path to manifest file")
	flag.StringVar(&folder, "folder", "", "Path to google storage folder that contains PNGs")
	flag.StringVar(&suffix, "suffix", ".png", "Suffix after .dcm. Typically .png for raw dicoms or .png.overlay.png for merged dicoms.")
	flag.StringVar(&sampleList, "samples", "", "To run in batch mode, provide a file that contains one sample (format: sampleid_instance) per line.")
	flag.StringVar(&DicomColumnName, "dicom_column_name", "dicom_file", "Name of the column in the manifest with the dicom.")
	flag.StringVar(&TimepointColumnName, "sequence_column_name", "trigger_time", "Name of the column that indicates the order of the images with an increasing number.")
	flag.IntVar(&delay, "delay", 2, "Milliseconds between each frame of the gif.")
	flag.BoolVar(&grid, "grid", true, "If multiple series are included, display as grid? (If false, will display sequentially)")
	flag.Parse()

	if manifest == "" || folder == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	folder = strings.TrimSuffix(folder, "/")

	// Initialize the Google Storage client, but only if our folder indicates
	// that we are pointing to a Google Storage path.
	if strings.HasPrefix(folder, "gs://") {
		var err error
		client, err = storage.NewClient(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}

	if sampleList != "" {
		if err := runBatch(sampleList, manifest, folder, suffix, delay, grid); err != nil {
			log.Fatalln(err)
		}

		return
	}

	if err := run(manifest, folder, suffix, delay, grid); err != nil {
		log.Fatalln(err)
	}
}

func runBatch(sampleList, manifest, folder, suffix string, delay int, grid bool) error {
	fmt.Println("Batch mode animated Gif maker")

	man, err := parseManifest(manifest)
	if err != nil {
		return err
	}

	f, err := os.Open(sampleList)
	if err != nil {
		return err
	}

	r := csv.NewReader(f)
	samples, err := r.ReadAll()
	if err != nil {
		return err
	}

	for _, sampleRow := range samples {
		if len(sampleRow) != 1 {
			continue
		}
		sample := sampleRow[0]

		input := strings.Split(sample, "_")
		if len(input) != 2 {
			fmt.Println("Expected sampleid separated from instance by an underscore (_)")
			continue
		}

		key := manifestKey{SampleID: input[0], Instance: input[1]}

		entries, exists := man[key]
		if !exists {
			fmt.Println(key, "not found in the manifest")
			continue
		}

		sort.Slice(entries, func(i, j int) bool { return entries[i].timepoint < entries[j].timepoint })

		pngs := make([]string, 0, len(entries))
		for _, entry := range entries {
			pngs = append(pngs, folder+"/"+entry.dicom+suffix)
		}

		outName := key.SampleID + "_" + key.Instance + ".gif"

		errchan := make(chan error)

		fmt.Printf("Fetching images for %+v %s_%s", key, key.SampleID, key.Instance)
		go func() {
			if grid {
				errchan <- makeOneGrid(pngs, outName, delay)
			} else {
				errchan <- makeOneGif(pngs, outName, delay)
			}
		}()

	WaitLoop:
		for {
			select {
			case err = <-errchan:
				fmt.Printf("\n")
				if err != nil {
					fmt.Println("Error making gif:", err.Error())
				}
				break WaitLoop
			}
		}

		if err == nil {
			fmt.Println("Successfully created", outName, "for", fmt.Sprintf("%s_%s", key.SampleID, key.Instance))
		}
	}

	return nil
}

func run(manifest, folder, suffix string, delay int, grid bool) error {

	fmt.Println("Animated Gif maker")

	man, err := parseManifest(manifest)
	if err != nil {
		return err
	}

	rdr := bufio.NewReader(os.Stdin)
	fmt.Println("We are aware of", len(man), "samples in the manifest")
	fmt.Println("An example of the sampleid_instance format is: 1234567_2")
	fmt.Println("Enter 'rand' for a random entry")
	fmt.Println("Enter 'q' to quit")
	fmt.Println("---------------------")

	tick := time.NewTicker(1 * time.Second)

	for {
		fmt.Print("[sampleid_instance]> ")
		text, err := rdr.ReadString('\n')
		if err != nil {
			return err
		}
		text = strings.ReplaceAll(text, "\n", "")

		if text == "q" {
			fmt.Println("quitting")
			break
		}

		var key manifestKey

		if text == "rand" {

			for k := range man {
				key = k
				break
			}
		} else {

			input := strings.Split(text, "_")
			if len(input) != 2 {
				fmt.Println("Expected sampleid separated from instance by an underscore (_)")
				continue
			}

			key = manifestKey{SampleID: input[0], Instance: input[1]}
		}

		entries, exists := man[key]
		if !exists {
			fmt.Println(key, "not found in the manifest")
			continue
		}

		sort.Slice(entries, func(i, j int) bool { return entries[i].timepoint < entries[j].timepoint })

		pngs := make([]string, 0, len(entries))
		for _, entry := range entries {
			pngs = append(pngs, folder+"/"+entry.dicom+suffix)
		}

		outName := key.SampleID + "_" + key.Instance + ".gif"

		errchan := make(chan error)

		fmt.Printf("Fetching images for %+v %s_%s", key, key.SampleID, key.Instance)
		started := time.Now()
		go func() {
			if grid {
				errchan <- makeOneGrid(pngs, outName, delay)
			} else {
				errchan <- makeOneGif(pngs, outName, delay)
			}
		}()

	WaitLoop:
		for {
			select {
			case err = <-errchan:
				fmt.Printf("\n")
				if err != nil {
					fmt.Println("Error making gif:", err.Error())
				}
				break WaitLoop
			case current := <-tick.C:
				fmt.Printf("\rFetching images for %+v (%s)", key, current.Sub(started))
			}
		}

		if err == nil {
			fmt.Println("Successfully created", outName, "for", fmt.Sprintf("%s_%s", key.SampleID, key.Instance))
		}
		continue
	}

	return nil
}
