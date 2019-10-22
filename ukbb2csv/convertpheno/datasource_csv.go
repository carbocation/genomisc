package convertpheno

import (
	"bufio"
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
)

const (
	DictionaryFieldIDCol      = 2
	DictionaryCodingFileIDCol = 14
)

func RunAllBackendCSV(phenoPaths []string, dictionaryCSV string) error {
	defer STDOUT.Flush()

	fmt.Printf("sample_id\tFieldID\tinstance\tarray_idx\tvalue\tcoding_file_id\n")

	knownFieldIDs := make(map[string]struct{})

	// Process each file. KnownFieldIDs is modified within.
	for _, phenoPath := range phenoPaths {
		if err := processOnePathCSV(dictionaryCSV, knownFieldIDs, phenoPath); err != nil {
			return pfx.Err(err)
		}
	}

	return nil
}

func processOnePathCSV(dictionaryCSV string, knownFieldIDs map[string]struct{}, phenoPath string) error {
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
		return pfx.Err(fmt.Errorf("Header parsing error: %v", err))
	}
	if err := parseHeadersCSV(dictionaryCSV, headRow, headers); err != nil {
		return pfx.Err(err)
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
			return pfx.Err(err)
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

func parseHeadersCSV(dictionaryCSV string, row []string, headers map[int]Column) error {
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

	// Parse dictionary CSV
	dict, err := os.Open(dictionaryCSV)
	if err != nil {
		return pfx.Err(err)
	}

	dictReader := csv.NewReader(dict)
	dictReader.Comma = '\t'
	dictReader.LazyQuotes = true

	dictRecords, err := dictReader.ReadAll()
	if err != nil {
		return pfx.Err(err)
	}

	// Now make it easy to figure out if we have a coding file to map to

	// Get the FieldID => coding_file_id map
	codinglookup := []CodingFileLookup{}

	for _, rec := range dictRecords {
		FieldID, err := strconv.ParseInt(rec[DictionaryFieldIDCol], 10, 64)
		if err != nil {
			continue
		}

		CodingFileID, err := strconv.ParseInt(rec[DictionaryCodingFileIDCol], 10, 64)
		if err != nil {
			continue
		}

		codinglookup = append(codinglookup, CodingFileLookup{
			FieldID:      FieldID,
			CodingFileID: bigquery.NullInt64{Int64: CodingFileID},
		})
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
