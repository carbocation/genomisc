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

{{/* The pattern for identifying the earliest date of an 
	exclusion criterion and the earliest date of an inclusion criterion 
	is identical, so this is templated here. */}}

{{define "include"}}
	-- From here out, we don't have to care about the FieldID or the value
	SELECT 
		sample_id, 
		first_date,
	FROM all_dated_fields hd
	WHERE FALSE
		{{.whichPart}}
	GROUP BY sample_id, first_date, value
{{end}}

{{define "exclude"}}
SELECT 
	sample_id, 
	has_disease, 
	incident_disease, 
	prevalent_disease, 
	date_censor, 
	DATE_DIFF(date_censor,birthdate, DAY)/365.25 age_censor, 
	DATE_DIFF(date_censor,birthdate, DAY) age_censor_days,
	birthdate, 
	enroll_date, 
	enroll_age, 
	enroll_age_days,
	death_date, 
	death_age, 
	death_age_days,
	computed_date, 
	missing_fields
FROM (
	SELECT 
		c.sample_id, 
		CASE 
			WHEN MIN(hd.first_date) IS NOT NULL THEN 1
			ELSE 0
		END has_disease,
		CASE 
			WHEN MIN(hd.first_date) > MIN(c.enroll_date) THEN 1
			WHEN MIN(hd.first_date) IS NOT NULL THEN NULL
			ELSE 0
		END incident_disease,
		CASE 
			WHEN MIN(hd.first_date) > MIN(c.enroll_date) THEN 0
			WHEN MIN(hd.first_date) IS NOT NULL THEN 1
			ELSE 0
		END prevalent_disease,
		CASE 
			WHEN MIN(hd.first_date) IS NOT NULL THEN MIN(hd.first_date)
			ELSE MIN(c.phenotype_censor_date)
		END date_censor,
		MIN(c.birthdate) birthdate,
		MIN(c.enroll_date) enroll_date,
		MIN(c.enroll_age) enroll_age,
		MIN(c.enroll_age_days) enroll_age_days,
		MIN(c.death_date) death_date,
		MIN(c.death_age) death_age,
		MIN(c.death_age_days) death_age_days,
		MIN(c.computed_date) computed_date,
		MIN(c.missing_fields) missing_fields
	FROM ` + "`{{.g.database}}.censor`" + ` c
	LEFT OUTER JOIN (
		{{if .g.use_gp}}
		SELECT * FROM ` + "`{{.g.materializedDatabase}}.materialized_gp_dates`" + `
		UNION DISTINCT
		{{end}}
		SELECT * FROM ` + "`{{.g.materializedDatabase}}.materialized_hesin_dates`" + `
		UNION DISTINCT
		SELECT * FROM ` + "`{{.g.materializedDatabase}}.materialized_special_dates`" + `
		UNION DISTINCT
		SELECT * FROM undated_fields
	) hd ON c.sample_id=hd.sample_id
	AND (
		FALSE
		{{/* The .includePart or .excludePart is passed here */}}
		{{.whichPart}}
	)
	GROUP BY 
		sample_id
)
{{end}}

WITH undated_fields AS (
	SELECT 
		p.sample_id, 
		p.FieldID, 
		p.value, 
		MIN(SAFE.PARSE_DATE("%E4Y-%m-%d", denroll.value)) first_date
	FROM ` + "`{{.database}}.phenotype`" + ` p
	JOIN ` + "`{{.database}}.phenotype`" + ` denroll ON denroll.FieldID=53 AND denroll.sample_id=p.sample_id AND denroll.instance = 0 AND denroll.array_idx = 0
	WHERE TRUE
		{{.standardPart}}
	GROUP BY 
		p.FieldID, 
		p.sample_id, 
		p.value
), 
all_dated_fields AS (
	SELECT * EXCEPT(source) FROM ` + "{{.materializedDatabase}}.materialized_hesin_dates_all" + `
	UNION DISTINCT
	SELECT * FROM ` + "{{.materializedDatabase}}.materialized_special_dates" + `
	UNION DISTINCT
	SELECT * FROM undated_fields
),
grouped_dated_fields_included_only AS (
	{{template "include" (mkMap "g" . "whichPart" .includePart)}}
),
included_only AS (
	SELECT 
		sample_id, 
		incident_number,
		has_disease, 
		incident_disease, 
		prevalent_disease, 
		date_censor, 
		DATE_DIFF(date_censor,birthdate, DAY)/365.25 age_censor, 
		DATE_DIFF(date_censor,birthdate, DAY) age_censor_days,
		birthdate, 
		enroll_date, 
		enroll_age, 
		enroll_age_days,
		death_date, 
		death_age, 
		death_age_days,
		computed_date, 
		missing_fields
	FROM (
        SELECT 
			c.sample_id, 
			ROW_NUMBER() OVER(PARTITION BY c.sample_id ORDER BY hd.first_date ASC) - 1 incident_number,
			CASE 
					WHEN hd.first_date IS NOT NULL THEN 1
					ELSE 0
			END has_disease,
			CASE 
					WHEN hd.first_date > c.enroll_date THEN 1
					WHEN hd.first_date IS NOT NULL THEN NULL
					ELSE 0
			END incident_disease,
			CASE 
					WHEN hd.first_date > c.enroll_date THEN 0
					WHEN hd.first_date IS NOT NULL THEN 1
					ELSE 0
			END prevalent_disease,
			CASE 
					WHEN hd.first_date IS NOT NULL THEN hd.first_date
					ELSE c.phenotype_censor_date
			END date_censor,
			c.birthdate birthdate,
			c.enroll_date enroll_date,
			c.enroll_age enroll_age,
			c.enroll_age_days enroll_age_days,
			c.death_date death_date,
			c.death_age death_age,
			c.death_age_days death_age_days,
			c.computed_date computed_date,
			c.missing_fields missing_fields
		FROM ` + "{{.database}}.censor" + ` c
		LEFT OUTER JOIN grouped_dated_fields_included_only hd ON c.sample_id=hd.sample_id
	)
), 
excluded_only AS (
	{{template "exclude" (mkMap "g" . "whichPart" .excludePart)}}
),
full_res AS (
	SELECT 
		c.sample_id, 
		COALESCE(incident_number, 0) incident_number,
		CASE 
			-- Enrollment occurred after exclusion:
			WHEN eo.has_disease = 1 AND SAFE.DATE_DIFF(c.enroll_date, eo.date_censor, DAY) > 0 THEN NULL
			-- Exclusion occurred after enrollment and prior to disease onset; we will censor:
			WHEN eo.has_disease = 1 AND io.has_disease = 1 AND SAFE.DATE_DIFF(io.date_censor,eo.date_censor, DAY) > 0 THEN 0
			-- Exclusion occurred after disease onset; we'll allow it:
			WHEN eo.has_disease = 1 AND io.has_disease = 1 AND SAFE.DATE_DIFF(eo.date_censor,io.date_censor, DAY) > 0 THEN io.has_disease 
			-- Met exclusion but no inclusion; occurred after enrollment (due to above rule); censor:
			WHEN eo.has_disease = 1 AND (io.has_disease = 0 OR io.has_disease IS NULL) THEN 0
			-- Didn't meet exclusion or inclusion means we censor at the date given by UKBB:
			WHEN io.has_disease IS NULL THEN 0
			ELSE io.has_disease
		END has_disease, 

		CASE 
			-- Enrollment occurred after exclusion:
			WHEN eo.has_disease = 1 AND SAFE.DATE_DIFF(c.enroll_date, eo.date_censor, DAY) > 0 THEN NULL
			-- Exclusion occurred after enrollment and prior to disease onset; we will censor:
			WHEN eo.has_disease = 1 AND io.has_disease = 1 AND SAFE.DATE_DIFF(io.date_censor,eo.date_censor, DAY) > 0 THEN 0
			-- Exclusion occurred after disease onset; we'll allow it:
			WHEN eo.has_disease = 1 AND io.has_disease = 1 AND SAFE.DATE_DIFF(eo.date_censor,io.date_censor, DAY) > 0 THEN io.incident_disease
			-- Met exclusion but no inclusion; occurred after enrollment (due to above rule); censor:
			WHEN eo.has_disease = 1 AND (io.has_disease = 0 OR io.has_disease IS NULL) THEN 0
			-- Didn't meet exclusion or inclusion means we censor at the date given by UKBB:
			WHEN io.has_disease IS NULL THEN 0
			ELSE io.incident_disease
		END incident_disease, 

		CASE 
			-- Enrollment occurred after exclusion:
			WHEN eo.has_disease = 1 AND SAFE.DATE_DIFF(c.enroll_date, eo.date_censor, DAY) > 0 THEN NULL
			-- Exclusion occurred after enrollment and prior to disease onset; we will exclude:
			WHEN eo.has_disease = 1 AND io.has_disease = 1 AND SAFE.DATE_DIFF(io.date_censor,eo.date_censor, DAY) > 0 THEN 0
			-- Exclusion occurred after disease onset; we'll allow it:
			WHEN eo.has_disease = 1 AND io.has_disease = 1 AND SAFE.DATE_DIFF(eo.date_censor,io.date_censor, DAY) > 0 THEN io.prevalent_disease
			-- Met exclusion but no inclusion; occurred after enrollment (due to above rule); censor:
			WHEN eo.has_disease = 1 AND (io.has_disease = 0 OR io.has_disease IS NULL) THEN 0
			-- Didn't meet exclusion or inclusion means we censor at the date given by UKBB:
			WHEN io.has_disease IS NULL THEN 0
			ELSE io.prevalent_disease
		END prevalent_disease, 

		CASE 
			-- Enrollment occurred after exclusion:
			WHEN eo.has_disease = 1 AND SAFE.DATE_DIFF(c.enroll_date, eo.date_censor, DAY) > 0 THEN 1
			-- Exclusion occurred after enrollment and prior to disease onset; we will exclude:
			WHEN eo.has_disease = 1 AND io.has_disease = 1 AND SAFE.DATE_DIFF(io.date_censor,eo.date_censor, DAY) > 0 THEN 1
			-- Exclusion occurred after disease onset; we'll allow it:
			WHEN eo.has_disease = 1 AND io.has_disease = 1 AND SAFE.DATE_DIFF(eo.date_censor,io.date_censor, DAY) > 0 THEN 0
			-- Met exclusion but no inclusion
			WHEN eo.has_disease = 1 AND (io.has_disease = 0 OR io.has_disease IS NULL) THEN 1
			-- Didn't get excluded:
			ELSE 0
		END met_exclusion, 

		CASE 
			-- Enrollment occurred after exclusion:
			WHEN eo.has_disease = 1 AND SAFE.DATE_DIFF(c.enroll_date, eo.date_censor, DAY) > 0 THEN eo.date_censor
			-- Exclusion occurred after enrollment and prior to disease onset; we will exclude:
			WHEN eo.has_disease = 1 AND io.has_disease = 1 AND SAFE.DATE_DIFF(io.date_censor,eo.date_censor, DAY) > 0 THEN eo.date_censor
			-- Exclusion occurred after disease onset; we'll allow it:
			WHEN eo.has_disease = 1 AND io.has_disease = 1 AND SAFE.DATE_DIFF(eo.date_censor,io.date_censor, DAY) > 0 THEN io.date_censor
			-- Met exclusion but no inclusion
			WHEN eo.has_disease = 1 AND (io.has_disease = 0 OR io.has_disease IS NULL) THEN eo.date_censor
			-- Didn't meet exclusion or inclusion means we censor at the date given by UKBB:
			WHEN io.has_disease IS NULL THEN c.phenotype_censor_date
			ELSE io.date_censor
		END date_censor, 

		-- If you modify age_censor, don't forget to modify age_censor_days equivalently
		CASE
			-- Enrollment occurred after exclusion:
			WHEN eo.has_disease = 1 AND SAFE.DATE_DIFF(c.enroll_date, eo.date_censor, DAY) > 0 THEN DATE_DIFF(eo.date_censor,c.birthdate, DAY)/365.25
			-- Exclusion occurred after enrollment and prior to disease onset; we will exclude:
			WHEN eo.has_disease = 1 AND io.has_disease = 1 AND SAFE.DATE_DIFF(io.date_censor,eo.date_censor, DAY) > 0 THEN DATE_DIFF(eo.date_censor,c.birthdate, DAY)/365.25
			-- Exclusion occurred after disease onset; we'll allow it:
			WHEN eo.has_disease = 1 AND io.has_disease = 1 AND SAFE.DATE_DIFF(eo.date_censor,io.date_censor, DAY) > 0 THEN DATE_DIFF(io.date_censor,c.birthdate, DAY)/365.25 
			-- Met exclusion but no inclusion
			WHEN eo.has_disease = 1 AND (io.has_disease = 0 OR io.has_disease IS NULL) THEN DATE_DIFF(eo.date_censor,c.birthdate, DAY)/365.25
			-- Didn't meet exclusion or inclusion means we censor at the date given by UKBB:
			WHEN io.has_disease IS NULL THEN c.phenotype_censor_age
			ELSE DATE_DIFF(io.date_censor,c.birthdate, DAY)/365.25 
		END age_censor, 

		-- Designed to be a duplicate of age_censor but with days instead of years
		CASE
			WHEN eo.has_disease = 1 AND SAFE.DATE_DIFF(c.enroll_date, eo.date_censor, DAY) > 0 THEN DATE_DIFF(eo.date_censor,c.birthdate, DAY)
			WHEN eo.has_disease = 1 AND io.has_disease = 1 AND SAFE.DATE_DIFF(io.date_censor,eo.date_censor, DAY) > 0 THEN DATE_DIFF(eo.date_censor,c.birthdate, DAY)
			WHEN eo.has_disease = 1 AND io.has_disease = 1 AND SAFE.DATE_DIFF(eo.date_censor,io.date_censor, DAY) > 0 THEN DATE_DIFF(io.date_censor,c.birthdate, DAY)
			WHEN eo.has_disease = 1 AND (io.has_disease = 0 OR io.has_disease IS NULL) THEN DATE_DIFF(eo.date_censor,c.birthdate, DAY)
			WHEN io.has_disease IS NULL THEN c.phenotype_censor_age_days
			ELSE DATE_DIFF(io.date_censor,c.birthdate, DAY)
		END age_censor_days, 

		c.birthdate, 
		c.enroll_date, 
		c.enroll_age, 
		c.enroll_age_days,
		CASE 
			WHEN c.death_date IS NULL THEN 0 
			ELSE 1 
		END has_died,
		CASE 
			WHEN c.death_date IS NULL THEN c.death_censor_date 
			ELSE c.death_date 
		END death_date, 
		CASE WHEN c.death_date IS NULL THEN c.death_censor_age 
			ELSE c.death_age 
		END death_age, 
		CASE WHEN c.death_date IS NULL THEN c.death_censor_age_days 
			ELSE c.death_age_days 
		END death_age_days, 
		c.computed_date, 
		c.missing_fields
	FROM ` + "`{{.database}}.censor`" + ` c
	LEFT JOIN included_only io ON io.sample_id=c.sample_id
	LEFT JOIN excluded_only eo ON eo.sample_id=c.sample_id
	ORDER BY 
		has_disease DESC, 
		incident_disease DESC, 
		age_censor ASC
),
diffed AS (
	SELECT 
		query.sample_id, 
		query.incident_number, 
		CASE 
			WHEN query.incident_number = 0 THEN query.enroll_date
			ELSE comparator.date_censor
		END date_start,
		CASE 
			WHEN query.incident_number = 0 THEN SAFE.DATE_DIFF(query.date_censor, query.enroll_date, DAY)
			ELSE SAFE.DATE_DIFF(query.date_censor, comparator.date_censor, DAY)
		END days_elapsed,
		query.* EXCEPT(sample_id, incident_number)
	FROM full_res query
	LEFT JOIN full_res comparator ON 
		query.sample_id = comparator.sample_id 
		AND (
			FALSE
			OR query.incident_number = (comparator.incident_number + 1)
		)
)

SELECT *
FROM diffed
WHERE TRUE

`))
