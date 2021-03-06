package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/henghuang/nifti"
)

func main() {
	var filename, output, imageSuffix string
	var isGrayscale bool

	flag.StringVar(&filename, "file", "", "Name of .nii or .nii.gz file to convert to PNGs. ")
	flag.StringVar(&output, "out", "", "Name of folder where the pngs will be emitted. Filenames will be {orig_filename}.z{z depth}_t{time}.png.")
	flag.BoolVar(&isGrayscale, "grayscale", false, "If true, creates a grayscale image. Otherwise creates a color image.")
	flag.StringVar(&imageSuffix, "suffix", ".png", "Generally, use .png for raw images and .png.mask.png for annotations. Note that this does not determine filetype - files are always png.")
	flag.Parse()

	if filename == "" || output == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	prefix := filepath.Base(filename)
	prefix = strings.TrimSuffix(prefix, ".nii.gz")
	prefix = strings.TrimSuffix(prefix, ".nii")

	if stat, err := os.Stat(output); err == nil && stat.IsDir() {
		// path is a directory already
	} else {
		os.MkdirAll(output, os.ModePerm)
	}

	niftiImage, err := SafelyNiftiParse(filename, true)
	if err != nil {
		log.Fatalln(err)
	}

	niftiHeader, err := SafelyNiftiHeaderParse(filename)
	if err != nil {
		log.Fatalln(err)
	}

	if err := nifti2png(niftiImage, niftiHeader, prefix, output, imageSuffix, isGrayscale); err != nil {
		log.Fatalln(err)
	}
}

func nifti2png(input nifti.Nifti1Image, niftiHeader nifti.Nifti1Header, prefix, output, imageSuffix string, isGrayscale bool) error {
	dims := input.GetDims()
	xm, ym, zm, tm := dims[0], dims[1], dims[2], dims[3]

	rect := image.Rect(0, 0, xm, ym)
	img := image.NewGray16(rect)
	colImg := image.NewRGBA(rect)
	var grayCol color.Color
	var col color.Color

	fmt.Printf("sample_id\tdicom_file\tb_slice\tinstance_number\theight_mm\twidth_mm\tdepth_mm\tinstance\n")

	// March forward in time
	for t := 0; t < tm; t++ {
		// And down the stack
		for z := 0; z < zm; z++ {
			maxIntensity := 0.0
			// Then across x and y
			for i := 0; i < 2; i++ {
				for x := 0; x < xm; x++ {
					for y := 0; y < ym; y++ {
						intensity := float64(input.GetAt(x, y, z, t))
						if i == 0 {
							if intensity > maxIntensity {
								maxIntensity = intensity
							}

							continue
						}

						grayCol = color.Gray16{Y: applyPythonicWindowScaling(intensity, maxIntensity)}
						if isGrayscale {
							img.Set(x, y, grayCol)
						} else {
							col = color.RGBA64Model.Convert(grayCol)
							colImg.Set(x, y, col)
						}
					}
				}
			}

			f, err := os.Create(filepath.Join(output, fmt.Sprintf("%s.z%06d_t%06d%s", prefix, z, t, imageSuffix)))
			if err != nil {
				return err
			}
			fw := bufio.NewWriter(f)

			if isGrayscale {
				err = png.Encode(fw, img)
			} else {
				err = png.Encode(fw, colImg)
			}
			if err != nil {
				return err
			}
			// Emit metadata about each PNG
			sampleID := strings.Split(prefix, "_")[0]
			fmt.Printf("%s\t%s\t%d\t%d\t%g\t%g\t%g\t%d\n", sampleID, fmt.Sprintf("%s.z%06d_t%06d", prefix, z, t), z, t, niftiHeader.Pixdim[1], niftiHeader.Pixdim[2], niftiHeader.Pixdim[3], 0)

			fw.Flush()
			f.Close()

		}
	}

	return nil
}

func applyPythonicWindowScaling(intensity, maxIntensity float64) uint16 {
	if intensity < 0 {
		intensity = 0
	}

	return uint16(float64(math.MaxUint16) * intensity / maxIntensity)
}
