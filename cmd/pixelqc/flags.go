package main

import (
	"sort"
	"strings"
)

// map[sampleID] => set of flags
type SampleFlags map[string]flagSet

func (s SampleFlags) AddFlag(sample, flag string) {
	samp, exists := s[sample]
	if !exists {
		samp = make(flagSet)
	}
	samp[flag] = struct{}{}
	s[sample] = samp
}

type flagSet map[string]struct{}

func (fs flagSet) String() string {
	if len(fs) == 0 {
		return ""
	}

	sb := make([]string, 0, len(fs))
	for v := range fs {
		sb = append(sb, v)
	}

	sort.Strings(sb)

	return strings.Join(sb, "|")
}
