package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/carbocation/pfx"
)

// ImportBGIFromGoogleStorage copies the BGEN index from google storage to the
// local temp directory. This is necessary because SQLite reads from a filename,
// instead of a reader, and therefore can't be managed over the wire.
func ImportBGIFromGoogleStorage(bgiPath string) (memfsBIGPath string, err error) {
	filename := os.TempDir() + "/" + filepath.Base(bgiPath)

	// If we have already copied the file over, don't duplicate work
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		return filename, nil
	}

	// Otherwise, create a google storage client with default settings and use
	// it to fetch the bgi
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return "", pfx.Err(err)
	}

	// Detect the bucket and the path to the actual file
	pathParts := strings.SplitN(strings.TrimPrefix(bgiPath, "gs://"), "/", 2)
	if len(pathParts) != 2 {
		return "", fmt.Errorf("Tried to split your google storage path into 2 parts, but got %d: %v", len(pathParts), pathParts)
	}
	bucketName := pathParts[0]
	pathName := pathParts[1]

	// Open the bucket with default credentials
	bkt := client.Bucket(bucketName)
	handle := bkt.Object(pathName)

	rc, err := handle.NewReader(context.Background())
	if err != nil {
		return "", pfx.Err(err)
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return "", pfx.Err(err)
	}

	err = ioutil.WriteFile(filename, data, 0666)

	return filename, pfx.Err(err)
}
