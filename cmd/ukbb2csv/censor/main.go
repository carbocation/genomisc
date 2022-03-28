package main

import (
	"context"
	"flag"
	"log"
	"os"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	_ "github.com/carbocation/genomisc/compileinfoprint"
)

const NullMarker = "NA"

type SamplePheno struct {
	SampleID     int64              `bigquery:"sample_id"`
	Value        string             `bigquery:"value"`
	FieldID      int64              `bigquery:"FieldID"`
	Instance     int64              `bigquery:"instance"`
	ArrayIDX     int64              `bigquery:"array_idx"`
	CodingFileID bigquery.NullInt64 `bigquery:"coding_file_id"`
}

type WrappedBigQuery struct {
	Context  context.Context
	Client   *bigquery.Client
	Project  string
	Database string
}

func main() {
	var (
		phenoCensorDateString string
		deathCensorDateString string
		usePhenoTableDeath    bool
		displayQuery          bool
		BQ                    = &WrappedBigQuery{}
	)

	flag.StringVar(&phenoCensorDateString, "pheno_censor", "", "With format YYYY-MM-DD, please provide the Hospital Data censor date from https://biobank.ctsu.ox.ac.uk/crystal/exinfo.cgi?src=Data_providers_and_dates")
	flag.StringVar(&deathCensorDateString, "death_censor", "", "With format YYYY-MM-DD, please provide the Death censor date from https://biobank.ctsu.ox.ac.uk/crystal/exinfo.cgi?src=Data_providers_and_dates")
	flag.StringVar(&BQ.Project, "project", "", "Name of the Google Cloud project that hosts your BigQuery database instance")
	flag.BoolVar(&usePhenoTableDeath, "phenodeath", false, "Use phenotype table instead of the downloaded death table to define death? (Optional; default: use death table.)")
	flag.BoolVar(&displayQuery, "display_query", false, "Display the query that will be run?")
	flag.StringVar(&BQ.Database, "bigquery", "", "BigQuery source database name")
	flag.Parse()

	if phenoCensorDateString == "" || deathCensorDateString == "" || BQ.Project == "" || BQ.Database == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Println("Using bigquery database", BQ.Database)
	log.Println("Output uses", NullMarker, "in place of null values. Please specify this when loading data into bigquery.")

	log.Println("Producing censoring table")
	// if err := Censor(BQ, deathCensorDateString, phenoCensorDateString, usePhenoTableDeath); err != nil {
	// 	log.Fatalln("Fatal error:", err)
	// }

	// parse censoring dates
	dcd, err := civil.ParseDate(deathCensorDateString)
	if err != nil {
		log.Fatalln(err)
	}
	deathCensorDate := bigquery.NullDate{
		Date:  dcd,
		Valid: true,
	}
	pcd, err := civil.ParseDate(phenoCensorDateString)
	if err != nil {
		log.Fatalln(err)
	}
	phenoCensorDate := bigquery.NullDate{
		Date:  pcd,
		Valid: true,
	}

	// connect to BigQuery
	BQ.Context = context.Background()
	BQ.Client, err = bigquery.NewClient(BQ.Context, BQ.Project)
	if err != nil {
		log.Fatalf("Connecting to BigQuery: %s\n", err)
	}
	defer BQ.Client.Close()

	// build and execute the query
	query, err := BuildQuery(BQ, displayQuery)
	if err != nil {
		log.Fatalln(err)
	}
	if err := ExecuteQuery(BQ, query, deathCensorDate, phenoCensorDate); err != nil {
		log.Fatalln(err)
	}
}
