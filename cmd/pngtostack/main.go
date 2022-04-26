package main

import (
	"context"
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
)

// Safe for concurrent use by multiple goroutines
var client *storage.Client

func main() {
	var input string
	var widthMM, heightMM, rescaleFactor, xGap, yGap, xOriginOffset, yOriginOffset, floatImgOpacity float64
	var breakAfter int
	flag.StringVar(&input, "input", "", "One of: (a) a folder containing png files, (b) a single png file, or (c) a .tar.gz file containing png files")
	flag.Float64Var(&widthMM, "width", 152, "Width of the output image in millimeters")
	flag.Float64Var(&heightMM, "height", 152, "Height of the output image in millimeters")
	flag.Float64Var(&rescaleFactor, "rescale", 1., "Rescale factor for the output image")
	flag.Float64Var(&xGap, "xgap", 50, "X-gap between images")
	flag.Float64Var(&yGap, "ygap", 50, "Y-gap between images")
	flag.Float64Var(&xOriginOffset, "xoffset", 0, "X-offset between top left point of output image and first input image")
	flag.Float64Var(&yOriginOffset, "yoffset", 0, "Y-offset between top left point of output image and first input image")
	flag.Float64Var(&floatImgOpacity, "opacity", 0.8, "Opacity of overlay images. From 0 (transparent) to 1 (opaque)")
	flag.IntVar(&breakAfter, "break", 0, "Break after this many images. (0 means no break, process all images.)")
	flag.Parse()

	if input == "" {
		flag.Usage()
		os.Exit(1)
	}

	imgOpacity := uint8(floatImgOpacity * 255)

	var err error
	if strings.HasPrefix(input, "gs://") {
		client, err = storage.NewClient(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}

	var images []image.Image
	switch {
	case strings.HasSuffix(input, ".tar.gz"):
		images, err = loadImagesFromTarGz(input)
	case strings.HasSuffix(input, ".gif"):
		images, err = loadImagesFromGIF(input)
	case strings.HasSuffix(input, ".png"):
		imgReader, err := os.Open(input)
		if err != nil {
			log.Fatalln(err)
		}
		img, err := png.Decode(imgReader)
		images = []image.Image{img}
	default:
		images, err = loadImagesFromFolder(input)
	}
	if err != nil {
		log.Fatalln(err)
	}

	out := pngStack(images, xGap, yGap, xOriginOffset, yOriginOffset, widthMM, heightMM, rescaleFactor, imgOpacity, breakAfter)
	if err := png.Encode(os.Stdout, out); err != nil {
		log.Fatalln(err)
	}
}

func pngStack(images []image.Image, xGap, yGap, xOriginOffset, yOriginOffset, widthMM, heightMM, rescaleFactor float64, imgOpacity uint8, breakAfter int) image.Image {
	// Nature Genetics permits 180mm, for example
	c := canvas.New(widthMM, heightMM)

	for i, imgIn := range images {
		if i > breakAfter && breakAfter > 0 {
			break
		}
		imgIn = setImageOpacity(imgIn, imgOpacity)

		// initialYOffset := float64(imgIn.Bounds().Dy())

		ctx := canvas.NewContext(c)
		ctx.SetCoordSystem(canvas.CartesianIV)

		// Master rescale factor
		ctx.Scale(rescaleFactor, rescaleFactor)

		initialYOffset := ctx.Height()

		addImageToContext(imgIn, ctx, xOriginOffset+xGap*float64(i), initialYOffset+yOriginOffset+yGap*float64(i))
	}

	img := rasterizer.Draw(c, canvas.Resolution(1.9), canvas.DefaultColorSpace)

	return img
}

func addImageToContext(img image.Image, ctx *canvas.Context, xOffsetPixels, yOffsetPixels float64) {

	// Reduce the x-width
	ctx.Scale(0.4, 1)

	// Shear upwards
	ctx.Shear(0, 0.2)

	ctx.DrawImage(xOffsetPixels, yOffsetPixels, img, 1)
	ctx.Close()

	return
}

func setImageOpacity(src image.Image, alpha uint8) image.Image {
	dst := image.NewRGBA(src.Bounds())
	uni := image.NewUniform(color.RGBA{0, 0, 0, alpha})
	draw.DrawMask(dst, dst.Bounds(), src, image.Point{}, uni, image.Point{}, draw.Over)
	return dst
}

func savePNGToFile(img image.Image, outName string) error {
	f, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

func savePNGToWriter(img image.Image, w io.Writer) error {
	return png.Encode(w, img)
}
