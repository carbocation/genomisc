package applyprsgcp

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/carbocation/pfx"
)

// ImportBGIFromGoogleStorage copies the BGEN index from google storage to the
// local temp directory. This is necessary because SQLite reads from a filename,
// instead of a reader, and therefore can't be managed over the wire.
func ImportBGIFromGoogleStorage(bgiPath string, client *storage.Client) (memfsBIGPath string, err error) {
	// Inject randomness into the filename so that ImportBGIFromGoogleStorage
	// can be used simultaneously by multiple processes.
	filename := os.TempDir() + "/" + RandHeteroglyphs(20) + "_" + filepath.Base(bgiPath)

	// If we have already copied the file over, don't duplicate work
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		return filename, nil
	}

	// Otherwise, use the global google storage client

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
		return "", pfx.Err(fmt.Sprintf("%v (%s)", err, bgiPath))
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return "", pfx.Err(err)
	}

	err = ioutil.WriteFile(filename, data, 0666)

	return filename, pfx.Err(err)
}

// RandHeteroglyphs produces a string of n symbols which do not look like one
// another. (Derived to be the opposite of homoglyphs, which are symbols which
// look similar to one another and cannot be quickly distinguished.)
func RandHeteroglyphs(n int) string {
	var letters = []rune("abcdefghkmnpqrstwxyz")
	lenLetters := len(letters)
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(lenLetters)]
	}
	return string(b)
}
