package main

import (
	"fmt"

	"cloud.google.com/go/bigquery"
)

func BoolToInt(input bool) int {
	if input {
		return 1
	}

	return 0
}

func TimesToFractionalYears(earlier, later bigquery.NullDate) string {
	if !earlier.Valid || !later.Valid {
		return "NA"
	}

	return fmt.Sprintf("%.6f", float64(later.Date.DaysSince(earlier.Date))/365.25)
}

func TimesToDays(earlier, later bigquery.NullDate) string {
	if !earlier.Valid || !later.Valid {
		return "NA"
	}

	return fmt.Sprintf("%d", later.Date.DaysSince(earlier.Date))
}
