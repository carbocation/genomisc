package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/carbocation/genomisc"
	"google.golang.org/api/iterator"
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

var (
	BufferSize = 4096 * 8
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
)

func main() {
	defer STDOUT.Flush()

	var (
		acknowledge bool
		phenoPath   string
		headers     = make(map[int]Column)
		BQ          = &WrappedBigQuery{}
	)
	flag.StringVar(&BQ.Project, "project", "broad-ml4cvd", "Name of the Google Cloud project that hosts your BigQuery database instance")
	flag.StringVar(&BQ.Database, "bigquery", "ukbb7089_201903", "BigQuery phenotype database name")
	flag.StringVar(&phenoPath, "pheno", "", "phenotype file for the UKBB")
	flag.BoolVar(&acknowledge, "ack", false, "Acknowledge the limitations of the tool")
	flag.Parse()

	if phenoPath == "" {
		flag.PrintDefaults()
		log.Fatalln()
	}

	if !acknowledge {
		fmt.Fprintln(os.Stderr, "!! -- !!")
		fmt.Fprintln(os.Stderr, "NOTE")
		fmt.Fprintln(os.Stderr, "!! -- !!")
		fmt.Fprintf(os.Stderr, `This tool checks the %s:%s.phenotype table to avoid duplicating data. 
However, it only looks at data that is already present in the BigQuery table. 
Any data that has been emitted to disk but not yet loaded into BigQuery cannot be seen by this tool.
!! -- !!
Please re-run this tool with the --ack flag to demonstrate that you understand this limitation.%s`, BQ.Project, BQ.Database, "\n")
		os.Exit(1)
	}

	phenoPath = genomisc.ExpandHome(phenoPath)
	log.Printf("Converting %s\n", phenoPath)

	// phenotype file

	f, err := os.Open(phenoPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	delim := genomisc.DetermineDelimiter(f)

	f.Seek(0, 0)

	// Buffered reader
	br := bufio.NewReaderSize(f, BufferSize)
	fileCSV := csv.NewReader(br)
	fileCSV.Comma = delim

	log.Printf("Determined phenotype delimiter to be \"%s\"\n", string(delim))

	// Map out known FieldIDs so we don't add duplicate phenotype records
	// connect to BigQuery
	BQ.Context = context.Background()
	BQ.Client, err = bigquery.NewClient(BQ.Context, BQ.Project)
	if err != nil {
		log.Fatalln("Connecting to BigQuery:", err)
	}
	defer BQ.Client.Close()
	knownFieldIDs, err := FindExistingFields(BQ)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Found", len(knownFieldIDs), "FieldIDs already in the database, will ignore those FieldIDs in the new data, if present.")

	// Map the headers
	headRow, err := fileCSV.Read()
	if err != nil {
		log.Fatalln(err)
	}
	if err := parseHeaders(BQ, headRow, headers); err != nil {
		log.Fatalln(err)
	}

	// Track which fields are new
	newFieldIDs := make(map[string]struct{})
	for col := range headRow {
		if col == 0 {
			continue
		}

		if _, exists := knownFieldIDs[headers[col].FieldID]; !exists {
			// This FieldID is new to us -- store it in the DB later
			newFieldIDs[headers[col].FieldID] = struct{}{}
		}
	}

	knownSlice := make([]int, 0, len(knownFieldIDs))
	for v := range knownFieldIDs {
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

	fmt.Fprintf(os.Stderr, "Already known fields that will *not* be added: %v\n", knownSlice)
	fmt.Fprintf(os.Stderr, "New fields that *will* be added: %v\n", newSlice)
	fmt.Fprintf(os.Stderr, "Based on an analysis of the header, %d fields are already present and will be ignored, while %d fields will be added.\n", len(knownFieldIDs), len(newFieldIDs))

	fmt.Printf("sample_id\tFieldID\tinstance\tarray_idx\tvalue\tcoding_file_id\n")

	var addedPhenoCount int
	sample := 1
	for ; ; sample++ {
		// Log to screen more frequently during profiling.
		if sample%1e10 == 0 {
			log.Println("Inserting sample", sample)
		}
		if sample%10000 == 0 {
			log.Println("Saw sample", sample)
		}

		row, err := fileCSV.Read()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln(err)
		}

		// Handle each sample

		sampleID := row[0]
		pheno := SamplePheno{}
		for col := range row {
			if col == 0 {
				// Skip the sample ID
				continue
			}

			pheno.Column = headers[col]
			pheno.SampleID = sampleID
			pheno.Value = row[col]

			if _, exists := knownFieldIDs[pheno.FieldID]; exists {
				// Skip any FieldIDs that were inserted on previous runs. Note
				// that this precludes updates on the database where the update
				// consists of adding new entries, such as a new array_idx to a
				// previously known FieldID.
				continue
			}

			if strings.TrimSpace(pheno.Value) == "" {
				// Skip blanks
				continue
			}

			fmt.Fprintf(STDOUT, "%s\t%s\t%s\t%s\t%s\t%v\n", pheno.SampleID, pheno.FieldID, pheno.Instance, pheno.ArrayIDX, pheno.Value, pheno.CodingFileID)

			addedPhenoCount++
		}
	}

	log.Println("Populated the table with", addedPhenoCount, "new records from", len(newFieldIDs), "previously unseen FieldIDs the phenotype file")
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

type codingLookup struct {
	FieldID      int64              `bigquery:"FieldID"`
	CodingFileID bigquery.NullInt64 `bigquery:"coding_file_id"`
}

func parseHeaders(wbq *WrappedBigQuery, row []string, headers map[int]Column) error {
	for col, header := range row {
		dash := strings.Index(header, "-")
		dot := strings.Index(header, ".")

		// The very first field is just "eid"
		if dash == -1 || dot == -1 {
			if col > 0 {
				return fmt.Errorf("All columns after the 0th are expected to be of format XX-XX.XX, but column %d is %s", col, header)
			}
			headers[col] = Column{
				FieldID: header,
			}
			continue
		}

		// Other fields have all 3 parts
		headers[col] = Column{
			FieldID:  header[:dash],
			Instance: header[dash+1 : dot],
			ArrayIDX: header[dot+1:],
		}
	}

	// Now make it easy to figure out if we have a coding file to map to

	// Get the FieldID => coding_file_id map
	codinglookup := []codingLookup{}
	// if err := db.Select(&codinglookup, "SELECT FieldID, coding_file_id FROM dictionary WHERE coding_file_id IS NOT NULL"); err != nil {
	// 	return err
	// }
	query := wbq.Client.Query(fmt.Sprintf(`SELECT FieldID, coding_file_id
FROM %s.dictionary
WHERE coding_file_id IS NOT NULL`, wbq.Database))
	itr, err := query.Read(wbq.Context)
	if err != nil {
		return err
	}
	for {
		var values codingLookup

		err := itr.Next(&values)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		codinglookup = append(codinglookup, values)
	}

	// Annotate each header column with the CodingFileID that applies to
	// the FieldID in that column.
	fm := fieldMap(headers)
	for _, lookup := range codinglookup {
		relevantHeaders := fm[strconv.FormatInt(lookup.FieldID, 10)]
		for _, v := range relevantHeaders {
			whichHeader := headers[v]
			whichHeader.CodingFileID = lookup.CodingFileID
			headers[v] = whichHeader
		}
	}

	return nil
}

func FindExistingFields(wbq *WrappedBigQuery) (map[string]struct{}, error) {
	// Map out known FieldIDs so we don't add duplicate phenotype records
	knownFieldIDs := make(map[string]struct{})

	query := wbq.Client.Query(fmt.Sprintf(`SELECT DISTINCT p.FieldID
	FROM %s.phenotype p
	WHERE p.FieldID IS NOT NULL
`, wbq.Database))
	itr, err := query.Read(wbq.Context)
	if err != nil && strings.Contains(err.Error(), "Error 404") {
		// Not an error; the table just doesn't exist yet
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	for {
		var values struct {
			FieldID int64 `bigquery:"FieldID"`
		}
		err := itr.Next(&values)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		knownFieldIDs[strconv.FormatInt(values.FieldID, 10)] = struct{}{}
	}

	return knownFieldIDs, nil
}
