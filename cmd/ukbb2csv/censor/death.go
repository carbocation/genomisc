package main

import (
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type deathDate struct {
	SampleID  int64             `bigquery:"sample_id"`
	DeathDate bigquery.NullDate `bigquery:"date_of_death"`
}

func BigQueryDeath(wbq *WrappedBigQuery) (map[int64]deathDate, error) {
	out := make(map[int64]deathDate)

	query := wbq.Client.Query(fmt.Sprintf(`
WITH all_samples AS (
	SELECT 
		sample_id
	FROM %s.phenotype
	GROUP BY sample_id
), death AS (
	SELECT 
	eid sample_id,
	MAX(date_of_death) date_of_death
	FROM %s.death
	GROUP BY sample_id
)
	
SELECT 
	sample_id,
	death.date_of_death
FROM all_samples
JOIN death USING(sample_id)
ORDER BY date_of_death DESC NULLS FIRST
	
`, wbq.Database, wbq.Database))

	itr, err := query.Read(wbq.Context)
	if err != nil {
		return nil, err
	}
	for {
		var values deathDate
		err := itr.Next(&values)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		// Take only the first, since we use this for things like enrollment
		// date. If someone came to a follow-up visit, we don't want to say that
		// they "enrolled" at the time of their follow-up, for example. Relies
		// on sort order specified above in the query.
		if _, exists := out[values.SampleID]; exists {
			continue
		}
		out[values.SampleID] = values
	}

	return out, nil
}

func (BQ *WrappedBigQuery) AddDiedDate(out map[int64]CensorResult) (int, error) {

	timeZone, err := time.LoadLocation("UTC")
	if err != nil {
		return 0, err
	}

	// Died
	res, err := BigQueryDeath(BQ)
	if err != nil {
		return 0, err
	}
	N := len(res)

	for k, v := range res {
		entry := out[k]

		// No need to update - already will be null
		if !v.DeathDate.Valid {
			continue
		}

		// Extract the year, month, and day from the null.Date object returned
		// by BigQuery

		// Previous approach failed to trigger entry.died.Valid to become true
		// entry.died.Time = time.Date(v.DeathDate.Date.Year, v.DeathDate.Date.Month, v.DeathDate.Date.Day, 0, 0, 0, 0, timeZone)

		entry.died.Scan(time.Date(v.DeathDate.Date.Year, v.DeathDate.Date.Month, v.DeathDate.Date.Day, 0, 0, 0, 0, timeZone))

		out[k] = entry
	}

	return N, nil
}

func (BQ *WrappedBigQuery) AddDiedDatePhenoTable(out map[int64]CensorResult) (int, error) {

	// Died
	res, err := BigQuerySingleFieldFirst(BQ, 40000)
	if err != nil {
		return 0, err
	}
	N := len(res)

	for k, v := range res {
		entry := out[k]

		diedDate, err := time.Parse("2006-01-02", v)
		if err != nil {
			log.Println("Died date parsing issue for", entry)
		}
		if err := entry.died.Scan(diedDate); err != nil {
			log.Println("Died date parsing issue for", entry)
		}

		out[k] = entry
	}

	return N, nil
}
