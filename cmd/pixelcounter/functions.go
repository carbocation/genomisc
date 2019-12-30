package main

import (
	"os"
)

// Via https://flaviocopes.com/go-list-files/
func scanFolder(dirname string) ([]os.FileInfo, error) {

	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}

	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}

	return files, nil
}
