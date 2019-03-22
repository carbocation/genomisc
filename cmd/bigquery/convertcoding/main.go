package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type WrappedBigQuery struct {
	Context  context.Context
	Client   *bigquery.Client
	Project  string
	Database string
}

var (
	CodingFileRootURL   = `https://biobank.ctsu.ox.ac.uk/crystal/codown.cgi`
	CodingFileDelimiter = '\t'
)

func main() {
	var (
		BQ = &WrappedBigQuery{}
	)
	flag.StringVar(&BQ.Project, "project", "broad-ml4cvd", "Name of the Google Cloud project that hosts your BigQuery database instance")
	flag.StringVar(&BQ.Database, "bigquery", "ukbb7089_201903", "BigQuery phenotype database name")
	flag.StringVar(&CodingFileRootURL, "codingurl", "https://biobank.ctsu.ox.ac.uk/crystal/codown.cgi", "Path to the root URL of the coding files")
	flag.Parse()

	if CodingFileRootURL == "" || BQ.Project == "" || BQ.Database == "" {
		flag.PrintDefaults()
		log.Fatalln()
	}

	log.Println("Using bigquery database", BQ.Database)

	// connect to BigQuery
	var err error
	BQ.Context = context.Background()
	BQ.Client, err = bigquery.NewClient(BQ.Context, BQ.Project)
	if err != nil {
		log.Fatalln("Connecting to BigQuery:", err)
	}
	defer BQ.Client.Close()

	if err := ConvertCoding(BQ); err != nil {
		log.Fatalln(err)
	}

	log.Println("The coding table CSV is now created")
}

func ConvertCoding(wbq *WrappedBigQuery) error {
	out := make([]int64, 0)

	query := wbq.Client.Query(fmt.Sprintf(`SELECT DISTINCT coding_file_id 
	FROM %s.dictionary
	ORDER BY coding_file_id ASC
`, wbq.Database))
	itr, err := query.Read(wbq.Context)
	if err != nil {
		return err
	}
	for {
		var values struct {
			CodingFileID bigquery.NullInt64 `bigquery:"coding_file_id"`
		}
		err := itr.Next(&values)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		if !values.CodingFileID.Valid {
			continue
		}

		out = append(out, values.CodingFileID.Int64)
	}

	log.Println("Populating", len(out), "coding file entries required by the dictionary table")

	var resp *http.Response

	fmt.Printf("coding_file_id\tcoding\tmeaning\tnode\tparent\tselectable\n")

CodingIDLoop:
	for _, id := range out {
		for retries := 0; ; retries++ {
			resp, err = http.Get(fmt.Sprintf("%s?id=%d", CodingFileRootURL, id))
			if err != nil && retries < 5 {
				time.Sleep(5 * time.Second)
				continue
			} else if err != nil && retries >= 5 {
				log.Printf("Error downloading coding file ID %d\n", id)
				return err
			}
			break
		}

		reader := csv.NewReader(resp.Body)
		reader.Comma = CodingFileDelimiter
		reader.LazyQuotes = true
		// TODO: Do we need to eliminate leading space?
		// reader.TrimLeadingSpace = true

		for j := 0; ; j++ {
			row, err := reader.Read()
			if err != nil && err == io.EOF {
				resp.Body.Close()
				break
			} else if err != nil {
				buf := bytes.NewBuffer(nil)
				io.Copy(buf, resp.Body)
				if strings.Contains(buf.String(), "internal error") {
					log.Println("Coding file", id, "is not permitted to be downloaded from the UKBB")
					continue CodingIDLoop
				}
			}
			if j < 1 {
				log.Printf("Converting data from coding file ID#%d\n", id)
				log.Printf("Header (%d elements): %+v\n", len(row), row)
			}

			// don't insert the header
			if j < 1 {
				continue
			}

			if len(row) == 2 {
				fmt.Printf("%d\t%s\t%s\t\t\t\n", id, row[0], row[1])
			} else if len(row) == 5 {
				fmt.Printf("%d\t%s\t%s\t%s\t%s\t%s\n", id, row[0], row[1], row[2], row[3], row[4])
			} else {
				// Probably a bogus number of fields, like line 6 in https://biobank.ctsu.ox.ac.uk/crystal/codown.cgi?id=47
				// First, try to eliminate consecutive blank columns. If that doesn't work, error.
				keep := make([]string, 0)
				for _, col := range row {
					if strings.TrimSpace(col) == "" {
						continue
					}
					keep = append(keep, col)
				}

				log.Printf("%+v\n", row)
				log.Printf("Bad format for coding ID %d since it seems to contain %d columns. Attempting to delete consecutive delimiters.\n", id, len(row))

				if len(keep) == 2 {
					fmt.Printf("%d\t%s\t%s\t\t\t\n", id, keep[0], keep[1])
					continue
				} else if len(keep) == 5 {
					fmt.Printf("%d\t%s\t%s\t%s\t%s\t%s\n", id, keep[0], keep[1], keep[2], keep[3], keep[4])
					continue
				}

				// Failed to rescue
				return fmt.Errorf("Unclear how to process coding ID %d since it seems to contain %d columns", id, len(row))
			}
		}
	}

	return nil
}
