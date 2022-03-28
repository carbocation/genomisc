package main

import (
	"fmt"
	"strings"
)

type counters struct {
	N              int
	NValid         int
	Sex            int
	Ethnicity      int
	BirthMonth     int
	BirthYear      int
	LossToFollowUp int
	HasGPData      int
	HasDied        int
}

func (c *counters) CountValidFields(s CensorResult) {
	c.N++

	if s.IsValid() {
		c.NValid++
	} else {
		// Stop counting the rest if the sample isn't valid.
		return
	}

	if s.Sex.Valid {
		c.Sex++
	}
	if s.Ethnicity.Valid {
		c.Ethnicity++
	}
	if s.BornMonth.Valid {
		c.BirthMonth++
	}
	if s.BornYear.Valid {
		c.BirthYear++
	}
	if s.Lost.Valid {
		c.LossToFollowUp++
	}
	if s.HasPrimaryCareData {
		c.HasGPData++
	}
	if s.Died.Valid {
		c.HasDied++
	}
}

func (c counters) String() string {
	out := strings.Builder{}

	out.WriteString(fmt.Sprintf("%d participants identified.\n", c.N))
	out.WriteString(fmt.Sprintf("%d participants with valid data.\n", c.NValid))
	out.WriteString(fmt.Sprintf("%d valid participants with sex.\n", c.Sex))
	out.WriteString(fmt.Sprintf("%d valid participants with ethnicity.\n", c.Ethnicity))
	out.WriteString(fmt.Sprintf("%d valid participants with birthmonth.\n", c.BirthMonth))
	out.WriteString(fmt.Sprintf("%d valid participants with birthyear.\n", c.BirthYear))
	out.WriteString(fmt.Sprintf("%d valid participants were lost to follow-up.\n", c.LossToFollowUp))
	out.WriteString(fmt.Sprintf("%d valid participants have GP data.\n", c.HasGPData))
	out.WriteString(fmt.Sprintf("%d valid participants have died.\n", c.HasDied))

	return out.String()
}
