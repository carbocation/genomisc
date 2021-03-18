// +build !cgo

package prsworker

import (
	"github.com/carbocation/bgen"
)

func OpenBGI(path string) (*bgen.BGIIndex, error) {
	return bgen.OpenBGI(path)
}
