// Deletes files from `folder` not matching names found in `permitted-list`,
// with `suffix` optionally being appended to `permitted-list` to facilitate
// matching.
package main

import (
	"compress/gzip"
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var folder, permittedList, suffix string

	flag.StringVar(&folder, "folder", "", "Folder to scan for files and destroy files not found in permitted-list")
	flag.StringVar(&permittedList, "permitted-list", "", "File containing a single column, each of which is a permitted filename. May be gzipped.")
	flag.StringVar(&suffix, "suffix", "", "Suffix to append to the filenames in permitted-list, if needed, to facilitate matching with filenames in folder")
	flag.Parse()

	if folder == "" || permittedList == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := func() error {

		// Populate list of permitted files

		permitted, err := os.Open(permittedList)
		if err != nil {
			return err
		}
		defer permitted.Close()

		var ungzippedReader io.Reader

		if strings.HasSuffix(permittedList, ".gz") {
			ungzippedReader, err = gzip.NewReader(permitted)
			if err != nil {
				return err
			}
		} else {
			ungzippedReader = permitted
		}

		r := csv.NewReader(ungzippedReader)
		allowedSlice, err := r.ReadAll()
		if err != nil {
			return err
		}

		allowed := make(map[string]struct{})
		for _, col := range allowedSlice {
			if len(col) != 1 {
				continue
			}

			// Append the suffix here to simplify existence checking later
			allowed[col[0]+suffix] = struct{}{}
		}

		// Scan the folder and delete any non-permitted files
		if err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			// Not allowed? Delete it
			if _, exists := allowed[info.Name()]; !exists {
				return os.Remove(path)
			}

			return nil
		}); err != nil {
			return err
		}

		return nil

	}(); err != nil {
		log.Fatalln((err))
	}
}
