package main

import "testing"

func TestFastSlice(t *testing.T) {
	fs := NewFastSlice()
	input := "Hello there"
	input2 := "Goodbye"
	fs.Add(input)
	fs.Add(input2)
	fs.Add(input)
	fs.Add(input2)
	if x := fs.Add(input); x != 0 {
		t.Errorf("Got wrong position (%d) for %s", x, input)
	}
	if x := fs.Add(input2); x != 1 {
		t.Errorf("Got wrong position (%d) for %s", x, input2)
	}
}
