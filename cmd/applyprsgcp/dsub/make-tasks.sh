PROJECT_PATH="gs://ukbb_v2/projects/jamesp/projects/gwas/lv20k/v42/prs"
VPJ=100000

gsutil -m cp ${PROJECT_PATH}/*.v42.results ./
gsutil cp "gs://ukbb_v2/projects/jamesp/bin/tasker.exe" ./
chmod +x ./tasker.exe

/bin/ls *.v42.results | head -n 1 | xargs -I {} bash -c "./tasker.exe -header=false -input '${PROJECT_PATH}/{}' -layout BOLTBGEN -output '${PROJECT_PATH}/output' -prs {} -variants_per_job ${VPJ}" | head -n 1 > tasks.tsv
/bin/ls *.v42.results | xargs -I {} bash -c "./tasker.exe -header=false -input '${PROJECT_PATH}/{}' -layout BOLTBGEN -output '${PROJECT_PATH}/output' -prs {} -variants_per_job ${VPJ} | tail -n +2" >> tasks.tsv