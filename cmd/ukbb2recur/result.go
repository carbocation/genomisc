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
	SampleID                    int64              `bigquery:"sample_id"`
	IncidentNumber              bigquery.NullInt64 `bigquery:"incident_number"`
	StatusStart                 StatusEnum         `bigquery:"status_start"`
	StatusEnd                   StatusEnum         `bigquery:"status_end"`
	StartDate                   bigquery.NullDate  `bigquery:"start_date"`
	EndDate                     bigquery.NullDate  `bigquery:"end_date"`
	BirthDate                   bigquery.NullDate  `bigquery:"birthdate"`
	StartAgeDays                bigquery.NullInt64 `bigquery:"start_age_days"`
	EndAgeDays                  bigquery.NullInt64 `bigquery:"end_age_days"`
	SurvivalDays                bigquery.NullInt64 `bigquery:"days_since_start_date"`
	EnrollDate                  bigquery.NullDate  `bigquery:"enroll_date"`
	DaysSinceEnrollDate         bigquery.NullInt64 `bigquery:"days_since_enroll_date"`
	IsFinalRecord               bigquery.NullBool  `bigquery:"is_final_record"`
	FirstEventDate              bigquery.NullDate  `bigquery:"first_event_date"`
	FirstEventAgeDays           bigquery.NullInt64 `bigquery:"first_event_age_days"`
	DaysSinceFirstEventDate     bigquery.NullInt64 `bigquery:"days_since_first_event_date"`
	StartDaysSinceLastEventDate int                // Will always be 0 unless using time-varying covariates
}

func ExecuteQuery(BQ *WrappedBigQuery, query *bigquery.Query, diseaseName string, missingFields []string, timeVaryingDays, timeVaryingDaysAfterEvent int) error {
	defer STDOUT.Flush()

	itr, err := query.Read(BQ.Context)
	if err != nil {
		return pfx.Err(fmt.Sprint(err.Error(), query.Parameters))
	}
	todayDate := time.Now().Format("2006-01-02")
	missing := strings.Join(missingFields, ",")
	fmt.Fprintf(STDOUT, "disease\tsample_id\tincident_number\tstatus_start\tstatus_end\tis_final\tstatus_start_int\tstatus_end_int\tstart_date\tend_date\tstart_age_days\tend_age_days\tsurvival_days\tstatus_start_raw\tstatus_end_raw\tdays_since_enroll_date\tfirst_event_date\tfirst_event_age_days\tdays_since_first_event_date\tstart_days_since_last_event_date\tcomputed_date\tmissing_fields\n")
	for {
		var r Result
		err := itr.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return pfx.Err(err)
		}

		// Special case: we are not doing any time-varying
		if timeVaryingDays <= 0 {
			printRow(r, diseaseName, todayDate, missing)
			continue
		}

		// Copy the result; we may mutate this copy
		mutableR := r

		// For most states, we will increment by `timeVaryingDays`.
		timeInterval := timeVaryingDays

		// If the last event was a disease event, then, if set, we will advance
		// the clock by a different, shorter amount of time. This will decay by
		// 2x every iteration over the loop (e.g., 30d, 60d, 120d) and is also
		// capped such that it will never be greater than `timeVaryingDays`.
		if mutableR.StatusStart == Disease {
			timeInterval = timeVaryingDaysAfterEvent
		}

		for {

			// If the survival time represented by this row is less than or
			// equal to the time-varying duration that we are trying to achieve,
			// we are done.
			if int(mutableR.SurvivalDays.Int64) <= timeInterval {
				printRow(mutableR, diseaseName, todayDate, missing)
				break
			}

			// If the row represent a longer survival time than we want, split
			// up the period and create new rows.
			timeVaryingEntry := mutableR

			// Most fields are simply copied forward, explicitly including
			// start_date, incident_number, and status_start.

			// The end status becomes censored
			timeVaryingEntry.StatusEnd = NoDisease
			timeVaryingEntry.IsFinalRecord.Bool = false

			// The end status becomes set to timeInterval after the start
			// date, and all end-date-related fields need to be updated to match
			// that.
			timeVaryingEntry.EndDate.Date = timeVaryingEntry.StartDate.Date.AddDays(timeInterval)
			timeVaryingEntry.SurvivalDays.Int64 = int64(timeVaryingEntry.EndDate.Date.DaysSince(timeVaryingEntry.StartDate.Date))
			timeVaryingEntry.DaysSinceFirstEventDate.Int64 = int64(timeVaryingEntry.EndDate.Date.DaysSince(timeVaryingEntry.FirstEventDate.Date))
			timeVaryingEntry.DaysSinceEnrollDate.Int64 = int64(timeVaryingEntry.EndDate.Date.DaysSince(timeVaryingEntry.EnrollDate.Date))
			timeVaryingEntry.EndAgeDays.Int64 = int64(timeVaryingEntry.EndDate.Date.DaysSince(timeVaryingEntry.BirthDate.Date))

			// Emit the updated row
			printRow(timeVaryingEntry, diseaseName, todayDate, missing)

			// Finally, modify start values in the parent row (moving them
			// forward beyond the end values of the child row we just created)
			// and then run the loop again until we have created enough time
			// chunks to complete the period of time.
			mutableR.StartDate.Date = mutableR.StartDate.Date.AddDays(timeInterval)
			mutableR.StartAgeDays.Int64 = int64(mutableR.StartDate.Date.DaysSince(mutableR.BirthDate.Date))
			mutableR.SurvivalDays.Int64 = int64(mutableR.EndDate.Date.DaysSince(mutableR.StartDate.Date))

			if mutableR.StatusStart == Disease {
				// Reach back into the unmodified row, since the unmodified
				// start date is the same as the event date of the last event.
				mutableR.StartDaysSinceLastEventDate = mutableR.StartDate.Date.DaysSince(r.StartDate.Date)

				// Decay the interval over time so we don't constantly deal with
				// 30-day chunks forever after the first disease instance, but
				// never go longer than the user-set timeVaryingDays interval
				timeInterval *= 2
				if timeInterval > timeVaryingDays {
					timeInterval = timeVaryingDays
				}
			}
		}
	}

	return nil
}

func printRow(r Result, diseaseName, todayDate, missing string) (int, error) {

	return fmt.Fprintf(STDOUT,
		"%s\t%d\t%s\t%s\t%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
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
		NA(r.FirstEventDate),
		NA(r.FirstEventAgeDays),
		NA(r.DaysSinceFirstEventDate),
		r.StartDaysSinceLastEventDate,
		todayDate,
		missing,
	)
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
