WITH all_samples AS (
	SELECT 
		sample_id
	FROM `{{.database}}.phenotype`
	GROUP BY sample_id
)
, enrollment AS (
  SELECT sample_id, value
  FROM `{{.database}}.phenotype`
  WHERE TRUE
    AND FieldID = 53
    AND instance = 0
    AND array_idx = 0
)
, birthyear AS (
  SELECT sample_id, value
  FROM `{{.database}}.phenotype`
  WHERE TRUE
    AND FieldID = 34
    AND instance = 0
    AND array_idx = 0
)
, birthmonth AS (
  SELECT sample_id, value
  FROM `{{.database}}.phenotype`
  WHERE TRUE
    AND FieldID = 52
    AND instance = 0
    AND array_idx = 0
)
, lostdate AS (
  SELECT sample_id, value
  FROM `{{.database}}.phenotype`
  WHERE TRUE
    AND FieldID = 191
    AND instance = 0
    AND array_idx = 0
)
, primarycareflag AS (
  SELECT sample_id, value
  FROM `{{.database}}.phenotype`
  WHERE TRUE
    AND FieldID = 42038
    AND instance = 0
    AND array_idx = 0
)
, sex AS (
	SELECT 
		sample_id, 
		c.meaning value
	FROM `{{.database}}.phenotype` p
	JOIN `{{.database}}.dictionary` d USING(FieldID)
	JOIN `{{.database}}.coding` c ON c.coding_file_id = d.coding_file_id AND c.coding = p.value
	WHERE TRUE
		-- Sex is not instanced
		AND d.FieldID = 31
)
, ethnicity AS (
	SELECT 
		sample_id, 
		SAFE.REPLACE(c.meaning, " ", "_") value
	FROM `{{.database}}.phenotype` p
	JOIN `{{.database}}.dictionary` d USING(FieldID)
	JOIN `{{.database}}.coding` c ON c.coding_file_id = d.coding_file_id AND c.coding = p.value
	WHERE TRUE
		AND d.FieldID = 21000
		AND p.instance = 0
)
, death_pheno AS (
  SELECT sample_id, 
    MAX(CAST(value AS DATE)) value
  FROM `{{.database}}.phenotype`
  WHERE TRUE
    AND FieldID = 40000
  GROUP BY sample_id
)
, death_hesin AS (
	SELECT 
	eid sample_id,
	MAX(date_of_death) value
	FROM `{{.database}}.death`
	GROUP BY sample_id
)
, death_all_duplicate AS (
  SELECT * FROM death_pheno
  UNION ALL
  SELECT * FROM death_hesin
)
, death_all AS (
  SELECT sample_id, MAX(value) value
  FROM death_all_duplicate 
  GROUP BY sample_id
)


SELECT 
  all_samples.sample_id,
  PARSE_DATE('%E4Y-%m-%d', enrollment.value) enrolled,
  birthyear.value born_year,
  birthmonth.value born_month,
  PARSE_DATE('%E4Y-%m-%d', lostdate.value) lost,
  COALESCE(CAST(primarycareflag.value AS INT), 0) > 0 has_primary_care_data,
  sex.value sex,
  ethnicity.value ethnicity,
  death_all.value died,
FROM all_samples
LEFT JOIN enrollment USING(sample_id)
LEFT JOIN birthyear USING(sample_id)
LEFT JOIN birthmonth USING(sample_id)
LEFT JOIN lostdate USING(sample_id)
LEFT JOIN primarycareflag USING(sample_id)
LEFT JOIN sex USING(sample_id)
LEFT JOIN ethnicity USING(sample_id)
LEFT JOIN death_all USING(sample_id)
ORDER BY sample_id ASC
