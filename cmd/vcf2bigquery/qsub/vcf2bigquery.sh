~/bin/vcf2bigquery.exe \
  --vcf ~/esp2/data/ukbb/exome/tranche1/tranche01.gatk.vcf.gz \
  --assembly grch38 \
  --chunksize ${CHUNK_SIZE} \
  --alt \
  --chunk ${SGE_TASK_ID} | gzip -c