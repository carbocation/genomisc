WITH death_dates AS (
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
FROM `{{.database}}.person` p
LEFT JOIN `{{.database}}.cb_search_person` cbsp USING(person_id)
LEFT JOIN death_dates d USING(person_id)
LEFT JOIN enrollment e USING(person_id)
