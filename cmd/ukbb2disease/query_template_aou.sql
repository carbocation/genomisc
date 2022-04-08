{{define "include_exclude"}}
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
	FROM censor c
	LEFT OUTER JOIN (
		SELECT * FROM ukbb_first_occurrence
		-- TODO: Add other data beyond ICD. E.g.:
		UNION DISTINCT
		SELECT * FROM ukbb_first_procedure_occurrence
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

WITH 
--------------------------------------------------------------------------------
-- CTEs to reconstruct pieces of the censor table
--------------------------------------------------------------------------------
death_dates AS (
  -- People can have multiple entries in the death table
	SELECT
		d.person_id,
		MAX(d.death_date) death_date,
	FROM `{{.database}}.death` d
	GROUP BY person_id
)
, enrollment AS (
  -- We're given age at enrollment, but here we compute an estimated *date* at
  -- enrollment
  SELECT
	cbsp.person_id,
	DATE_ADD(cbsp.dob, INTERVAL CAST(ROUND(cbsp.age_at_cdr*365.25, 0) AS INT64) DAY) enrolled, 
  FROM `{{.database}}.cb_search_person` cbsp
)
, censor_query AS (
	SELECT 
		p.person_id sample_id,
		e.enrolled,
		CAST(p.year_of_birth AS STRING) born_year,
		CAST(p.month_of_birth AS STRING) born_month,
		CAST(NULL AS DATE) lost,
		CASE WHEN cbsp.has_ehr_data IS NULL THEN FALSE ELSE cbsp.has_ehr_data >= 1 END has_primary_care_data,
		cbsp.sex_at_birth sex,
		CONCAT(cbsp.race, '|', cbsp.ethnicity) ethnicity,
		d.death_date died,
		-- Extension
		CAST(p.birth_datetime AS DATE) birthdate,
	FROM `{{.database}}.person` p
	LEFT JOIN `{{.database}}.cb_search_person` cbsp USING(person_id)
	LEFT JOIN death_dates d USING(person_id)
	LEFT JOIN enrollment e USING(person_id)
)
-- Mimicking the precomputed censor table
, censor AS (
	SELECT 
		sample_id,
		CAST(NULL AS STRING) missing_fields,
		CURRENT_DATE() computed_date,
		birthdate,
		enrolled enroll_date,
		SAFE.DATE_DIFF(enrolled, birthdate, DAY)/1.0 enroll_age_days,
		SAFE.DATE_DIFF(enrolled, birthdate, DAY)/365.25 enroll_age,
		died death_date,
		SAFE.DATE_DIFF(died, birthdate, DAY)/1.0 death_age_days,
		SAFE.DATE_DIFF(died, birthdate, DAY)/365.25 death_age,
		-- TODO: this date needs to be a query parameter
		PARSE_DATE("%F", "2021-04-01") phenotype_censor_date,
		SAFE.DATE_DIFF(PARSE_DATE("%F", "2021-04-01"), birthdate, DAY)/1.0 phenotype_censor_age_days,
		SAFE.DATE_DIFF(PARSE_DATE("%F", "2021-04-01"), birthdate, DAY)/365.25 phenotype_censor_age,
		-- TODO: this date needs to be a query parameter
		PARSE_DATE("%F", "2021-04-01") death_censor_date,
		SAFE.DATE_DIFF(PARSE_DATE("%F", "2021-04-01"), birthdate, DAY)/1.0 death_censor_age_days,
		SAFE.DATE_DIFF(PARSE_DATE("%F", "2021-04-01"), birthdate, DAY)/365.25 death_censor_age,
	FROM censor_query
)
--------------------------------------------------------------------------------
-- Usual ukbb2disease CTE queries
--------------------------------------------------------------------------------
, first_occurrence AS (
	SELECT 
		co.person_id, 
		c.vocabulary_id,
		c.concept_code,
		MIN(co.condition_start_date) condition_start_date,
	FROM `{{.database}}.condition_occurrence` co
	JOIN `{{.database}}.concept` c ON (c.concept_id = co.condition_concept_id OR c.concept_id = co.condition_source_concept_id)
	WHERE TRUE
		AND c.vocabulary_id IN ('ICD9CM', 'ICD10CM', 'SNOMED')
	GROUP BY 
		person_id, 
		vocabulary_id,
		concept_code
)
, ukbb_first_occurrence AS (
	-- Here we simulate the structure that we use for UK Biobank to faciliate
	-- the use of tools that expect that format. We assign ICD9CM codes to the
	-- UK Biobank FieldID for ICD9, and the ICD10CM codes to the UK Biobank
	-- FieldID for ICD10. Even though, note, that the -CM codes can differ
	-- somewhat from the non-CM codes used in UK Biobank.
	SELECT
		person_id sample_id,
		-- Note that the field types are strings, not integers, so we refer to
		-- FieldName rather than FieldID.
		vocabulary_id FieldName, -- E.g., 'ICD9CM' or 'SNOMED'
		-- CASE 
		--     WHEN vocabulary_id = 'ICD9CM' THEN 41203
		--     WHEN vocabulary_id = 'ICD10CM' THEN 41202
		--     ELSE NULL
		-- END FieldID,
		-- Since the UK Biobank ICD9 and ICD10 fields strip out the ".", one
		-- approach would be to do the same for other datasets. But the approach
		-- I'm taking for now for non-UK Biobank datasets is to require that the
		-- tabfiles be properly formatted (i.e., include the "." where it is
		-- supposed to be in an ICD code).
		-- 
		-- REPLACE(concept_code, '.', '') value,
		concept_code value,
		condition_start_date first_date,
	FROM first_occurrence
)
, first_procedure_occurrence AS (
	SELECT 
		co.person_id, 
		c.vocabulary_id,
		c.concept_code,
		MIN(co.procedure_date) condition_start_date,
	FROM `{{.database}}.procedure_occurrence` co
	JOIN `{{.database}}.concept` c ON (c.concept_id = co.procedure_concept_id OR c.concept_id = co.procedure_source_concept_id)
	WHERE TRUE
		AND c.vocabulary_id IN ('ICD9Proc', 'ICD10PCS', 'SNOMED', 'CPT4')
	GROUP BY 
		person_id, 
		vocabulary_id,
		concept_code
)
, ukbb_first_procedure_occurrence AS (
	SELECT
		person_id sample_id,
		-- Note that the field types are strings, not integers, so we refer to
		-- FieldName rather than FieldID.
		vocabulary_id FieldName, -- E.g., 'ICD9CM' or 'SNOMED'
		concept_code value,
		condition_start_date first_date,
	FROM first_procedure_occurrence
)
, included_only AS (
	{{template "include_exclude" (mkMap "g" . "whichPart" .includePart)}}
)
, excluded_only AS (
	{{template "include_exclude" (mkMap "g" . "whichPart" .excludePart)}}
)

SELECT 
	c.sample_id, 
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
FROM censor c
LEFT JOIN included_only io ON io.sample_id=c.sample_id
LEFT JOIN excluded_only eo ON eo.sample_id=c.sample_id
ORDER BY 
	has_disease DESC, 
	incident_disease DESC, 
	age_censor ASC

