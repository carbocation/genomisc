package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
)

// FetchImageFromContainer will attempt to fetch an image that is inside a
// .tar.gz (if specified) or a raw image (if no container is specified)
func (h *handler) FetchImageFromContainer(entry ManifestEntry, showOverlay bool) (io.ReadCloser, string, error) {

	imageFilename := entry.Dicom
	rootPath := h.Global.MergedRoot
	if !showOverlay {
		imageFilename = MergedNameToRawName(entry.Dicom, addSuffix, removeSuffix)
		rootPath = h.Global.RawRoot
	}

	// If we're explicitly not nested, then we're not in a container - look for
	// the image directly.
	if !h.manifest.Nested {
		imReader, _, err := bulkprocess.MaybeOpenFromGoogleStorage(rootPath+"/"+imageFilename, h.Global.storageClient)
		if err != nil {
			return nil, imageFilename, err
		}
		defer imReader.Close()
		imageBytes, err := ioutil.ReadAll(imReader)
		if err != nil {
			return nil, imageFilename, err
		}

		return io.NopCloser(bytes.NewReader(imageBytes)), imageFilename, nil
	}

	// Open the .tar.gz
	containerName := entry.Zip
	imReader, _, err := bulkprocess.MaybeOpenFromGoogleStorage(rootPath+"/"+containerName, h.Global.storageClient)
	if err != nil {
		return nil, imageFilename, err
	}
	defer imReader.Close()

	gzr, err := gzip.NewReader(imReader)
	if err != nil {
		return nil, imageFilename, err
	}

	tarReader := tar.NewReader(gzr)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, imageFilename, err
		}

		if header.Name == imageFilename && header.Typeflag == tar.TypeReg {
			imageBytes, err := ioutil.ReadAll(tarReader)
			if err != nil {
				return nil, imageFilename, err
			}

			return io.NopCloser(bytes.NewReader(imageBytes)), imageFilename, nil
		}
	}

	return nil, imageFilename, fmt.Errorf("image %s not found in container %s under path %s", entry.Dicom, entry.Zip, rootPath)
}
