package genomisc

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/carbocation/pfx"
)

type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// Decorates a Google Storage object handle with io.Reader, io.Seeker and
// io.Closer. Derived from
// https://github.com/googleapis/google-cloud-go/issues/1124#issuecomment-419070541
type GSReadSeekCloser struct {
	*storage.ObjectHandle
	Context context.Context
	r       *storage.Reader
	offset  int64 // initial offset
	pos     int64 // current position (like 'seen' in storage.Reader)
	Closer  *func() error
}

func (s *GSReadSeekCloser) Read(buf []byte) (int, error) {
	var err error
	if s.r == nil {
		// Note: the -1 for length is necessary because we don't have all the information
		// from the request that http.ServeContent received.
		s.r, err = s.NewRangeReader(s.Context, s.offset, -1)
		if err != nil {
			return 0, err
		}
	}
	n, err := s.r.Read(buf)
	if err != nil {
		return 0, err
	}
	s.pos += int64(n)

	return n, nil
}

func (s *GSReadSeekCloser) Seek(offset int64, whence int) (int64, error) {

	// Seeking is not actually possible. As a proxy, we close the current
	// connection, reset the reader, and reset the offset.
	var newOffset int64

	switch whence {
	case io.SeekStart:
		// Redundant, but being explicit:
		newOffset = 0
	case io.SeekCurrent:
		newOffset = s.offset
	case io.SeekEnd:
		return 0, fmt.Errorf("io.Seeker 'whence' value %d is not implemented", whence)
	}

	// Close the reader and remove it.
	s.r.Close()
	s.r = nil

	// Update our offset
	s.offset = newOffset

	// Not sure that this is the correct thing to do for pos:
	// s.pos = s.offset
	s.pos = 0

	return s.offset, nil
}

// Satisfies io.Closer. If o.close is not set, this is a nop.
func (o *GSReadSeekCloser) Close() error {
	if o.Closer != nil {
		return (*o.Closer)()
	}

	return nil
}

func MaybeOpenSeekerFromGoogleStorage(path string, client *storage.Client) (ReadSeekCloser, int64, error) {
	if client != nil && strings.HasPrefix(path, "gs://") {
		// Detect the bucket and the path to the actual file
		pathParts := strings.SplitN(strings.TrimPrefix(path, "gs://"), "/", 2)
		if len(pathParts) != 2 {
			return nil, 0, fmt.Errorf("Tried to split your google storage path into 2 parts, but got %d: %v", len(pathParts), pathParts)
		}
		bucketName := pathParts[0]
		pathName := pathParts[1]

		// Open the bucket with default credentials
		bkt := client.Bucket(bucketName)
		handle := bkt.Object(pathName)

		wrappedHandle := &GSReadSeekCloser{
			ObjectHandle: handle,
			Context:      context.Background(),

			// Because Close() is called after every read, the final Close() is a
			// nop for this type, and can be left nil
		}

		// Make a hard call to get the filesize
		attrs, err := wrappedHandle.ObjectHandle.Attrs(wrappedHandle.Context)
		if err != nil {
			return nil, 0, pfx.Err(fmt.Errorf("%s: %s", path, err))
		}

		return wrappedHandle, attrs.Size, nil // wrappedHandle.storageReader.Attrs.Size, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return f, 0, err
	}
	fstat, err := f.Stat()
	if err != nil {
		return f, 0, err
	}
	return f, fstat.Size(), nil
}
