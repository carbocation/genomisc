package bulkprocess

import (
	"math"
	"strconv"
)

// VENC is a struct containing the values necessary for converting a pixel
// intensity to a velocity for that pixel
type VENC struct {
	FlowVenc   float64
	BitsStored float64
}

// NewVENC consumes the raw string values for FlowVenc and BitsStored, returning
// a VENC object that is ready to be used for calculating pixel velocities
func NewVENC(FlowVenc, BitsStored string) (VENC, error) {
	out := VENC{}

	flowVenc, err := strconv.ParseFloat(FlowVenc, 64)
	if err != nil {
		return out, err
	}
	out.FlowVenc = flowVenc

	bitsStored, err := strconv.ParseFloat(BitsStored, 64)
	if err != nil {
		return out, err
	}
	out.BitsStored = bitsStored

	return out, nil
}

// PixelIntensityToVelocity converts a pixel intensity (from DICOM format) into
// a value that ranges from -flowvenc to +flowvenc.
func (v VENC) PixelIntensityToVelocity(pixelIntensity float64) (velocity float64) {

	// Note: I think it should be (2^BS - 1) because, e.g., that's the dynamic
	// range in a 0-based system that has a min value of 0 and a max value of
	// 4095 with a BS of 12.
	slope := 2 * v.FlowVenc / (math.Pow(2, v.BitsStored) - 1)

	intercept := -v.FlowVenc

	return pixelIntensity*slope + intercept
}
