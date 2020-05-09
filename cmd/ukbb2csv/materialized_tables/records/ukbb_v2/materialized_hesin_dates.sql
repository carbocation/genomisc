WITH oper4_primary AS (
  SELECT 
    41200 FieldID, 
    sec.eid, 
    oper4 code, 
    CASE 
      WHEN sec.opdate IS NOT NULL THEN PARSE_DATE('%d/%m/%Y', sec.opdate)
      WHEN h.epistart IS NOT NULL THEN PARSE_DATE('%Y%m%d', h.epistart)
      ELSE PARSE_DATE('%Y%m%d', h.admidate)
    END vdate
  FROM `ukbb-analyses.ukbb7089_201909.hesin_oper` sec 
  JOIN `ukbb-analyses.ukbb7089_201909.hesin` h ON sec.eid=h.eid AND sec.ins_index=h.ins_index
  WHERE TRUE
    AND sec.level = 1
    AND sec.oper4 IS NOT NULL
), diag_icd10_primary AS (
  SELECT 
    41202 FieldID, 
    sec.eid, 
    diag_icd10 code, 
    CASE 
      WHEN h.epistart IS NOT NULL THEN PARSE_DATE('%Y%m%d', h.epistart)
      ELSE PARSE_DATE('%Y%m%d', h.admidate)
    END vdate
  FROM `ukbb-analyses.ukbb7089_201909.hesin_diag` sec 
  JOIN `ukbb-analyses.ukbb7089_201909.hesin` h ON sec.eid=h.eid AND sec.ins_index=h.ins_index
  WHERE TRUE
    AND sec.level = 1
    AND sec.diag_icd10 IS NOT NULL
), diag_icd9_primary AS (
  SELECT 
    41203 FieldID, 
    sec.eid, 
    diag_icd9 code, 
    CASE 
      WHEN h.epistart IS NOT NULL THEN PARSE_DATE('%Y%m%d', h.epistart)
      ELSE PARSE_DATE('%Y%m%d', h.admidate)
    END vdate
  FROM `ukbb-analyses.ukbb7089_201909.hesin_diag` sec 
  JOIN `ukbb-analyses.ukbb7089_201909.hesin` h ON sec.eid=h.eid AND sec.ins_index=h.ins_index
  WHERE TRUE
    AND sec.level = 1
    AND sec.diag_icd9 IS NOT NULL
), oper4_secondary AS (
  SELECT 
    41210 FieldID, 
    sec.eid, 
    oper4 code, 
    CASE 
      WHEN sec.opdate IS NOT NULL THEN PARSE_DATE('%d/%m/%Y', sec.opdate)
      WHEN h.epistart IS NOT NULL THEN PARSE_DATE('%Y%m%d', h.epistart)
      ELSE PARSE_DATE('%Y%m%d', h.admidate)
    END vdate
  FROM `ukbb-analyses.ukbb7089_201909.hesin_oper` sec 
  JOIN `ukbb-analyses.ukbb7089_201909.hesin` h ON sec.eid=h.eid AND sec.ins_index=h.ins_index
  WHERE TRUE
    AND sec.level > 1
    AND sec.oper4 IS NOT NULL
), diag_icd10_secondary AS (
  SELECT 
    41204 FieldID, 
    sec.eid, 
    diag_icd10 code, 
    CASE 
      WHEN h.epistart IS NOT NULL THEN PARSE_DATE('%Y%m%d', h.epistart)
      ELSE PARSE_DATE('%Y%m%d', h.admidate)
    END vdate
  FROM `ukbb-analyses.ukbb7089_201909.hesin_diag` sec 
  JOIN `ukbb-analyses.ukbb7089_201909.hesin` h ON sec.eid=h.eid AND sec.ins_index=h.ins_index
  WHERE TRUE
    AND sec.level > 1
    AND sec.diag_icd10 IS NOT NULL
), diag_icd9_secondary AS (
  SELECT 
    41205 FieldID, 
    sec.eid, 
    diag_icd9 code, 
    CASE 
      WHEN h.epistart IS NOT NULL THEN PARSE_DATE('%Y%m%d', h.epistart)
      ELSE PARSE_DATE('%Y%m%d', h.admidate)
    END vdate
  FROM `ukbb-analyses.ukbb7089_201909.hesin_diag` sec 
  JOIN `ukbb-analyses.ukbb7089_201909.hesin` h ON sec.eid=h.eid AND sec.ins_index=h.ins_index
  WHERE TRUE
    AND sec.level > 1
    AND sec.diag_icd9 IS NOT NULL
)

SELECT 
  diagnostics.eid sample_id, 
  diagnostics.FieldID, 
  diagnostics.code value, 
  CASE 
    WHEN MIN(vdate) IS NULL THEN MIN(PARSE_DATE("%E4Y-%m-%d", p.value))
    ELSE MIN(vdate)
  END first_date
FROM (
  SELECT * FROM oper4_primary
  UNION DISTINCT
  SELECT * FROM diag_icd10_primary
  UNION DISTINCT
  SELECT * FROM diag_icd9_primary
  UNION DISTINCT
  SELECT * FROM oper4_secondary
  UNION DISTINCT
  SELECT * FROM diag_icd10_secondary
  UNION DISTINCT
  SELECT * FROM diag_icd9_secondary
) diagnostics
JOIN `ukbb-analyses.ukbb7089_201909.phenotype` p ON p.sample_id = diagnostics.eid AND p.array_idx=0 AND p.instance=0 AND p.FieldID=53
GROUP BY diagnostics.eid, diagnostics.FieldID, diagnostics.code
ORDER BY first_date ASC
