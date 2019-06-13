package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/carbocation/pfx"
)

func MaybeOpenFromGoogleStorage(path string, client *storage.Client) (ReaderAtCloser, int64, error) {
	if strings.HasPrefix(path, "gs://") {
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

		wrappedHandle := &GSReaderAtCloser{
			ObjectHandle: handle,
			Context:      context.Background(),
			// Because Close() is called after every read, the final Close() is a
			// nop for this type, and can be left nil
		}

		return wrappedHandle, wrappedHandle.storageReader.Attrs.Size, nil
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

type ReaderAtCloser interface {
	io.Reader
	io.ReaderAt
	io.Closer
}

// Decorates a Google Storage object handle with ReadAt
type GSReaderAtCloser struct {
	*storage.ObjectHandle
	Context       context.Context
	Closer        *func() error
	storageReader *storage.Reader
}

// Read satisfies io.Reader.
func (o *GSReaderAtCloser) Read(p []byte) (int, error) {
	if o.storageReader == nil {
		rdr, err := o.NewReader(o.Context)
		if err != nil {
			return 0, pfx.Err(err)
		}
		o.storageReader = rdr
	}

	return o.storageReader.Read(p)
}

// ReadAt satisfies io.ReaderAt. Note that this is dependent upon making p a
// buffer of the desired length to be read by NewRangeReader.
func (o *GSReaderAtCloser) ReadAt(p []byte, offset int64) (n int, err error) {
	rdr, err := o.NewRangeReader(o.Context, offset, int64(len(p)))
	if err != nil {
		return 0, err
	}
	defer rdr.Close()

	return rdr.Read(p)
}

// Satisfies io.Closer. If o.close is not set, this is a nop.
func (o *GSReaderAtCloser) Close() error {
	var err error

	if o.storageReader != nil {
		err = o.storageReader.Close()
		// TODO: Make sure we can actually use this error. Currently it will
		// always be overwritten.
	}

	if o.Closer != nil {
		err = (*o.Closer)()
	}

	return err
}
