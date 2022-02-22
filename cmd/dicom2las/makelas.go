package main

import (
	"fmt"
	"image"
	"log"
	"math"

	"github.com/jblindsay/lidario"
)

type Dicom3DPoint struct {
	X, Y, Z   float64
	Intensity uint16
}

func dicomTo3DPoints(dicomEntries []manifestEntry, imgMap map[string]image.Image) ([]Dicom3DPoint, error) {

	voxels := 0
	for _, entry := range dicomEntries {
		img := imgMap[entry.dicom]
		voxels += img.Bounds().Dx() * img.Bounds().Dy()
	}

	points := make([]Dicom3DPoint, 0, voxels)
	minX := math.MaxFloat64
	minY := math.MaxFloat64
	minZ := math.MaxFloat64

	for _, entry := range dicomEntries {
		currentImage := imgMap[entry.dicom]

		x0 := entry.ImagePositionPatientX
		y0 := entry.ImagePositionPatientY
		z := entry.ImagePositionPatientZ

		img, ok := currentImage.(*image.Gray16)
		if !ok {
			return nil, fmt.Errorf("expected DICOM image to be grayscale image, but got %T", img)
		}

		for i := 0; i < img.Bounds().Max.X; i++ {
			x := x0 + float64(i)*entry.PixelWidthNativeX
			for j := 0; j < img.Bounds().Max.Y; j++ {
				y := y0 + float64(j)*entry.PixelWidthNativeY

				points = append(points,
					Dicom3DPoint{
						X:         x,
						Y:         y,
						Z:         z,
						Intensity: img.Gray16At(i, j).Y,
					},
				)

				// Find min coordinate
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if z < minZ {
					minZ = z
				}
			}
		}
	}

	// Offset all points by the minimum coordinate
	for i := range points {
		points[i].X -= minX
		points[i].Y -= minY
		points[i].Z -= minZ
	}

	return points, nil
}

func makeLAS(dicomEntries []manifestEntry, imgMap map[string]image.Image, outName string) error {
	// Create a new LAS file
	newLf, err := lidario.NewLasFile(outName, "w")
	if err != nil {
		return err
	}
	defer newLf.Close()

	voxels, err := dicomTo3DPoints(dicomEntries, imgMap)
	if err != nil {
		return err
	}

	if err := newLf.AddHeader(lidario.LasHeader{
		PointFormatID: 2,
	}); err != nil {
		return err
	}

	for _, voxel := range voxels {
		p := &lidario.PointRecord2{
			PointRecord0: &lidario.PointRecord0{},
			RGB:          &lidario.RgbData{},
		}

		p.X = voxel.X
		p.Y = voxel.Y
		p.Z = voxel.Z
		p.Intensity = voxel.Intensity
		p.RGB.Red = voxel.Intensity
		p.RGB.Blue = voxel.Intensity
		p.RGB.Green = voxel.Intensity

		if err := newLf.AddLasPoint(p); err != nil {
			return err
		}
	}

	if err := newLf.Close(); err != nil {
		return err
	}

	return nil
}

func makeLASDemo(dicomEntries []manifestEntry, imgMap map[string]image.Image, outName string) error {
	// Create a new LAS file
	newLf, err := lidario.NewLasFile(outName, "w")
	if err != nil {
		return err
	}
	defer newLf.Close()

	if err := newLf.AddHeader(lidario.LasHeader{}); err != nil {
		return err
	}

	pts := make([]lidario.LasPointer, 0, 10)
	for i := 0; i < cap(pts); i++ {
		p := &lidario.PointRecord2{
			PointRecord0: &lidario.PointRecord0{},
			RGB:          &lidario.RgbData{},
		}

		p.X = 0
		p.Y = 0
		p.Z = float64(i)
		p.RGB.Red = 0
		p.RGB.Blue = math.MaxUint16 / 2
		p.RGB.Green = math.MaxUint16

		pts = append(pts, p)
	}

	log.Println("Processed", len(pts), "points")

	// newLf.Header.NumberPoints = 10

	if err := newLf.AddLasPoints([]lidario.LasPointer(pts)); err != nil {
		return err
	}

	return newLf.Close()
}
