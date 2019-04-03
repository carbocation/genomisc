mkdir -p ~/hptmp/exome/linearize/missingonly/errors/

# Chunksize 500 => 5760 chunks

qsub \
  -cwd \
  -sync ry \
  -t 1-5760:1 \
  -l h_vmem=2g \
  -binding linear:2 -pe smp 2 \
  -l h_rt="02:00:00" \
  -v CHUNK_SIZE=500 \
  -o ~/hptmp/exome/linearize/missingonly/linear.\$TASK_ID.tsv.gz \
  -e ~/hptmp/exome/linearize/missingonly/errors/error.\$TASK_ID \
  vcf2bigquery.sh