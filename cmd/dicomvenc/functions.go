package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/aybabtme/uniplot/histogram"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/pfx"
	"github.com/suyashkumar/dicom/dicomtag"
	"gonum.org/v1/gonum/stat"
)

type segmentationPixelBulkProperties struct {
	PixelVelocities                                []float64
	PixelVelocityVelocityStandardDeviationCMPerSec float64
	PixelVelocitySumCMPerSec                       float64
	PixelVelocityAbsoluteSumCMPerSec               float64
	PixelVelocityMinCMPerSec                       float64
	PixelVelocityMaxCMPerSec                       float64
	PixelVelocity01PctCMPerSec                     float64
	PixelVelocity99PctCMPerSec                     float64
	VTI99pctCM                                     float64
	VTIMaxCM                                       float64
	VTIMeanCM                                      float64
	AbsFlowCM3PerSec                               float64
	FlowCM3PerSec                                  float64
}

func describeSegmentationPixels(v []vencPixel, dt, pxHeightCM, pxWidthCM float64) (out segmentationPixelBulkProperties) {
	// Get a measure of dispersion of velocity estimates across the pixels
	out.PixelVelocities = make([]float64, 0, len(v))
	for _, px := range v {
		out.PixelVelocities = append(out.PixelVelocities, px.FlowVenc)
	}
	out.PixelVelocityVelocityStandardDeviationCMPerSec = stat.StdDev(out.PixelVelocities, nil)

	out.PixelVelocityMinCMPerSec = math.MaxFloat32
	out.PixelVelocityMaxCMPerSec = -1.0 * math.MaxFloat32
	for _, px := range v {
		out.PixelVelocitySumCMPerSec += px.FlowVenc
		out.PixelVelocityAbsoluteSumCMPerSec += math.Abs(px.FlowVenc)
		if px.FlowVenc < out.PixelVelocityMinCMPerSec {
			out.PixelVelocityMinCMPerSec = px.FlowVenc
		}
		if px.FlowVenc > out.PixelVelocityMaxCMPerSec {
			out.PixelVelocityMaxCMPerSec = px.FlowVenc
		}
	}

	// Get VTI (units are cm/contraction; here, just unitless cm) This is
	// the integral of venc (cm/sec) over the unit of time (portion of a
	// second). Specifically, we want a peak velocity. To try to reduce
	// error, will take 99% value rather than the very top pixel. Since this
	// is directional, will assess the mean velocity first, and then take
	// the extremum that is directionally consistent with the bulk flow.
	sort.Float64Slice(out.PixelVelocities).Sort()
	out.PixelVelocity01PctCMPerSec = stat.Quantile(0.01, stat.LinInterp, out.PixelVelocities, nil)
	out.PixelVelocity99PctCMPerSec = stat.Quantile(0.99, stat.LinInterp, out.PixelVelocities, nil)

	out.VTI99pctCM = dt * out.PixelVelocity99PctCMPerSec
	out.VTIMaxCM = dt * out.PixelVelocityMaxCMPerSec

	if out.PixelVelocitySumCMPerSec < 0 {
		out.VTI99pctCM = dt * out.PixelVelocity01PctCMPerSec
		out.VTIMaxCM = dt * out.PixelVelocityMinCMPerSec
	}

	out.VTIMeanCM = dt * out.PixelVelocitySumCMPerSec / float64(len(v))

	// Convert to units of "cm^3 / sec"
	out.AbsFlowCM3PerSec = out.PixelVelocityAbsoluteSumCMPerSec * pxHeightCM * pxWidthCM
	out.FlowCM3PerSec = out.PixelVelocitySumCMPerSec * pxHeightCM * pxWidthCM

	return
}

// deltaT returns the assumed amount of time (in seconds) belonging to each
// image in a cycle. It is the entire duration that it takes to acquire all
// images (N), divided by N. Here we assume it is the same for every image in a
// sequence - in all inspected images, this seems to be true.
func deltaT(tagMap map[dicomtag.Tag][]interface{}) (deltaTMS float64, err error) {

	nominalInterval, cardiacNumberOfImages := 0.0, 0.0

	// nominalInterval
	val, exists := tagMap[dicomtag.NominalInterval]
	if !exists {
		return 0, pfx.Err(fmt.Errorf("NominalInterval not found"))
	}
	for _, v := range val {
		nominalInterval, err = strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return 0, pfx.Err(err)
		}
		break
	}

	// cardiacNumberOfImages
	val, exists = tagMap[dicomtag.CardiacNumberOfImages]
	if !exists {
		return 0, pfx.Err(fmt.Errorf("CardiacNumberOfImages not found"))
	}
	for _, v := range val {
		cardiacNumberOfImages, err = strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return 0, pfx.Err(err)
		}
		break
	}

	// Convert milliseconds to seconds
	return (1 / 1000.0) * nominalInterval / cardiacNumberOfImages, nil
}

func pixelHeightWidthCM(tagMap map[dicomtag.Tag][]interface{}) (pxHeightCM, pxWidthCM float64, err error) {
	val, exists := tagMap[dicomtag.PixelSpacing]
	if !exists {
		return 0, 0, pfx.Err(fmt.Errorf("PixelSpacing not found"))
	}

	for k, v := range val {
		if k == 0 {
			pxHeightCM, err = strconv.ParseFloat(v.(string), 32)
			if err != nil {
				continue
			}
		} else if k == 1 {
			pxWidthCM, err = strconv.ParseFloat(v.(string), 32)
			if err != nil {
				continue
			}
		}
	}

	// They're in mm -- convert to cm
	pxHeightCM *= 0.1
	pxWidthCM *= 0.1

	return
}

func fetchFlowVenc(tagMap map[dicomtag.Tag][]interface{}) (*bulkprocess.VENC, error) {

	var bitsStored uint16
	var venc string

	// Bits Stored:
	bs, exists := tagMap[dicomtag.BitsStored]
	if !exists {
		return nil, fmt.Errorf("BitsStored not found")
	}

	for _, v := range bs {
		bitsStored = v.(uint16)
		break
	}

	// Flow VENC:

	// Siemens header data requires special treatment
	ve, exists := tagMap[dicomtag.Tag{Group: 0x0029, Element: 0x1010}]
	if !exists {
		return nil, fmt.Errorf("Siemens header not found")
	}

	for _, v := range ve {
		sc, err := bulkprocess.ParseSiemensHeader(v)
		if err != nil {
			return nil, err
		}

		for _, v := range sc.Slice() {
			if v.Name != "FlowVenc" {
				continue
			}

			for _, encoding := range v.SubElementData {
				venc = encoding
				break
			}
		}
	}

	out, err := bulkprocess.NewVENC(venc, bitsStored)

	return &out, err
}

func deAlias(pixdat segmentationPixelBulkProperties, flowVenc *bulkprocess.VENC, dt, pxHeightCM, pxWidthCM float64, pixels []vencPixel) (outPixdat segmentationPixelBulkProperties, unwrappedPixels []vencPixel, unwrapped bool) {
	// Generate a histogram. The number of buckets is arbitrary. TODO: find a
	// rational bucket count.
	hist := histogram.Hist(25, pixdat.PixelVelocities)

	unwrappedV := make([]vencPixel, len(pixels))
	copy(unwrappedV, pixels)

	if pixdat.PixelVelocitySumCMPerSec > 0 {
		// If the bulk flow is positive, start at the most negative value:

		zeroBucket := false
		wrapPoint := math.Inf(-1)
		for _, bucket := range hist.Buckets {
			wrapPoint = bucket.Max
			if bucket.Count == 0 {
				zeroBucket = true
				break
			}
		}

		// If we never saw a zero bucket, there might not actually be aliasing -
		// the full range might be used
		if zeroBucket {
			for k, vp := range unwrappedV {
				if unwrappedV[k].FlowVenc < wrapPoint {
					unwrappedV[k].FlowVenc = 2.0*flowVenc.FlowVenc + vp.FlowVenc
				}
			}

			// And replace our stats
			outPixdat = describeSegmentationPixels(unwrappedV, dt, pxHeightCM, pxWidthCM)
			unwrapped = true
		}
	} else if pixdat.PixelVelocitySumCMPerSec < 0 {
		// If the flow is negative, reverse the sort, so we can start at the
		// highest values and go downward.

		zeroBucket := false
		wrapPoint := math.Inf(1)
		sort.Slice(hist.Buckets, func(i, j int) bool {
			return hist.Buckets[i].Min < hist.Buckets[j].Min
		})

		for _, bucket := range hist.Buckets {
			wrapPoint = bucket.Min
			if bucket.Count == 0 {
				zeroBucket = true
				break
			}
		}

		// If we never saw a zero bucket, there might not actually be aliasing -
		// the full range might be used
		if zeroBucket {
			for k, vp := range unwrappedV {
				if unwrappedV[k].FlowVenc > wrapPoint {
					unwrappedV[k].FlowVenc = -2.0*flowVenc.FlowVenc + vp.FlowVenc
				}
			}

			// And replace our stats
			outPixdat = describeSegmentationPixels(unwrappedV, dt, pxHeightCM, pxWidthCM)
			unwrapped = true
		}
	}

	return outPixdat, unwrappedV, unwrapped
}
