package main

import (
	"math/rand"
	"net"
	"net/http"
	"strings"
)

// RandHeteroglyphs produces a string of n symbols which do
// not look like one another. (Derived to be the opposite of
// homoglyphs, which are symbols which look similar to one
// another and cannot be quickly distinguished.)
func RandHeteroglyphs(n int) string {
	var letters = []rune("abcdefghkmnpqrstwxyz")
	lenLetters := len(letters)
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(lenLetters)]
	}
	return string(b)
}

// GetIPAddress returns a user's IP address, even if your Go app is sitting behind
// a reverse proxy.
//From https://groups.google.com/forum/?fromgroups#!topic/golang-nuts/lomWKs0kOfE
func GetIPAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIp := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIp == "" && hdrForwardedFor == "" {
		hostWithoutPort, _, _ := net.SplitHostPort(r.RemoteAddr)
		return hostWithoutPort
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		// TODO: should return first non-local address
		return parts[0]
	}
	return hdrRealIp
}

func MaxInt(x, y int) int {
	if x > y {
		return x
	}

	return y
}
