package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func OpenFileOrURL(input string) ([]byte, error) {
	var f io.ReadCloser

	if strings.HasPrefix(input, "http") {
		resp, err := http.Get(input)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		f = resp.Body

		// fileBytes, err := ioutil.ReadAll(resp.Body)
		// if err != nil {
		// 	return nil, err
		// }

		// return fileBytes, nil
	} else {
		file, err := os.Open(input)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		f = file
	}

	return ioutil.ReadAll(f)
}
