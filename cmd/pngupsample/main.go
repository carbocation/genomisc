// pngupsample ingests a PNG and parameters that indicate (1) how to resize
// pixels and (2) which subset of the image to upsample, then produces the
// subsetted, upsampled PNG. Upsampling is intentionally basic (nearest
// neighbor) to minimize induction of artifact.
package main

import (
	"flag"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/carbocation/genomisc/overlay"
)

func main() {
	var basePath, outputFolder string
	var topLeftX, topLeftY, bottomRightX, bottomRightY, scale, dilate int

	flag.StringVar(&basePath, "base", "", "Path to source PNG.")
	flag.StringVar(&outputFolder, "output_folder", "", "Folder where output file should be created")
	flag.IntVar(&bottomRightX, "bottomrightx", 0, "0-based bottom right pixel, X. If zero, will be set to max X. Pixel number starts from (0,0) at the top left corner.")
	flag.IntVar(&bottomRightY, "bottomrighty", 0, "0-based bottom right pixel, Y. If zero, will be set to max Y.")
	flag.IntVar(&topLeftX, "topleftx", 0, "0-based top left pixel, X.")
	flag.IntVar(&topLeftY, "toplefty", 0, "0-based top left pixel, Y")
	flag.IntVar(&scale, "scale", 1, "Each pixel in the input will become this many pixels in the output.")
	flag.IntVar(&dilate, "dilate", 0, "Expand top left to the top left by this much; expand bottom right to the bottom right by this much.")
	flag.Parse()

	if basePath == "" || outputFolder == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(basePath, outputFolder, topLeftX, topLeftY, bottomRightX, bottomRightY, scale, dilate); err != nil {
		log.Fatalln(err)
	}

}

func run(basePath, outputFolder string, topLeftX, topLeftY, bottomRightX, bottomRightY, scale, dilation int) error {

	baseImg, err := openImage(basePath)
	if err != nil {
		return err
	}

	outImg, err := overlay.SubsetAndRescaleImage(baseImg, topLeftX, topLeftY, bottomRightX, bottomRightY, scale, dilation)
	if err != nil {
		return err
	}

	// Create your output image file (scale_tXxtY_bXxbY_ScaleFactor_name)
	of, err := os.Create(outputFolder + "/scaled_" +
		strconv.Itoa(topLeftX) + "-" + strconv.Itoa(topLeftY) + "_" +
		strconv.Itoa(bottomRightX) + "-" + strconv.Itoa(bottomRightY) + "_" +
		strconv.Itoa(scale) + "_" +
		filepath.Base(basePath))
	if err != nil {
		return err
	}
	defer of.Close()

	// Write the PNG representation of your output image to the file
	if err := png.Encode(of, outImg); err != nil {
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
