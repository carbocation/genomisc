package bulkprocess

import (
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/dicomtag"
)

// DicomMeta holds a small subset of the available metadata which we consider to
// be useful from dicom images.
type DicomMeta struct {
	Date              string
	HasOverlay        bool
	OverlayFraction   float64
	OverlayRows       int
	OverlayCols       int
	InstanceNumber    string
	PatientX          float64
	PatientY          float64
	PatientZ          float64
	PixelHeightMM     float64
	PixelWidthMM      float64
	SliceThicknessMM  float64
	SeriesDescription string
}

// Takes in a dicom file (in bytes), emit meta-information
func DicomToMetadata(dicomReader io.Reader) (*DicomMeta, error) {
	dcm, err := ioutil.ReadAll(dicomReader)
	if err != nil {
		return nil, err
	}

	p, err := dicom.NewParserFromBytes(dcm, nil)
	if err != nil {
		return nil, err
	}

	parsedData, err := p.Parse(dicom.ParseOptions{
		DropPixelData: true,
	})
	if parsedData == nil || err != nil {
		return nil, fmt.Errorf("Error reading dicom: %v", err)
	}

	output := &DicomMeta{}
	for _, elem := range parsedData.Elements {
		if elem.Tag.Compare(dicomtag.Tag{Group: 0x6000, Element: 0x3000}) == 0 {
			output.HasOverlay = true

			activeCells := 0
			totalCells := 0

			for _, enclosed := range elem.Value {
				cellVals, ok := enclosed.([]byte)
				if !ok {
					continue
				}

				for _, cell := range cellVals {
					totalCells++
					if cell != 0 {
						activeCells++
					}
				}
			}

			if totalCells > 0 {
				output.OverlayFraction = float64(activeCells) / float64(totalCells)
			}
		}

		if elem.Tag.Compare(dicomtag.Tag{Group: 0x6000, Element: 0x0010}) == 0 {
			output.OverlayRows = int(elem.Value[0].(uint16))
		}

		if elem.Tag.Compare(dicomtag.Tag{Group: 0x6000, Element: 0x0011}) == 0 {
			output.OverlayCols = int(elem.Value[0].(uint16))
		}

		if elem.Tag.Compare(dicomtag.Tag{Group: 0x0020, Element: 0x0013}) == 0 {
			output.InstanceNumber = elem.Value[0].(string)
		}

		if elem.Tag.Compare(dicomtag.Tag{Group: 0x0020, Element: 0x0032}) == 0 {
			output.PatientX, err = strconv.ParseFloat(elem.Value[0].(string), 32)
			if err != nil {
				continue
			}
			output.PatientY, err = strconv.ParseFloat(elem.Value[1].(string), 32)
			if err != nil {
				continue
			}
			output.PatientZ, err = strconv.ParseFloat(elem.Value[2].(string), 32)
			if err != nil {
				continue
			}
		}

		if elem.Tag.Compare(dicomtag.Tag{Group: 0x0028, Element: 0x0030}) == 0 {
			for k, v := range elem.Value {
				if k == 0 {
					output.PixelHeightMM, err = strconv.ParseFloat(v.(string), 32)
					if err != nil {
						continue
					}
				} else if k == 1 {
					output.PixelWidthMM, err = strconv.ParseFloat(v.(string), 32)
					if err != nil {
						continue
					}
				}
			}
		}

		if elem.Tag == dicomtag.SliceThickness {
			for k, v := range elem.Value {
				if k == 0 {
					output.SliceThicknessMM, err = strconv.ParseFloat(v.(string), 32)
					if err != nil {
						continue
					}
				}
			}
		}

		if elem.Tag == dicomtag.SeriesDescription {
			for k, v := range elem.Value {
				if k == 0 {
					output.SeriesDescription = v.(string)
				}
			}
		}

		if elem.Tag == dicomtag.Date || elem.Tag == dicomtag.DateTime || elem.Tag == dicomtag.AcquisitionDate {
			for k, v := range elem.Value {
				if k == 0 && len(v.(string)) > 0 {
					output.Date = v.(string)
				}
			}
		}

	}

	return output, nil
}
