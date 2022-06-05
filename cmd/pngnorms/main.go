package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/carbocation/genomisc/overlay"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/runningvariance"
)

const CheckEveryPixels = 100_000_000

type Stat struct {
	runningvariance.RunningStat
	Min float64
	Max float64
}

func NewStat() *Stat {
	return &Stat{
		*runningvariance.NewRunningStat(),
		math.MaxFloat64,
		math.SmallestNonzeroFloat64,
	}
}

// var channelNorms := make()
type ChannelNorms struct {
	m  sync.Mutex
	rv []*Stat
}

func (c *ChannelNorms) Push(x float64, channel int) error {
	c.m.Lock()
	defer c.m.Unlock()

	if channel > len(c.rv)-1 {
		return fmt.Errorf("channel %d out of range (%d channels)", channel, len(c.rv))
	}

	if c.rv[channel] == nil {
		log.Println("Initializing channel", channel)
		c.rv[channel] = NewStat()
	}

	c.rv[channel].Push(x)
	if c.rv[channel].N%CheckEveryPixels == 0 {
		log.Println("Channel", channel, "pixel #", c.rv[channel].N, ". Min/max:", c.rv[channel].Min, c.rv[channel].Max, "Running mean: ", c.rv[channel].Mean(), "Std:", c.rv[channel].StandardDeviation())
	}
	if x > c.rv[channel].Max {
		c.rv[channel].Max = x
	}
	if x < c.rv[channel].Min {
		c.rv[channel].Min = x
	}

	return nil
}

// Safe for concurrent use by multiple goroutines
var client *storage.Client

func main() {
	var filePath string
	var forceGrayscale bool

	flag.StringVar(&filePath, "path", "", "path to the folder with image files (or .tar.gz of image files) whose norms you wish to obtain")
	flag.BoolVar(&forceGrayscale, "force-grayscale", false, "convert color model to grayscale regardless of true color model")
	flag.Parse()

	if filePath == "" {
		flag.Usage()
		return
	}

	// Initialize the Google Storage client only if we're pointing to Google
	// Storage paths.
	if strings.HasPrefix(filePath, "gs://") {
		var err error
		client, err = storage.NewClient(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}

	cn := &ChannelNorms{}
	if forceGrayscale {
		cn.rv = make([]*Stat, 1)
	} else {
		cn.rv = make([]*Stat, 3)
	}

	if err := runFolder(filePath, forceGrayscale, cn); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Unscaled:")
	for i, channel := range cn.rv {
		fmt.Println("Channel", i, "based on", channel.N, "pixels")
		fmt.Println("Max:", channel.Max)
		fmt.Println("Min:", channel.Min)
		fmt.Println("Mean:", channel.Mean())
		fmt.Println("Std:", channel.StandardDeviation())
	}

	fmt.Println("Scaled:")
	for i, channel := range cn.rv {
		fmt.Println("Channel", i, "based on", channel.N, "pixels")
		fmt.Println("Max:", channel.Max/channel.Max)
		fmt.Println("Min:", channel.Min/channel.Max)
		fmt.Println("Mean:", channel.Mean()/channel.Max)
		fmt.Println("Std:", channel.StandardDeviation()/channel.Max)
	}
}

func processOneImage(img image.Image, forceGrayscale bool, cn *ChannelNorms) error {

	// Iterate over all pixels:
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			nativePixel := img.At(x, y)

			if forceGrayscale {
				col := color.Gray16Model.Convert(nativePixel)
				if err := cn.Push(float64(col.(color.Gray16).Y), 0); err != nil {
					return err
				}
			} else {
				r, g, b, _ := nativePixel.RGBA()
				rv := []float64{float64(r), float64(g), float64(b)}
				for i := 0; i < len(rv); i++ {
					if err := cn.Push(rv[i], i); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func runFolder(path1 string, forceGrayscale bool, cn *ChannelNorms) error {

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
			var err error
			if strings.HasSuffix(file, ".tar.gz") {
				if err = processOneTarGZFilepath(path1+"/"+file, forceGrayscale, cn); err != nil {
					log.Println(err)
				}
			} else if strings.HasSuffix(file, ".png") ||
				strings.HasSuffix(file, ".gif") ||
				strings.HasSuffix(file, ".bmp") ||
				strings.HasSuffix(file, ".jpeg") ||
				strings.HasSuffix(file, ".jpg") {
				if err = processOneImageFromPath(path1+"/"+file, forceGrayscale, cn); err != nil {
					log.Printf("%s: %s\n", file, err)
				}
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

func processOneTarGZFilepath(filePath1 string, forceGrayscale bool, cn *ChannelNorms) error {
	// Reader: Open and stream/ungzip the tar.gz
	f, _, err := bulkprocess.MaybeOpenFromGoogleStorage(filePath1, client)
	if err != nil {
		return fmt.Errorf("%s: %w", filePath1, err)
	}
	defer f.Close()
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("%s: %w", filePath1, err)
	}
	defer gzr.Close()
	tarReader := tar.NewReader(gzr)

	// Iterate over tarfile contents, processing all non-directory files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("%s: %w", filePath1, err)
		} else if header.Typeflag != tar.TypeReg {
			continue
		}

		// Decode image bytes from the tar reader
		newImg, err := bulkprocess.DecodeImageFromReader(tarReader)
		if err != nil {
			log.Println(fmt.Errorf("%s->%s: %w", filePath1, header.Name, err))
			continue
		}

		// Process image
		err = processOneImage(newImg, forceGrayscale, cn)
		if err != nil {
			log.Printf("%s: %v\n", header.Name, err)
			continue
		}
	}

	return nil
}

func processOneImageFromPath(path string, forceGrayscale bool, cn *ChannelNorms) error {
	// Open the files
	img, err := overlay.OpenImageFromLocalFileOrGoogleStorage(path, nil)
	if err != nil {
		return err
	}

	if err = processOneImage(img, forceGrayscale, cn); err != nil {
		return err
	}

	return nil
}

// Via https://flaviocopes.com/go-list-files/
func scanFolder(dirname string) ([]os.FileInfo, error) {

	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}

	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}

	return files, nil
}
