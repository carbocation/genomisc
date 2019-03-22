package main

import (
	"fmt"
	"time"

	"gopkg.in/guregu/null.v3"
)

type CensorResult struct {
	SampleID int64

	// Guaranteed
	enrolled time.Time
	computed time.Time // Date this was computed

	// May be null appropriately
	died null.Time
	lost null.Time

	// Unsure
	phenoCensored time.Time
	deathCensored time.Time

	bornYear  string
	bornMonth string
}

func (s CensorResult) Born() time.Time {
	month := s.bornMonth
	if month == "" {
		month = "7"
	}

	dt, err := time.Parse("2006-01-02", fmt.Sprintf("%04s-%02s-01", s.bornYear, month))
	if err != nil {
		return time.Time{}
	}

	return dt
}

func (s CensorResult) DiedString() string {
	if !s.died.Valid {
		return NullMarker
	}

	return TimeToUKBDate(s.died.Time)
}

func (s CensorResult) DeathCensored() time.Time {
	if s.died.Valid {
		return s.died.Time
	}

	if s.lost.Valid {
		return s.lost.Time
	}

	return s.deathCensored
}

func (s CensorResult) PhenoCensored() time.Time {
	if s.died.Valid {
		return s.died.Time
	}

	if s.lost.Valid {
		return s.lost.Time
	}

	return s.phenoCensored
}
