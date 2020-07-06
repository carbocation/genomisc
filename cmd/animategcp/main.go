package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	DicomColumnName     = "dicom_file"
	TimepointColumnName = "trigger_time"
	SampleIDColumnName  = "sample_id"
	InstanceColumnName  = "instance"
)

func main() {
	var manifest, folder, suffix string
	var delay int
	flag.StringVar(&manifest, "manifest", "", "Path to manifest file")
	flag.StringVar(&folder, "folder", "", "Path to google storage folder that contains PNGs")
	flag.StringVar(&suffix, "suffix", ".png", "Suffix after .dcm. Typically .png for raw dicoms or .png.overlay.png for merged dicoms.")
	flag.IntVar(&delay, "delay", 1, "Milliseconds between each frame of the gif.")
	flag.Parse()

	if manifest == "" || folder == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := run(manifest, folder, suffix, delay); err != nil {
		log.Fatalln(err)
	}

	log.Println("Quitting")
}

func run(manifest, folder, suffix string, delay int) error {

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

		fmt.Printf("Fetching images for %+v", key)
		started := time.Now()
		go func() {
			errchan <- makeOneGif(pngs, outName, delay)
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
			fmt.Println("Successfully created", outName)
		}
		continue
	}

	return nil
}
