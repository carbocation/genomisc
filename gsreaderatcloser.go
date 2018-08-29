package genomisc

import (
	"context"

	"cloud.google.com/go/storage"
)

// Decorates a Google Storage object handle with ReadAt
type GSReaderAtCloser struct {
	storage.ObjectHandle
	Context context.Context
	close   *func() error
}

// ReadAt satisfies io.ReaderAt. Note that this is dependent upon making p a
// buffer of the desired length to be read by NewRangeReader.
func (o GSReaderAtCloser) ReadAt(p []byte, offset int64) (n int, err error) {
	rdr, err := o.NewRangeReader(o.Context, offset, int64(len(p)))
	if err != nil {
		return 0, err
	}

	return rdr.Read(p)
}

// Satisfies io.Closer. If o.close is not set, this is a nop.
func (o GSReaderAtCloser) Close() error {
	if o.close != nil {
		return (*o.close)()
	}

	return nil
}
