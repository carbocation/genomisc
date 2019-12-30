package main

import (
	"encoding/json"
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"

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

func main() {
	var basePath, overlayPath, outputFolder, jsonConfig string
	var opacity uint

	flag.StringVar(&basePath, "base", "", "Path to base image")
	flag.StringVar(&overlayPath, "overlay", "", "Path to overlay image")
	flag.StringVar(&outputFolder, "output_folder", "", "Folder where output file should be created")
	flag.StringVar(&jsonConfig, "config", "", "JSONConfig file from the github.com/carbocation/genomisc/overlay package")
	flag.UintVar(&opacity, "opacity", 128, "Opacity of the overlay, from 0-255.")
	flag.Parse()

	if basePath == "" || overlayPath == "" || jsonConfig == "" {
		flag.Usage()
		os.Exit(1)
	}

	config, err := overlay.ParseJSONConfigFromPath(jsonConfig)
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(config, basePath, overlayPath, outputFolder, uint8(opacity)); err != nil {
		log.Fatalln(err)
	}

}

func run(config overlay.JSONConfig, basePath, overlayPath, outputFolder string, opacity uint8) error {

	baseImg, err := openImage(basePath)
	if err != nil {
		return err
	}

	rawOverlayImg, err := openImage(overlayPath)
	if err != nil {
		return err
	}

	// Use the JSONConfig to identify the desired human-visible colors for the
	// overlay:
	overlayImg, err := config.Labels.DecodeImageFromImageSegment(rawOverlayImg)
	if err != nil {
		return err
	}

	// Perform overlay
	oImg := image.NewRGBA(baseImg.Bounds())

	// Copy the base image to a new output image satisfying the draw.Image
	// interface:
	draw.Draw(oImg, oImg.Bounds(), baseImg, image.ZP, draw.Src)

	// Overlay the new image on top, with defined opacity:
	mask := image.NewUniform(color.Alpha{opacity})
	draw.DrawMask(oImg, oImg.Bounds(), overlayImg, image.ZP, mask, image.ZP, draw.Over)

	// Create your output image file
	of, err := os.Create(outputFolder + "/" + filepath.Base(basePath) + ".overlay.png")
	if err != nil {
		return err
	}
	defer of.Close()

	// Write the PNG representation of your output image to the file
	if err := png.Encode(of, oImg); err != nil {
		return err
	}

	return nil
}

func openImage(pathTo string) (image.Image, error) {
	f, err := os.Open(pathTo)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)

	return img, err
}
