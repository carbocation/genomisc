WITH read2 AS (
  SELECT gp.eid, mr2i10.icd10_code, c.meaning, PARSE_DATE('%d/%m/%Y',event_dt) event_dt
  FROM `ukbb-analyses.ukbb7089_201910.gp_clinical` gp
  JOIN `ukbb-analyses.ukbb7089_201910.map_read_v2_icd10` mr2i10 ON mr2i10.read_code = gp.read_2
  JOIN `ukbb-analyses.ukbb7089_201910.dictionary` d ON d.FieldID = 41270
  JOIN `ukbb-analyses.ukbb7089_201910.coding` c ON c.coding_file_id = d.coding_file_id AND c.coding = mr2i10.icd10_code 
  WHERE TRUE
    -- For now, specify that we only want mappings that take a read2 code to a single ICD10 code
    AND mr2i10.icd10_code_def = 1
), read3 AS (
  SELECT gp.eid, mr3i10.icd10_code, c.meaning, PARSE_DATE('%d/%m/%Y',event_dt) event_dt
  FROM `ukbb-analyses.ukbb7089_201910.gp_clinical` gp
  JOIN `ukbb-analyses.ukbb7089_201910.map_read_v3_icd10` mr3i10 ON mr3i10.read_code = gp.read_3
  JOIN `ukbb-analyses.ukbb7089_201910.dictionary` d ON d.FieldID = 41270
  JOIN `ukbb-analyses.ukbb7089_201910.coding` c ON c.coding_file_id = d.coding_file_id AND c.coding = mr3i10.icd10_code 
  WHERE TRUE
    -- If it's mandatory to refine further, then we exclude
    AND mr3i10.refine_flag != 'M'
    -- For now, only permit exact one-to-one mappings ('E') or default mappings ('D')
    AND mr3i10.mapping_status IN ('E', 'D')
), gp_clinical_data AS (
  SELECT * FROM read2
  UNION ALL
  SELECT * FROM read3
  WHERE TRUE
)

-- For matching the materialized_hesin_dates format / production:
-- Treating all of these as a primary ICD10 diagnosis for now.
SELECT eid sample_id, 41202 FieldID, icd10_code value, MIN(event_dt) first_date
FROM gp_clinical_data
GROUP BY eid, icd10_code

-- For inspecting:
-- SELECT eid, icd10_code, meaning, MIN(event_dt) event_dt
-- FROM gp_clinical_data
-- GROUP BY eid, icd10_code, meaning
-- ORDER BY eid ASC, event_dt ASC