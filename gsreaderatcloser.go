package genomisc

import (
	"context"

	"cloud.google.com/go/storage"
)

// Decorates a Google Storage object handle with ReadAt
type GSReaderAtCloser struct {
	*storage.ObjectHandle
	Context context.Context
	Closer  *func() error
}

// ReadAt satisfies io.ReaderAt. Note that this is dependent upon making p a
// buffer of the desired length to be read by NewRangeReader.
func (o GSReaderAtCloser) ReadAt(p []byte, offset int64) (n int, err error) {
	rdr, err := o.NewRangeReader(o.Context, offset, int64(len(p)))
	if err != nil {
		return 0, err
	}
	defer rdr.Close()

	// Apparently with Google Cloud Storage we get to manually handle the fact
	// that the reader will not read as much as we request, but instead may read
	// less than that. And so here, we manually check if that's the case; if so,
	// we loop until we have read at least as much as we requested.
	var nBytes int
	for {
		innerBytes, err := rdr.Read(p[nBytes:])
		nBytes += innerBytes
		if err != nil {
			return nBytes, err
		}

		if rdr.Remain() <= 0 {
			break
		}
	}

	return nBytes, nil
}

// Satisfies io.Closer. If o.close is not set, this is a nop.
func (o GSReaderAtCloser) Close() error {
	if o.Closer != nil {
		return (*o.Closer)()
	}

	return nil
}
