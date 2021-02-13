package main

import (
	"fmt"

	"github.com/minio/blake2b-simd"
)

type fastSlice struct {
	m map[string]int
	s []string
}

func NewFastSlice() *fastSlice {
	fs := &fastSlice{
		m: make(map[string]int),
		s: make([]string, 0),
	}

	return fs
}

func (fs *fastSlice) Slice() []string {
	return fs.s
}

func (fs *fastSlice) Add(input string) (position int) {
	h := fs.BlakeHash(input)

	if pos, exists := fs.m[h]; exists {
		return pos
	}

	position = len(fs.s)
	fs.s = append(fs.s, input)
	fs.m[h] = position

	return position
}

func (fs *fastSlice) BlakeHash(input string) (hash string) {
	h, err := blake2b.New(&blake2b.Config{Size: 32})
	if err != nil {
		return ""
	}
	if _, err := h.Write([]byte(input)); err != nil {
		return ""
	}

	return fmt.Sprintf("%X", h.Sum(nil))
}
