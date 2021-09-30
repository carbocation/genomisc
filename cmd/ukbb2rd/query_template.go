package main

import "text/template"

// mkMap allows you to create a map within a template, so that you can pass more
// than one parameter to a template block. Inspired by
// https://stackoverflow.com/a/25013152/199475 .
func mkMap(args ...interface{}) map[interface{}]interface{} {
	out := make(map[interface{}]interface{})
	for k, v := range args {
		if k%2 == 0 {
			continue
		}
		out[args[k-1]] = v
	}
	return out
}

// TODO: create a `has_died` field and apply the censor table's
// death_censor_date properly. Rename death_date to death_censor_date. Rename
// death_age to death_censor_age.
//
// TODO: Resolve age_censor vs enroll_age. Choose one or the other (likely the
// latter, so you end up with enroll_age, censor_age, death_censor_age).
var queryTemplate = template.Must(template.New("").Funcs(template.FuncMap(map[string]interface{}{"mkMap": mkMap})).Parse(`
WITH censor_table AS (
	SELECT * FROM ` + "`{{.database}}.censor`" + `
  ),
  undated_fields AS (
	SELECT 
	  p.sample_id, 
	  p.FieldID, 
	  p.value, 
	  MIN(SAFE.PARSE_DATE("%E4Y-%m-%d", denroll.value)) first_date
	FROM ` + "`{{.database}}.phenotype`" + ` p
	JOIN ` + "`{{.database}}.phenotype`" + ` denroll ON denroll.FieldID=53 AND denroll.sample_id=p.sample_id AND denroll.instance = 0 AND denroll.array_idx = 0
	WHERE TRUE
	  AND FALSE
	GROUP BY 
	  p.FieldID, 
	  p.sample_id, 
	  p.value
  ), 
  all_dated_fields AS (
	SELECT * EXCEPT(source) FROM ` + "`{{.materializedDatabase}}.materialized_hesin_dates_all`" + `
	UNION DISTINCT
	SELECT * FROM ` + "`{{.materializedDatabase}}.materialized_special_dates`" + `
	UNION DISTINCT
	SELECT * FROM undated_fields
  ),
  grouped_dated_fields_included_only AS (
	-- From here out, we don't have to care about the FieldID or the value
	SELECT 
	  sample_id, 
	  first_date event_date,
	  1 status_end,
	FROM all_dated_fields hd
	WHERE FALSE
	{{.includePart}}
	GROUP BY sample_id, first_date 
  ),
  grouped_dated_fields_excluded_only AS (
	-- From here out, we don't have to care about the FieldID or the value
	SELECT 
	  sample_id, 
	  MIN(first_date) event_date,
	  -1 status_end,
	FROM all_dated_fields hd
	WHERE FALSE
	{{.excludePart}}
	GROUP BY sample_id
  ),
  censoring_terminations AS (
	SELECT 
	  sample_id, 
	  CASE
		WHEN c.death_date IS NOT NULL THEN c.death_date
		WHEN c.lost_to_followup_date IS NOT NULL THEN c.lost_to_followup_date 
		ELSE c.phenotype_censor_date 
	  END event_date,
	  CASE
		WHEN c.death_date IS NOT NULL THEN 2
		WHEN c.lost_to_followup_date IS NOT NULL THEN 3
		ELSE 0
	  END status_end,
	FROM censor_table c
  ),
  censoring_excluding_terminations AS (
	-- When the person hit an exclusion criterion **and** that exclusion criterion occurred before the
	-- censoring criteria, apply the exclusion criterion
	SELECT 
	  sample_id,
	  CASE
		WHEN excl.event_date IS NOT NULL AND excl.event_date < ct.event_date THEN excl.event_date 
		ELSE ct.event_date
	  END event_date,
	  CASE
		WHEN excl.event_date IS NOT NULL AND excl.event_date < ct.event_date THEN excl.status_end
		ELSE ct.status_end
	  END status_end,
	FROM censoring_terminations ct
	LEFT JOIN grouped_dated_fields_excluded_only excl USING(sample_id)
  ),
  included_prior_to_censoring_or_exclusions AS (
	-- Event dates must not come after either censoring or exclusion cutoff dates
	SELECT include.* 
	FROM grouped_dated_fields_included_only include
	JOIN censoring_excluding_terminations censoring USING(sample_id)
	WHERE TRUE
	  AND include.event_date < censoring.event_date
  ),
  grouped_dated_fields_with_censoring AS (
	SELECT * FROM included_prior_to_censoring_or_exclusions
	UNION ALL
	SELECT * FROM censoring_excluding_terminations
  ), 
  full_res AS (
	SELECT 
	  grouped_dated_fields_with_censoring.*,
	  ROW_NUMBER() OVER(PARTITION BY sample_id ORDER BY event_date ASC, status_end ASC) - 1 incident_number,
	  enroll_date,
	FROM grouped_dated_fields_with_censoring
	JOIN censor_table c USING(sample_id)
  ),
  diffed AS (
	SELECT 
	  query.* EXCEPT(event_date),
	  COALESCE(comparator.status_end, 0) status_start,
	  comparator.event_date start_date,
	  query.event_date end_date,
	  SAFE.DATE_DIFF(query.event_date, comparator.event_date, DAY) days_since_start_date,
	  SAFE.DATE_DIFF(query.event_date, query.enroll_date, DAY) days_since_enroll_date,
	FROM full_res query
	LEFT JOIN full_res comparator ON 
	  query.sample_id = comparator.sample_id 
	  AND (
		FALSE
		OR query.incident_number = (comparator.incident_number + 1)
	  )
	WHERE TRUE
	  -- Only keep dates that are incident from the perspective of enrollment
		AND query.event_date > query.enroll_date
  )
  
  SELECT 
	sample_id,
	incident_number,
	status_start,
	status_end,
	-- c.enroll_date,
	-- When (i) a prevalent event occurs, then (ii) the person enrolls in UKBB, then (iii) an incident
	-- event occurs, for the incident event we will not count the time from the prevalent 
	-- event, but instead will count the time from enrollment.
	CASE 
	  WHEN start_date IS NULL THEN diffed.enroll_date 
	  WHEN start_date < c.enroll_date AND end_date > c.enroll_date THEN c.enroll_date
	  ELSE start_date 
	END start_date,
	end_date,
	CASE 
	  WHEN start_date IS NULL THEN days_since_enroll_date 
	  WHEN start_date < c.enroll_date AND end_date > c.enroll_date THEN SAFE.DATE_DIFF(c.enroll_date, start_date, DAY)
	  ELSE days_since_start_date
	END days_since_start_date,
	start_date start_date_unedited,
	days_since_start_date days_since_start_date_unedited,
	days_since_enroll_date,
  FROM diffed
  JOIN censor_table c USING(sample_id)
  WHERE TRUE
  ORDER BY sample_id, incident_number
`))
