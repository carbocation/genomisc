WITH dated_fields AS (
  SELECT p.FieldID, p.sample_id eid, p.value code, cod.meaning,
    CASE
      WHEN SAFE.PARSE_DATE("%E4Y-%m-%d", d.value) IS NULL THEN SAFE.PARSE_DATE("%E4Y-%m-%d", denroll.value)
      WHEN cod.meaning LIKE ('%unknown%') THEN SAFE.PARSE_DATE("%E4Y-%m-%d", denroll.value)
      ELSE SAFE.PARSE_DATE("%E4Y-%m-%d", d.value)
    END vdate
  FROM `ukbb-analyses.ukbb7089_201909.phenotype` p
  JOIN `ukbb-analyses.ukbb7089_201909.phenotype` denroll ON denroll.FieldID=53 AND denroll.sample_id=p.sample_id AND denroll.instance = 0 AND denroll.array_idx = 0
  JOIN `ukbb-analyses.ukbb7089_201909.phenotype` d ON d.sample_id=p.sample_id AND d.instance = p.instance AND d.array_idx = p.array_idx
    AND (
      FALSE
      -- p for phenotype field, d for date field
      OR (p.FieldID=42013 AND d.FieldID=42012)
      OR (p.FieldID=42011 AND d.FieldID=42010)
      OR (p.FieldID=42009 AND d.FieldID=42008)
      OR (p.FieldID=42007 AND d.FieldID=42006)
      OR (p.FieldID=42001 AND d.FieldID=42000)
      OR (p.FieldID=40021 AND d.FieldID=40005)
    )
  LEFT JOIN `ukbb-analyses.ukbb7089_201909.coding` cod ON cod.coding_file_id = d.coding_file_id AND cod.coding = d.value
), 
dated_fields_fractional AS (
  SELECT p.FieldID, p.sample_id eid, p.value code, cod.meaning,
  CASE
      WHEN SAFE.PARSE_DATE("%Y", d.value) IS NULL THEN SAFE.PARSE_DATE("%E4Y-%m-%d", denroll.value)
      WHEN cod.meaning LIKE ('%unknown%') THEN SAFE.PARSE_DATE("%E4Y-%m-%d", denroll.value)
      ELSE SAFE.PARSE_DATE("%Y", d.value)
    END vdate
  FROM `ukbb-analyses.ukbb7089_201909.phenotype` p
  JOIN `ukbb-analyses.ukbb7089_201909.phenotype` denroll ON denroll.FieldID=53 AND denroll.sample_id=p.sample_id AND denroll.instance = 0 AND denroll.array_idx = 0
  JOIN `ukbb-analyses.ukbb7089_201909.phenotype` d ON d.sample_id=p.sample_id AND d.instance = p.instance AND d.array_idx = p.array_idx
    AND (
      FALSE
      OR (p.FieldID=20004 AND d.FieldID=20010)
      OR (p.FieldID=20002 AND d.FieldID=20008)
      OR (p.FieldID=20001 AND d.FieldID=20006)
    )
  LEFT JOIN `ukbb-analyses.ukbb7089_201909.coding` cod ON cod.coding_file_id = d.coding_file_id AND cod.coding = d.value
)

SELECT 
  diagnostics.eid sample_id, diagnostics.FieldID, diagnostics.code value, MIN(vdate) first_date
FROM (
  SELECT * FROM dated_fields
  UNION DISTINCT
  SELECT * FROM dated_fields_fractional
) diagnostics
WHERE TRUE
  AND vdate IS NOT NULL
GROUP BY diagnostics.eid, diagnostics.FieldID, diagnostics.code
;