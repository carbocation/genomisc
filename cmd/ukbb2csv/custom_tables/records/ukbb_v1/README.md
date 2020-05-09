# ~~Hospital Episode Statistics (HES) and general practice (GP) event records~~

*This approach is deprecated, please see the v2 folder*

These are available through an SQL interface from the Downloads tab in the UK Biobank showcase.

Table names:

* HES (Note: there are others, these are the ones we use)
  * HESIN
  * HESIN_OPER
  * HESIN_DIAG
* GP
  * gp_registrations
  * gp_scripts
  * gp_clinical


In v1, when fetching HESIN_DIAG data, I would split it into diag_icd9 and diag_icd10 tables:

ICD9: `SELECT eid, ins_index, arr_index, level, diag_icd9, diag_icd9_nb FROM hesin_diag WHERE diag_icd9 IS NOT NULL`

ICD10: `SELECT eid, ins_index, arr_index, level, diag_icd10, diag_icd10_nb FROM hesin_diag WHERE diag_icd10 IS NOT NULL`

For the others, I would download the full table.
