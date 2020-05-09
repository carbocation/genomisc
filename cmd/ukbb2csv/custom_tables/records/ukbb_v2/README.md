# Hospital Episode Statistics (HES) and general practice (GP) event records

These are available through an SQL interface from the Downloads tab in the UK Biobank showcase.

Table names:

* HES (Note: there are others, these are the ones I use)
  * hesin
  * hesin_oper
  * hesin_diag
* GP
  * gp_clinical
  * gp_registrations (not used)
  * gp_scripts (not used)


Download the full tables from the table download tool and import into BigQuery.

Note that the `map_read_v2_icd10` and `map_read_v3_icd10` tables required to materialize `gp_clinical` are 
from [spreadsheets that are provided courtesy of the UK Biobank](http://biobank.ndph.ox.ac.uk/showcase/refer.cgi?id=592).
