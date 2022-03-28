package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/carbocation/pfx"
	"google.golang.org/api/iterator"
)

func ExecuteQuery(BQ *WrappedBigQuery, query *bigquery.Query, deathCensorDate, phenoCensorDate bigquery.NullDate) error {
	itr, err := query.Read(BQ.Context)
	if err != nil {
		return pfx.Err(fmt.Sprint(err.Error(), query.Parameters))
	}

	now := time.Now()

	log.Printf("Censoring death at %v and missing phenotypes at %v\n", deathCensorDate, phenoCensorDate)

	// Print header
	fmt.Printf("sample_id\tsex\tethnicity\tbirthdate\tenroll_date\tenroll_age\tenroll_age_days\tdeath_date\tdeath_age\tdeath_age_days\tdeath_censor_date\tdeath_censor_age\tdeath_censor_age_days\tphenotype_censor_date\tphenotype_censor_age\tphenotype_censor_age_days\tlost_to_followup_date\tlost_to_followup_age\tlost_to_followup_age_days\tcomputed_date\tmissing_fields\thas_gp_data\n")

	counts := &counters{}
	for {
		var v CensorResult
		err := itr.Next(&v)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return pfx.Err(err)
		}

		// Attach the censoring data
		v.deathCensored = deathCensorDate
		v.phenoCensored = phenoCensorDate

		// Count before any exclusions
		counts.CountValidFields(v)

		// Some samples have been removed
		if v.SampleID <= 0 {
			log.Printf("Removing sample %d\n", v.SampleID)
			continue
		}

		if !v.Enrolled.Valid {
			log.Printf("No valid enrollment time for sample %d. Estimated enrollment was %s. Skipping.\n", v.SampleID, v.Enrolled.Date)
			continue
		}

		if !v.Born().Valid {
			log.Printf("No valid birthdate for sample %d. Estimated birthdate was %s. Skipping.\n", v.SampleID, v.Born())
			continue
		}

		// Nonfatal but note missingness
		if !v.Sex.Valid {
			v.Missing = append(v.Missing, "sex")
		}
		if !v.Ethnicity.Valid {
			v.Missing = append(v.Missing, "ethnicity")
		}
		if !v.BornMonth.Valid {
			v.Missing = append(v.Missing, "birth_month")
		}

		fmt.Printf("%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\n",
			v.SampleID,
			NA(v.Sex),
			NA(v.Ethnicity),
			NA(v.Born()),
			NA(v.Enrolled),
			TimesToFractionalYears(v.Born(), v.Enrolled),
			TimesToDays(v.Born(), v.Enrolled),
			NA(v.Died),
			TimesToFractionalYears(v.Born(), v.Died),
			TimesToDays(v.Born(), v.Died),
			v.DeathCensored(),
			TimesToFractionalYears(v.Born(), v.DeathCensored()),
			TimesToDays(v.Born(), v.DeathCensored()),
			v.PhenoCensored(),
			TimesToFractionalYears(v.Born(), v.PhenoCensored()),
			TimesToDays(v.Born(), v.PhenoCensored()),
			NA(v.Lost),
			TimesToFractionalYears(v.Born(), v.Lost),
			TimesToDays(v.Born(), v.Lost),
			now.Format("2006-01-02"),
			v.MissingToString(),
			BoolToInt(v.HasPrimaryCareData),
		)
	}

	if v := counts.String(); v != "" {
		fmt.Fprintf(os.Stderr, "%s", v)
	}

	return nil
}

// We'll define 'NA' to be our indicator for a null string for BigQuery
func NA(input fmt.Stringer) string {
	invalid := NullMarker

	switch v := input.(type) {
	case bigquery.NullInt64:
		if !v.Valid {
			return invalid
		}
	case bigquery.NullFloat64:
		if !v.Valid {
			return invalid
		}
	case bigquery.NullString:
		if !v.Valid {
			return invalid
		}
	case bigquery.NullDate:
		if !v.Valid {
			return invalid
		}
	}

	return input.String()
}
