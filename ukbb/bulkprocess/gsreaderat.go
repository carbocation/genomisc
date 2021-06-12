package bulkprocess

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/carbocation/genomisc"
	"github.com/carbocation/pfx"
	"google.golang.org/api/iterator"
)

// Deprecated: Use github.com/carbocation/genomisc.ReaderAtCloser instead
type ReaderAtCloser = genomisc.ReaderAtCloser

// Deprecated: Use github.com/carbocation/genomisc.GSReaderAtCloser instead
type GSReaderAtCloser = genomisc.GSReaderAtCloser

func ListFromGoogleStorage(path string, client *storage.Client) ([]string, error) {

	output := make([]string, 0)

	// Detect the bucket and the path to the actual file
	pathParts := strings.SplitN(strings.TrimPrefix(path, "gs://"), "/", 2)
	if len(pathParts) != 2 {
		return nil, fmt.Errorf("Tried to split your google storage path into 2 parts, but got %d: %v", len(pathParts), pathParts)
	}
	bucketName := pathParts[0]
	pathName := pathParts[1]

	// Open the bucket with default credentials
	bkt := client.Bucket(bucketName)
	query := &storage.Query{
		Prefix: pathName,
	}
	query.SetAttrSelection([]string{"Name"})

	objectIterator := bkt.Objects(context.Background(), query)
	for {
		attrs, err := objectIterator.Next()
		if err != nil && err != iterator.Done {
			return nil, err
		} else if err == iterator.Done {
			break
		}

		// filename := strings.TrimSuffix(path, "/") + "/" + filepath.Base(attrs.Name)
		filename := filepath.Base(attrs.Name)

		output = append(output, filename)
	}

	return output, nil

}

func MaybeOpenFromGoogleStorage(path string, client *storage.Client) (ReaderAtCloser, int64, error) {
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

		wrappedHandle := &GSReaderAtCloser{
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
