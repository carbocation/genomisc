package main

import (
	"fmt"

	"github.com/henghuang/nifti"
)

// SafelyNiftiParse consumes panics emitted by the nifti library, which are
// inappropriate and must be captured in order to turn them into recoverable
// errors.
func SafelyNiftiParse(filename string, rdata bool) (parsedData nifti.Nifti1Image, err error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf("%v", panicErr)
		}
	}()

	parsedData.LoadImage(filename, rdata)

	return
}

// SafelyNiftiHeaderParse consumes panics emitted by the nifti library, which are
// inappropriate and must be captured in order to turn them into recoverable
// errors.
func SafelyNiftiHeaderParse(filename string) (parsedData nifti.Nifti1Header, err error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf("%v", panicErr)
		}
	}()

	parsedData.LoadHeader(filename)

	return
}
