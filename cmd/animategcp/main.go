package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/carbocation/pfx"
	"github.com/ericpauley/go-quantize/quantize"
)

const (
	DicomColumnName     = "dicom_file"
	TimepointColumnName = "trigger_time"
	SampleIDColumnName  = "sample_id"
	InstanceColumnName  = "instance"
)

func main() {
	var manifest, folder, suffix string
	flag.StringVar(&manifest, "manifest", "", "Path to manifest file")
	flag.StringVar(&folder, "folder", "", "Path to google storage folder that contains PNGs")
	flag.StringVar(&suffix, "suffix", ".png", "Suffix after .dcm. Typically .png for raw dicoms or .png.overlay.png for merged dicoms.")
	flag.Parse()

	if manifest == "" || folder == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := run(manifest, folder, suffix); err != nil {
		log.Fatalln(err)
	}

	log.Println("Quitting")
}

type manifestEntry struct {
	dicom     string
	timepoint float64
}

type manifestKey struct {
	SampleID string
	Instance string
}

func parseManifest(manifestPath string) (map[manifestKey][]manifestEntry, error) {
	man, err := os.Open(manifestPath)
	if err != nil {
		return nil, err
	}
	cr := csv.NewReader(man)
	cr.Comma = '\t'
	manifest, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}

	out := make(map[manifestKey][]manifestEntry)
	var dicom, timepoint, sampleid, instance int
	for i, cols := range manifest {
		if i == 0 {
			for k, col := range cols {
				switch col {
				case DicomColumnName:
					dicom = k
				case TimepointColumnName:
					timepoint = k
				case SampleIDColumnName:
					sampleid = k
				case InstanceColumnName:
					instance = k
				}
			}
			continue
		}

		tp, err := strconv.ParseFloat(cols[timepoint], 64)
		if err != nil {
			return nil, err
		}

		key := manifestKey{SampleID: cols[sampleid], Instance: cols[instance]}
		value := manifestEntry{dicom: cols[dicom], timepoint: tp}

		entry := out[key]
		entry = append(entry, value)
		out[key] = entry
	}

	return out, nil
}

func run(manifest, folder, suffix string) error {
	man, err := parseManifest(manifest)
	if err != nil {
		return err
	}

	// pngs := []string{
	// 	"gs://ml4cvd/jamesp/annotation/flow/v20200702/apply/output/merged_pngs/1.3.12.2.1107.5.2.18.141243.2017042612302443114006926.dcm.png.overlay.png",
	// }

	rdr := bufio.NewReader(os.Stdin)
	fmt.Println("Animated Gif maker")
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

		fmt.Printf("Fetching images for %v", key)
		started := time.Now()
		go func() {
			errchan <- makeOneGif(pngs, outName)
		}()

	WaitLoop:
		for {
			select {
			case err := <-errchan:
				fmt.Printf("\n")
				if err != nil {
					fmt.Println("Error making gif:", err.Error())
				}
				break WaitLoop
			case current := <-tick.C:
				fmt.Printf("\rFetching images for %+v (%s)", key, current.Sub(started))
			}
		}

		fmt.Println("Successfully created", outName)
		continue
	}

	return nil
}

type gsData struct {
	path   string
	reader *bytes.Reader
}

func makeOneGif(pngs []string, outName string) error {
	outGif := &gif.GIF{}

	quantizer := quantize.MedianCutQuantizer{
		Aggregation:    quantize.Mode,
		Weighting:      nil,
		AddTransparent: false,
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return err
	}

	fetches := make(chan gsData)

	for _, input := range pngs {
		go func(input string) {
			pngReader, err := ImportFileFromGoogleStorage(input, client)
			if err != nil {
				log.Println(err)
			}
			dat := gsData{
				reader: pngReader,
				path:   input,
			}

			fetches <- dat

		}(input)
	}

	pngDats := make(map[string]gsData)
	for range pngs {
		dat := <-fetches
		pngDats[dat.path] = dat
	}

	sortedPngDats := make([]gsData, 0, len(pngs))
	for _, png := range pngs {
		sortedPngDats = append(sortedPngDats, pngDats[png])
	}

	// Loop this
	for _, input := range sortedPngDats {
		// pngReader, err := ImportFileFromGoogleStorage(input, client)
		// if err != nil {
		// 	return err
		// }

		img, err := png.Decode(input.reader)
		if err != nil {
			return err
		}

		palette := quantizer.Quantize(make([]color.Color, 0, 256), img)

		palettedImage := image.NewPaletted(img.Bounds(), palette)
		draw.Draw(palettedImage, img.Bounds(), img, image.Point{}, draw.Over)
		outGif.Image = append(outGif.Image, palettedImage)
		outGif.Delay = append(outGif.Delay, 0)
	}

	// Save file
	f, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	return gif.EncodeAll(f, outGif)
}

// ImportFileFromGoogleStorage copies a file from google storage and returns it
// as bytes.
func ImportFileFromGoogleStorage(gsFilePath string, client *storage.Client) (*bytes.Reader, error) {

	// Detect the bucket and the path to the actual file
	pathParts := strings.SplitN(strings.TrimPrefix(gsFilePath, "gs://"), "/", 2)
	if len(pathParts) != 2 {
		return nil, fmt.Errorf("Tried to split your google storage path into 2 parts, but got %d: %v", len(pathParts), pathParts)
	}
	bucketName := pathParts[0]
	pathName := pathParts[1]

	// Open the bucket with default credentials
	bkt := client.Bucket(bucketName)
	handle := bkt.Object(pathName)

	rc, err := handle.NewReader(context.Background())
	if err != nil {
		return nil, pfx.Err(err)
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, pfx.Err(err)
	}

	br := bytes.NewReader(data)

	return br, pfx.Err(err)
}
