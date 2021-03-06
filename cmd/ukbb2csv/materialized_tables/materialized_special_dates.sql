-- Fields for which the date at which an event first happened was reported.
WITH dated_fields AS (
  SELECT p.FieldID, p.sample_id eid, p.value code, cod.meaning,
    CASE
      WHEN SAFE.PARSE_DATE("%E4Y-%m-%d", d.value) IS NULL THEN SAFE.PARSE_DATE("%E4Y-%m-%d", denroll.value)
      WHEN cod.meaning LIKE ('%unknown%') THEN SAFE.PARSE_DATE("%E4Y-%m-%d", denroll.value)
      ELSE SAFE.PARSE_DATE("%E4Y-%m-%d", d.value)
    END vdate
  FROM `ukbb-analyses.ukbb7089_201910.phenotype` p
  JOIN `ukbb-analyses.ukbb7089_201910.phenotype` denroll ON denroll.FieldID=53 AND denroll.sample_id=p.sample_id AND denroll.instance = 0 AND denroll.array_idx = 0
  JOIN `ukbb-analyses.ukbb7089_201910.phenotype` d ON d.sample_id=p.sample_id AND d.instance = p.instance AND d.array_idx = p.array_idx
    AND (
      FALSE
      -- p for phenotype field, d for date field
      OR (p.FieldID=42013 AND d.FieldID=42012) -- Subarachnoid hemorrhage
      OR (p.FieldID=42011 AND d.FieldID=42010) -- Intracerebral hemorrhage
      OR (p.FieldID=42009 AND d.FieldID=42008) -- Ischemic stroke
      OR (p.FieldID=42007 AND d.FieldID=42006) -- Stroke
      OR (p.FieldID=42001 AND d.FieldID=42000) -- MI
      OR (p.FieldID=42003 AND d.FieldID=42002) -- STEMI
      OR (p.FieldID=42005 AND d.FieldID=42004) -- NSTEMI
      OR (p.FieldID=40021 AND d.FieldID=40005) -- Cancer Registry record origin (easy way to check for existence of at least one cancer)
      OR (p.FieldID=40020 AND d.FieldID=40000) -- death
      OR (p.FieldID=40006 AND d.FieldID=40005) -- Cancer Registry ICD10
      OR (p.FieldID=40013 AND d.FieldID=40005) -- Cancer Registry ICD9
      OR (p.FieldID=42027 AND d.FieldID=42026) -- ESRD
      OR (p.FieldID=42019 AND d.FieldID=42018) -- All-cause dementia
      OR (p.FieldID=42021 AND d.FieldID=42020) -- Alzheimer dementia
      OR (p.FieldID=42023 AND d.FieldID=42022) -- Vascular dementia
      OR (p.FieldID=42025 AND d.FieldID=42024) -- Frontotemporal dementia
    )
  LEFT JOIN `ukbb-analyses.ukbb7089_201910.coding` cod ON cod.coding_file_id = d.coding_file_id AND cod.coding = d.value
), 
-- Fields for which the year at which an event first happened (rather than the
-- *date*) is reported.
dated_fields_fractional AS (
  SELECT p.FieldID, p.sample_id eid, p.value code, cod.meaning,
  CASE
      WHEN SAFE.PARSE_DATE("%Y", d.value) IS NULL THEN SAFE.PARSE_DATE("%E4Y-%m-%d", denroll.value)
      WHEN cod.meaning LIKE ('%unknown%') THEN SAFE.PARSE_DATE("%E4Y-%m-%d", denroll.value)
      ELSE SAFE.PARSE_DATE("%Y", d.value)
    END vdate
  FROM `ukbb-analyses.ukbb7089_201910.phenotype` p
  JOIN `ukbb-analyses.ukbb7089_201910.phenotype` denroll ON denroll.FieldID=53 AND denroll.sample_id=p.sample_id AND denroll.instance = 0 AND denroll.array_idx = 0
  JOIN `ukbb-analyses.ukbb7089_201910.phenotype` d ON d.sample_id=p.sample_id AND d.instance = p.instance AND d.array_idx = p.array_idx
    AND (
      FALSE
      OR (p.FieldID=20004 AND d.FieldID=20010)
      OR (p.FieldID=20002 AND d.FieldID=20008)
      OR (p.FieldID=20001 AND d.FieldID=20006)
    )
  LEFT JOIN `ukbb-analyses.ukbb7089_201910.coding` cod ON cod.coding_file_id = d.coding_file_id AND cod.coding = d.value
),
-- Fields for which the age (not date) of first event is explained in a field,
-- where the FieldID for the age lookup depends on the field value
self_reported_aged_subfields AS (
  SELECT p.FieldID, p.sample_id eid, p.value code, cod.meaning,
    CASE
      -- If the participant declined to state when the event occurred, assign it
      -- as of the date of enrollment
      WHEN p.FieldID=6150 AND d.value IN ('-1', '-3') THEN SAFE.PARSE_DATE("%E4Y-%m-%d", denroll.value)
      -- The decline values for FieldID 6152
      WHEN p.FieldID=6152 AND d.value IN ('-3', '-7') THEN SAFE.PARSE_DATE("%E4Y-%m-%d", denroll.value)
      -- The decline values for diabetes fields
      WHEN p.FieldID IN (2443) AND d.value IN ('-1', '-3') THEN SAFE.PARSE_DATE("%E4Y-%m-%d", denroll.value)
      WHEN p.FieldID IN (4041) AND d.value IN ('-1', '-2', '-3') THEN SAFE.PARSE_DATE("%E4Y-%m-%d", denroll.value)
      -- If the participant did state when the event occurred, figure out the
      -- date of that event based on their age and their birthdate
      WHEN SAFE_CAST(SAFE_CAST(d.value AS FLOAT64)*365 AS INT64) IS NOT NULL THEN SAFE.DATE_ADD(c.birthdate, INTERVAL SAFE_CAST(SAFE_CAST(d.value AS FLOAT64)*365 AS INT64) DAY)
      -- If there was a parsing issue, assign it as the date of enrollment
      ELSE SAFE.PARSE_DATE("%E4Y-%m-%d", denroll.value)
    END vdate
  FROM `ukbb-analyses.ukbb7089_201910.censor` c
  JOIN `ukbb-analyses.ukbb7089_201910.phenotype` p on p.sample_id = c.sample_id
  JOIN `ukbb-analyses.ukbb7089_201910.phenotype` denroll ON denroll.FieldID=53 AND denroll.sample_id=p.sample_id
  JOIN `ukbb-analyses.ukbb7089_201910.phenotype` d ON d.sample_id=p.sample_id AND d.instance = p.instance AND d.array_idx = p.array_idx
    AND (
      FALSE
      -- p for phenotype field, d for date field
      
      -- FieldID 6150
      OR (p.FieldID=6150 AND p.value='1' AND d.FieldID=3894) -- Self-reported MI
      OR (p.FieldID=6150 AND p.value='2' AND d.FieldID=3627) -- Self-reported angina
      OR (p.FieldID=6150 AND p.value='3' AND d.FieldID=4056) -- Self-reported stroke
      OR (p.FieldID=6150 AND p.value='4' AND d.FieldID=2966) -- Self-reported hypertension
      
      -- FieldID 6152
      OR (p.FieldID=6152 AND p.value='5' AND d.FieldID=4012) -- Self-reported DVT
      OR (p.FieldID=6152 AND p.value='6' AND d.FieldID=3992) -- Self-reported Emphysema
      OR (p.FieldID=6152 AND p.value='7' AND d.FieldID=4022) -- Self-reported PE
      OR (p.FieldID=6152 AND p.value='8' AND d.FieldID=3786) -- Self-reported Asthma
      OR (p.FieldID=6152 AND p.value='9' AND d.FieldID=3761) -- Self-reported Hayfever/rhinitis/eczema
      
      -- Diabetes FieldIDs
      OR (p.FieldID=2443 AND p.value='1' AND d.FieldID=2976) -- Self-reported DM
      OR (p.FieldID=4041 AND p.value='1' AND d.FieldID=2976) -- Self-reported gestational diabetes
    )
    
  LEFT JOIN `ukbb-analyses.ukbb7089_201910.dictionary` dict ON dict.FieldID = d.FieldID 
  LEFT JOIN `ukbb-analyses.ukbb7089_201910.coding` cod ON cod.coding_file_id = d.coding_file_id AND cod.coding = d.value AND cod.coding_file_id = dict.coding_file_id 
), 
-- Fields for which the age at the time of event is known not to be 
-- recorded, and so therefore the date is assigned as the date of enrollment
self_reported_undated_fields AS (
  SELECT p.FieldID, p.sample_id eid, p.value code, '' meaning,
      -- Assign it as of the date of enrollment
      c.enroll_date vdate
  FROM `ukbb-analyses.ukbb7089_201910.censor` c
  JOIN `ukbb-analyses.ukbb7089_201910.phenotype` p on p.sample_id = c.sample_id AND 
  (
    FALSE
    OR (p.FieldID=3079 AND p.value='1') -- Pacemaker
  )
),
-- The following are used to construct values for the "First dates" precomputed
-- fields provided by UK Biobank
first_dates AS (
  -- Date
  SELECT 
    p.sample_id, 
    p.FieldID, 
    -- Caution: brittle, expects disease code to always be in 0-based position 1
    SPLIT(Field, ' ')[OFFSET(1)] value,
    p.value first_date
  FROM `ukbb-analyses.ukbb7089_201910.phenotype` p
  JOIN `ukbb-analyses.ukbb7089_201910.dictionary` d USING(FieldID, coding_file_id)
  WHERE TRUE
    AND Path LIKE '%First occurrences%'
    AND Field LIKE 'Date%'
), first_sources AS (
  -- Source
  SELECT 
    p.sample_id, 
    p.FieldID, 
    -- Caution: brittle, expects disease code to always be in 0-based position 4
    SPLIT(Field, ' ')[OFFSET(4)] value,
    p.value code
  FROM `ukbb-analyses.ukbb7089_201910.phenotype` p
  JOIN `ukbb-analyses.ukbb7089_201910.dictionary` d USING(FieldID, coding_file_id)
  WHERE TRUE
    AND Path LIKE '%First occurrences%'
    AND Field LIKE 'Source%'
), linked AS (
  -- Link them both
  SELECT 
    fs.sample_id ,
    fs.FieldID source_FieldID,
    fd.FieldID date_FieldID,
    fs.code,
    SAFE.PARSE_DATE("%E4Y-%m-%d", fd.first_date) first_date
  FROM first_dates fd
  JOIN first_sources fs ON fd.sample_id = fs.sample_id AND fd.value = fs.value 
), first_dates_summary_fields AS (
  -- Allow people to use the FieldID of the date or of the source:
  SELECT source_FieldID FieldID, sample_id eid, code, '' meaning, first_date vdate FROM linked 
  UNION ALL
  SELECT date_FieldID FieldID, sample_id eid, code, '' meaning, first_date vdate FROM linked 
)

SELECT 
  diagnostics.eid sample_id, 
  diagnostics.FieldID, 
  diagnostics.code value, 
  MIN(vdate) first_date
FROM (
  SELECT * FROM dated_fields
  UNION DISTINCT
  SELECT * FROM dated_fields_fractional
  UNION DISTINCT
  SELECT * FROM self_reported_aged_subfields
  UNION DISTINCT 
  SELECT * FROM self_reported_undated_fields
  UNION DISTINCT
  SELECT * FROM first_dates_summary_fields
) diagnostics
WHERE TRUE
  AND vdate IS NOT NULL
GROUP BY diagnostics.eid, diagnostics.FieldID, diagnostics.code
;