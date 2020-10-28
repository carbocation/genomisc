package bulkprocess

import (
	"fmt"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/element"
)

// SafelyDicomParse consumes panics emitted by the dicom library, which are
// inappropriate and must be captured in order to turn them into recoverable
// errors.
func SafelyDicomParse(p dicom.Parser, opts dicom.ParseOptions) (parsedData *element.DataSet, err error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf("%v", panicErr)
		}
	}()

	return p.Parse(opts)
}
