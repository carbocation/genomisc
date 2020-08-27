# Counting connected components with the `overlay` library

This README provides demonstration code, documenting how to use this library to
count connected components within an annotation PNG file.

```go

package main

import (
	"fmt"
	"log"

	"github.com/carbocation/genomisc/overlay"
)

func main() {

	// Config.json file that comports with the type definition within
	// json-config.go
	configPath := "./demo.config.json"

	// Your annotation PNG. Based on the config file, pixels contributing to
	// background should be colored #000000, pixels contributing to ascending
	// aorta #010101, and pixels contributing to descending aorta #020202. No
	// other colors should exist in this image.
	filePath := "./annotation.png"

	// The connected components tool performs a secondary count that ignores
	// connected components that are built from of a small number of pixels
	// below this cutoff. This allows it to ignore small errors, such as an
	// errant, noisy pixel. The units for this value are "pixels". Here,
	// clusters with fewer than 5 pixels are ignored.
	threshold := 5

	// Those are the configurable parameters. Subsequent lines open the image
	// and your config file and apply the connected components algorithm.

	config, err := overlay.ParseJSONConfigFromPath(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	rawOverlayImg, err := overlay.OpenImageFromLocalFile(filePath)
	if err != nil {
		log.Fatalln(err)
	}

	// Generate the function to count connected components within your
	// annotation file
	connected, err := overlay.NewConnected(rawOverlayImg)
	if err != nil {
		log.Fatalln(err)
	}

	// Perform the connected component count
	_, connectedCounts, _, thresholdedConnectedCounts, err := connected.Count(config.Labels, threshold)
	if err != nil {
		log.Fatalln(err)
	}

	// Report the results
	for label := range connectedCounts {
		fmt.Printf("Connected components for %s: %d\n", label.Label, connectedCounts[label])
		fmt.Printf("Connected components for %s, thresholded at %d pixels: %d\n", label.Label, threshold, thresholdedConnectedCounts[label])
	}
}

```
