package main

import (
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/carbocation/pfx"
	"google.golang.org/api/iterator"
)

type Result struct {
	SampleID            int64              `bigquery:"sample_id"`
	IncidentNumber      bigquery.NullInt64 `bigquery:"incident_number"`
	StatusStart         StatusEnum         `bigquery:"status_start"`
	StatusEnd           StatusEnum         `bigquery:"status_end"`
	StartDate           bigquery.NullDate  `bigquery:"start_date"`
	EndDate             bigquery.NullDate  `bigquery:"end_date"`
	StartAgeDays        bigquery.NullInt64 `bigquery:"start_age_days"`
	EndAgeDays          bigquery.NullInt64 `bigquery:"end_age_days"`
	SurvivalDays        bigquery.NullInt64 `bigquery:"days_since_start_date"`
	DaysSinceEnrollDate bigquery.NullInt64 `bigquery:"days_since_enroll_date"`
	IsFinalRecord       bigquery.NullBool  `bigquery:"is_final_record"`
}

func ExecuteQuery(BQ *WrappedBigQuery, query *bigquery.Query, diseaseName string, missingFields []string) error {
	defer STDOUT.Flush()

	itr, err := query.Read(BQ.Context)
	if err != nil {
		return pfx.Err(fmt.Sprint(err.Error(), query.Parameters))
	}
	todayDate := time.Now().Format("2006-01-02")
	missing := strings.Join(missingFields, ",")
	fmt.Fprintf(STDOUT, "disease\tsample_id\tincident_number\tstatus_start\tstatus_end\tis_final\tstatus_start_int\tstatus_end_int\tstart_date\tend_date\tstart_age_days\tend_age_days\tsurvival_days\tstatus_start_raw\tstatus_end_raw\tdays_since_enroll_date\tcomputed_date\tmissing_fields\n")
	for {
		var r Result
		err := itr.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return pfx.Err(err)
		}

		fmt.Fprintf(STDOUT, "%s\t%d\t%s\t%s\t%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			diseaseName,
			r.SampleID,
			NA(r.IncidentNumber),
			r.StatusStart.Simplify().String(),
			r.StatusEnd.Simplify().String(),
			NA(r.IsFinalRecord),
			int(r.StatusStart.Simplify()),
			int(r.StatusEnd.Simplify()),
			NA(r.StartDate),
			NA(r.EndDate),
			NA(r.StartAgeDays),
			NA(r.EndAgeDays),
			NA(r.SurvivalDays),
			r.StatusStart.String(),
			r.StatusEnd.String(),
			NA(r.DaysSinceEnrollDate),
			todayDate,
			missing,
		)
	}

	return nil
}

// NA emits an empty string instead of "NULL" since this plays better with
// BigQuery
func NA(input interface{}) interface{} {
	invalid := ""

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
	case bigquery.NullBool:
		if !v.Valid {
			return invalid
		}
	}

	return input
}
