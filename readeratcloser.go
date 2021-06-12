package genomisc

import "io"

type ReaderAtCloser interface {
	io.Reader
	io.ReaderAt
	io.Closer
}
