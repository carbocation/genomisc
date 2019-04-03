# Assumes you have generated the censor table and have placed it in google cloud storage
TABLE_LOC="gs://ml4cvd/projects/jamesp/bigquery/exome/gatk/tranche01/genotypes_alt"
DATASET="genetics"
LOC="US"
FILENAME="*.tsv.gz"

# For phenotypes, we expect to add repeatedly, so we don't replace here. Note:
# if you run append.sh with the same data twice, you'll just duplicate the
# contents of the table.
bq --location=${LOC} load \
 --replace \
 --field_delimiter "\t" \
 --quote "" \
 --source_format=CSV \
 --skip_leading_rows 1 \
 ${DATASET}.exome_genotype ${TABLE_LOC}/${FILENAME} \
 genotype.json