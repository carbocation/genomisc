package bulkprocess

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/dicomtag"
)

// DicomToTagMap takes in a dicom file (in bytes), emits a map of tags
func DicomToTagMap(dicomReader io.Reader) (map[dicomtag.Tag][]interface{}, error) {
	out := make(map[dicomtag.Tag][]interface{})

	dcm, err := ioutil.ReadAll(dicomReader)
	if err != nil {
		return nil, err
	}

	p, err := dicom.NewParserFromBytes(dcm, nil)
	if err != nil {
		return nil, err
	}

	parsedData, err := SafelyDicomParse(p, dicom.ParseOptions{
		DropPixelData: false,
	})
	if parsedData == nil || err != nil {
		return nil, fmt.Errorf("Error reading dicom: %v", err)
	}

	for _, elem := range parsedData.Elements {
		if elem == nil {
			continue
		}

		entry := out[elem.Tag]
		entry = elem.Value
		out[elem.Tag] = entry

	}

	return out, nil
}
