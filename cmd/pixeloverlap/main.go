package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	_ "github.com/carbocation/genomisc/compileinfoprint"
	"github.com/carbocation/genomisc/overlay"
)

func init() {
	flag.Usage = func() {
		flag.PrintDefaults()

		log.Println("Example JSONConfig file layout:")
		bts, err := json.MarshalIndent(overlay.JSONConfig{Labels: overlay.LabelMap{"Background": overlay.Label{Color: "", ID: 0}}}, "", "  ")
		if err == nil {
			log.Println(string(bts))
		}
	}
}

// Safe for concurrent use by multiple goroutines
var client *storage.Client

func main() {
	fmt.Fprintf(os.Stderr, "%q\n", os.Args)

	var threshold int
	var path1, path2, jsonConfig, manifest, suffix string

	flag.StringVar(&path1, "path1", "", "Path to folder with encoded overlay images (base/truth)")
	flag.StringVar(&path2, "path2", "", "Path to folder with encoded overlay images (comparator)")
	flag.StringVar(&jsonConfig, "config", "", "JSONConfig file from the github.com/carbocation/genomisc/overlay package")
	flag.StringVar(&manifest, "manifest", "", "(Optional) Path to manifest. If provided, will only look at files in the manifest rather than listing the entire directory's contents.")
	flag.StringVar(&suffix, "suffix", ".png.mask.png", "(Optional) Suffix after .dcm. Only used if using the -manifest option.")
	flag.Parse()

	if path1 == "" || path2 == "" || jsonConfig == "" {
		flag.Usage()
		os.Exit(1)
	}

	config, err := overlay.ParseJSONConfigFromPath(jsonConfig)
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}

	// Initialize the Google Storage client only if we're pointing to Google
	// Storage paths.
	if strings.HasPrefix(path1, "gs://") || strings.HasPrefix(path2, "gs://") {
		var err error
		client, err = storage.NewClient(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}

	fmt.Println(strings.Join([]string{"dicom", "LabelID", "Label", "Agree", "Only1", "Only2", "Kappa", "Dice", "Jaccard", "CountAgreement"}, "\t"))

	if manifest != "" {

		if err := runSlice(config, path1, path2, suffix, manifest, threshold); err != nil {
			log.Fatalln(err)
		}

		return
	}

	if err := runFolder(config, path1, path2, threshold); err != nil {
		log.Fatalln(err)
	}

}

func runSlice(config overlay.JSONConfig, path1, path2, suffix, manifest string, threshold int) error {

	dicoms, err := getDicomSlice(manifest)
	if err != nil {
		return err
	}

	concurrency := 4 * runtime.NumCPU()
	sem := make(chan bool, concurrency)

	// Process every image in the manifest
	for i, file := range dicoms {
		sem <- true
		go func(file string) {

			// The main purpose of this loop is to handle a specific filesystem
			// error (input/output error) that largely happens with GCSFuse, and
			// retry a few times before giving up.
			for loadAttempts, maxLoadAttempts := 1, 10; loadAttempts <= maxLoadAttempts; loadAttempts++ {

				err := processOneImage(path1+"/"+file+suffix, path2+"/"+file+suffix, file, config, threshold)

				if err != nil && loadAttempts == maxLoadAttempts {

					// We've exhausted our retries. Fail hard.
					log.Fatalln(err)

				} else if err != nil && strings.Contains(err.Error(), "input/output error") {

					// If it's an i/o error, we can retry
					log.Println("Sleeping 5s to recover from", err.Error(), ". Attempt #", loadAttempts)
					time.Sleep(5 * time.Second)
					continue

				} else if err != nil {

					// If it's an error that is not an i/o error, don't retry
					log.Println(err)
					break

				}

				// If no error, don't retry
				break
			}

			<-sem
		}(file)

		if (i+1)%1000 == 0 {
			log.Printf("Processed %d images\n", i+1)
		}
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return nil
}

func runFolder(config overlay.JSONConfig, path1, path2 string, threshold int) error {

	files, err := scanFolder(path1)
	if err != nil {
		return err
	}

	// concurrency := 4 * runtime.NumCPU()
	concurrency := 1
	sem := make(chan bool, concurrency)

	// Process every image in the folder
	for i, file := range files {
		if file.IsDir() {
			continue
		}

		sem <- true
		go func(file string) {
			if err := processOneImage(path1+"/"+file, path2+"/"+file, file, config, threshold); err != nil {
				log.Println(err)
			}
			<-sem
		}(file.Name())

		if (i+1)%1000 == 0 {
			log.Printf("Processed %d images\n", i+1)
		}
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return nil
}

func processOneImage(filePath1, filePath2, filename string, config overlay.JSONConfig, threshold int) error {
	// Heuristic: get dicom name
	dicom := strings.ReplaceAll(filename, ".png.mask.png", "")
	dicom = strings.ReplaceAll(dicom, ".mask.png", "")

	// Open the files
	overlay1, err := overlay.OpenImageFromLocalFileOrGoogleStorage(filePath1, client)
	if err != nil {
		return err
	}

	overlay2, err := overlay.OpenImageFromLocalFileOrGoogleStorage(filePath2, client)
	if err != nil {
		return err
	}

	// Make sure they have the same dimensions
	r1 := overlay1.Bounds()
	r2 := overlay2.Bounds()
	if r1 != r2 {
		return fmt.Errorf("Bounds differ between image 1 (%v) and image 2 (%v)", r1, r2)
	}

	// map[ID assigned by 1 & ID assigned by 2] => count
	confusionMatrix := make(map[assignment]int64)

	for y := 0; y < r1.Bounds().Max.Y; y++ {
		for x := 0; x < r1.Bounds().Max.X; x++ {

			col1, err := overlay.LabeledPixelToID(overlay1.At(x, y))
			if err != nil {
				return err
			}

			col2, err := overlay.LabeledPixelToID(overlay2.At(x, y))
			if err != nil {
				return err
			}

			key := assignment{
				ID1: col1,
				ID2: col2,
			}

			confusionMatrix[key]++
		}
	}

	labelEvaluations := make(map[overlay.Label]evalLabel)

	for _, label1 := range config.Labels.Sorted() {
		v1 := labelEvaluations[label1]

		for _, label2 := range config.Labels.Sorted() {

			cm := confusionMatrix[assignment{uint32(label1.ID), uint32(label2.ID)}]

			// The class labeled by person 1 always gets incremented.
			v1.Total += cm

			// If the class labeled by person 1 is the same as that by person 2,
			// then increment agreement and don't double count.
			if label2.ID == label1.ID {
				v1.Agreed += cm
				continue
			}

			// If the classes disagree, then we note that only person 1
			// annotated these pixels as this label
			v1.Only1 += cm

			// And we increment the total count and the only-2 count for the 2nd
			// person's label.
			v2 := labelEvaluations[label2]
			v2.Only2 += cm
			v2.Total += cm
			labelEvaluations[label2] = v2

			// fmt.Printf("%s\t%s\t%s\t%d\n", filename, label1.Label, label2.Label, cm)
		}
		labelEvaluations[label1] = v1
	}

	total := int64(r1.Bounds().Max.Y * r1.Bounds().Max.X)
	for _, label := range config.Labels.Sorted() {
		v := labelEvaluations[label]

		fmt.Printf("%s\t%d\t%s\t%d\t%d\t%d\t%g\t%g\t%g\t%g\n", dicom, label.ID, label.Label, v.Agreed, v.Only1, v.Only2, v.Kappa(total), v.Dice(), v.Jaccard(), v.CountAgreement())
	}

	return nil
}
