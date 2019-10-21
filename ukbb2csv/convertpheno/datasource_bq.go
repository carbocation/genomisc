package convertpheno

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/carbocation/genomisc"
	"github.com/carbocation/pfx"
	"google.golang.org/api/iterator"
)

func RunAllBackendBQ(phenoPaths []string, BQ *WrappedBigQuery) error {
	defer STDOUT.Flush()

	var err error

	// Map out known FieldIDs so we don't add duplicate phenotype records
	// connect to BigQuery
	BQ.Context = context.Background()
	BQ.Client, err = bigquery.NewClient(BQ.Context, BQ.Project)
	if err != nil {
		return fmt.Errorf("connecting to BigQuery: %v", err)
	}
	defer BQ.Client.Close()
	knownFieldIDs, err := findExistingFieldsBQ(BQ)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Found", len(knownFieldIDs), "FieldIDs already in the database, will ignore those FieldIDs in the new data, if present.")

	fmt.Printf("sample_id\tFieldID\tinstance\tarray_idx\tvalue\tcoding_file_id\n")

	// Process each file. KnownFieldIDs is modified within.
	for _, phenoPath := range phenoPaths {
		if err := processOnePathBQ(BQ, knownFieldIDs, phenoPath); err != nil {
			log.Fatalln(err)
		}
	}

	return nil
}

func processOnePathBQ(BQ *WrappedBigQuery, knownFieldIDs map[string]struct{}, phenoPath string) error {
	headers := make(map[int]Column)

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

	// Map the headers
	headRow, err := fileCSV.Read()
	if err != nil {
		log.Fatalln("Header parsing error:", err)
	}
	if err := parseHeadersBQ(BQ, headRow, headers); err != nil {
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

	summarizeImport(knownFieldIDs, newFieldIDs)

	if len(newFieldIDs) < 1 {
		log.Println("No new fields were found in file", phenoPath, " -- skipping")
		return nil
	}

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

	// Make the tool aware of new fields so it won't import them from the next
	// file.
	for key := range newFieldIDs {
		knownFieldIDs[key] = struct{}{}
	}

	return nil
}

func parseHeadersBQ(wbq *WrappedBigQuery, row []string, headers map[int]Column) error {
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

func findExistingFieldsBQ(wbq *WrappedBigQuery) (map[string]struct{}, error) {
	// Map out known FieldIDs so we don't add duplicate phenotype records
	knownFieldIDs := make(map[string]struct{})

	query := wbq.Client.Query(fmt.Sprintf(`SELECT DISTINCT p.FieldID
	FROM %s.phenotype p
	WHERE p.FieldID IS NOT NULL
`, wbq.Database))
	itr, err := query.Read(wbq.Context)
	if err != nil && strings.Contains(err.Error(), "Error 404") {
		// Not an error; the table just doesn't exist yet
		return knownFieldIDs, nil
	} else if err != nil {
		return nil, pfx.Err(err)
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
			return nil, pfx.Err(err)
		}
		knownFieldIDs[strconv.FormatInt(values.FieldID, 10)] = struct{}{}
	}

	return knownFieldIDs, nil
}
