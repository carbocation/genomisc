package main

import (
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
)

type CensorResult struct {
	SampleID  int64               `bigquery:"sample_id"`
	BornYear  bigquery.NullString `bigquery:"born_year"`
	BornMonth bigquery.NullString `bigquery:"born_month"`
	Enrolled  bigquery.NullDate
	Died      bigquery.NullDate
	Lost      bigquery.NullDate
	Sex       bigquery.NullString `bigquery:"sex"`
	Ethnicity bigquery.NullString `bigquery:"ethnicity"`

	// Flag since only ~50% of the UKBB has primary care data so this is a heavily biased sample
	HasPrimaryCareData bool `bigquery:"has_primary_care_data"`

	// We set these manually
	phenoCensored bigquery.NullDate
	deathCensored bigquery.NullDate
	computed      bigquery.NullDate // Date this was computed

	// In the database, populate with a list of fields that we would like to
	// have (e.g., month of birth, lost to followup) but which were not present,
	// so we know if the table was constructed from incomplete data
	Missing []string
}

// IsValid checks to make sure that the sample ID is > 0, that there is an
// enrollment date, and that we can assemble a valid birthdate. If not, then the
// sample is deemed invalid.
func (s CensorResult) IsValid() bool {
	return s.SampleID > 0 && s.Enrolled.Valid && s.Born().Valid
}

// Since birthdate is considered a sensitive value, we reconstruct an estimated
// birthdate from year and, if available, the month of birth.
func (s CensorResult) Born() bigquery.NullDate {
	if !s.BornYear.Valid {
		return bigquery.NullDate{}
	}

	// If we know year + month, then neutral assumption is that birthday is on
	// the middle day of the month. If we just know year, then assumption is
	// being born midway through the year (July 2).
	day := "15"

	month := s.BornMonth.StringVal
	if month == "" || !s.BornMonth.Valid {
		month = "7"
		day = "02"
	}

	dt, err := civil.ParseDate(fmt.Sprintf("%04s-%02s-%02s", s.BornYear.StringVal, month, day))
	if err != nil {
		return bigquery.NullDate{}
	}

	return bigquery.NullDate{
		Date:  dt,
		Valid: true,
	}
}

// Return date of death, or censoring date. TODO: If s.deathCensored comes
// before s.Died, even if s.Died is valid, do we want to censor?
func (s CensorResult) DeathCensored() bigquery.NullDate {
	if s.Died.Valid {
		return s.Died
	}

	if s.Lost.Valid {
		return s.Lost
	}

	return s.deathCensored
}

// Return date of phenotype censoring. TODO: If s.phenoCensored comes before
// s.Died or s.Lost, even if s.Died or s.Lost are valid, do we want to censor?
func (s CensorResult) PhenoCensored() bigquery.NullDate {
	if s.Died.Valid {
		return s.Died
	}

	if s.Lost.Valid {
		return s.Lost
	}

	return s.phenoCensored
}

func (s CensorResult) MissingToString() string {
	if res := strings.Join(s.Missing, "|"); len(res) > 0 {
		return res
	}

	return "NA"
}
