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

// Decorates a Google Storage object handle with an io.ReadSeekCloser. Derived
// from
// https://github.com/googleapis/google-cloud-go/issues/1124#issuecomment-419070541
type GSReadSeekCloser struct {
	*storage.ObjectHandle
	Context context.Context
	Closer  *func() error

	r        *storage.Reader
	offset   int64 // initial offset
	pos      int64 // current position (like 'seen' in storage.Reader)
	filesize int64 // if this is known and set, it enables io.SeekEnd
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

	// Set the offset for the next Read or Write to offset, interpreted
	// according to whence
	switch whence {
	case io.SeekStart:
		// Set our new offset relative to 0
		newOffset = offset
	case io.SeekCurrent:
		// Set our new offset relative to our current offset
		newOffset = s.offset + offset
	case io.SeekEnd:
		// Set our new offset relative to the end of the file. If the end of the
		// file is not known, fail.
		if s.filesize == 0 {
			return 0, fmt.Errorf("GSReadSeekCloser.Seek: io.Seeker 'whence' value %d is not implemented", whence)
		}

		newOffset = s.filesize - offset
	}

	// Close the current reader and remove it.
	s.r.Close()
	s.r = nil

	// Update our offset
	s.offset = newOffset

	// Not sure that this is the correct thing to do for pos. (Treating this as
	// how many bytes we have read from the current offset.)
	s.pos = 0

	// Return the new offset relative to the start of the file
	return s.offset, nil
}

// Satisfies io.Closer. If o.close is not set, this is a nop.
func (o *GSReadSeekCloser) Close() error {
	if o.Closer != nil {
		return (*o.Closer)()
	}

	return nil
}

func MaybeOpenSeekerFromGoogleStorage(path string, client *storage.Client) (io.ReadSeekCloser, error) {
	if client != nil && strings.HasPrefix(path, "gs://") {
		// Detect the bucket and the path to the actual file
		pathParts := strings.SplitN(strings.TrimPrefix(path, "gs://"), "/", 2)
		if len(pathParts) != 2 {
			return nil, fmt.Errorf("tried to split your google storage path on 'gs://', but got %d: %v", len(pathParts), pathParts)
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
			return nil, pfx.Err(fmt.Errorf("%s: %s", path, err))
		}

		// Enables io.SeekEnd
		wrappedHandle.filesize = attrs.Size

		return wrappedHandle, nil
	}

	return os.Open(path)
}
