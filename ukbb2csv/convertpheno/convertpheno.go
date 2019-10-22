package convertpheno

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"cloud.google.com/go/bigquery"
)

var (
	BufferSize = 4096 * 8
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
)

const (
	// pheno file field columns
	eid = iota
)

type SamplePheno struct {
	Column
	SampleID string `db:"sample_id"`
	Value    string `db:"value"`
}

type Column struct {
	FieldID      string             `db:"FieldID"`
	Instance     string             `db:"instance"`
	ArrayIDX     string             `db:"array_idx"`
	CodingFileID bigquery.NullInt64 `db:"coding_file_id"`
}

type WrappedBigQuery struct {
	Context  context.Context
	Client   *bigquery.Client
	Project  string
	Database string
}

type CodingFileLookup struct {
	FieldID      int64              `bigquery:"FieldID"`
	CodingFileID bigquery.NullInt64 `bigquery:"coding_file_id"`
}

func summarizeImport(knownFieldIDs, newFieldIDs map[string]struct{}) {
	knownSlice := make([]int, 0, len(knownFieldIDs))
	for v := range knownFieldIDs {
		if _, exists := newFieldIDs[v]; !exists {
			// If we aren't trying to add this field, ignore
			continue
		}

		if intVal, err := strconv.Atoi(v); err == nil {
			knownSlice = append(knownSlice, intVal)
		}
	}
	sort.IntSlice(knownSlice).Sort()

	newSlice := make([]int, 0, len(newFieldIDs))
	for v := range newFieldIDs {
		if intVal, err := strconv.Atoi(v); err == nil {
			newSlice = append(newSlice, intVal)
		}
	}
	sort.IntSlice(newSlice).Sort()

	fmt.Fprintf(os.Stderr, "Total known fields: %v. %v fields in the file are duplicate and will *not* be added: %v\n", len(knownFieldIDs), len(knownSlice), splitToString(knownSlice, ","))
	fmt.Fprintf(os.Stderr, "%d fields are new and *will* be added: %v\n", len(newFieldIDs), splitToString(newSlice, ","))
}

// Yields a map that notes which columns are associated with which field
func fieldMap(headers map[int]Column) map[string][]int {
	// Map FieldID => ith, jth, kth, (etc) Column(s) in the CSV
	out := make(map[string][]int)

	for i, v := range headers {
		out[v.FieldID] = append(out[v.FieldID], i)
	}

	return out
}

// Via https://stackoverflow.com/a/42159097/199475
func splitToString(a []int, sep string) string {
	if len(a) == 0 {
		return ""
	}

	b := make([]string, len(a))
	for i, v := range a {
		b[i] = strconv.Itoa(v)
	}
	return strings.Join(b, sep)
}
