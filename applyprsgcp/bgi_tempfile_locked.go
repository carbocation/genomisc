package applyprsgcp

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/carbocation/pfx"
)

var importBGIFromGoogleStorageLockedMutex = &sync.RWMutex{}
var importBGIFromGoogleStorageLookup = make(map[string]struct{})

// ImportBGIFromGoogleStorageLocked copies the BGEN index from google storage to
// the local temp directory. This is necessary because SQLite reads from a
// filename, instead of a reader, and therefore can't be managed over the wire.
func ImportBGIFromGoogleStorageLocked(bgiPath string, client *storage.Client) (memfsBIGPath string, newDownload bool, err error) {

	// The local filename will be:
	filename := os.TempDir() + "/" + filepath.Base(bgiPath)

	// Lock upgrading is racy. See https://github.com/golang/go/issues/4026#issuecomment-66069820

	// First step is to gain an RLock
	importBGIFromGoogleStorageLockedMutex.RLock()

	// Now we can ask if the file exists
	_, exists := importBGIFromGoogleStorageLookup[bgiPath]
	// Unlock regardless
	importBGIFromGoogleStorageLockedMutex.RUnlock()
	if exists {
		return filename, false, nil
	}

	// The file doesn't exist. It is no longer rlocked. Let's upgrade to a full
	// lock so we can fetch it in this goroutine.
	importBGIFromGoogleStorageLockedMutex.Lock()

	// Since it may take time to gain the Lock, check again: does the file
	// exist?
	_, exists = importBGIFromGoogleStorageLookup[bgiPath]
	if exists {
		importBGIFromGoogleStorageLockedMutex.Unlock()
		return filename, false, nil
	}

	// We now have the ball, and we know the file does not exist. Mutate.
	err = importBGIFromGoogleStorageLocked(bgiPath, filename, client)

	// Signify that we have the file now.
	importBGIFromGoogleStorageLookup[bgiPath] = struct{}{}

	// Signify that we are done and release the lock
	// delete(importBGIFromGoogleStorageLockedActive, bgiPath)
	importBGIFromGoogleStorageLockedMutex.Unlock()

	return filename, true, err
}

func importBGIFromGoogleStorageLocked(bgiPath, filename string, client *storage.Client) (err error) {
	// Detect the bucket and the path to the actual file
	pathParts := strings.SplitN(strings.TrimPrefix(bgiPath, "gs://"), "/", 2)
	if len(pathParts) != 2 {
		return fmt.Errorf("tried to split your google storage path into 2 parts, but got %d: %v", len(pathParts), pathParts)
	}
	bucketName := pathParts[0]
	pathName := pathParts[1]

	// Open the bucket with default credentials
	bkt := client.Bucket(bucketName)
	handle := bkt.Object(pathName)

	rc, err := handle.NewReader(context.Background())
	if err != nil {
		return pfx.Err(fmt.Sprintf("%v (%s)", err, bgiPath))
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return pfx.Err(err)
	}

	err = ioutil.WriteFile(filename, data, 0666)

	return pfx.Err(err)
}
