NAME="v42"

PROJECT_PATH="gs://ukbb_v2/projects/jamesp/projects/gwas/lv20k/v42/prs"

BILLING_PROJECT="ukbb-analyses"

MAX_PREEMPTION=10

# Should not need to edit:

DOCKER_IMAGE="gcr.io/ukbb-analyses/saige:0.35.8.3-jpp"

# Launch step 1 and block until completion
echo "Launching PRS processors with up to ${MAX_PREEMPTION} preemptible attempts"
if dsub \
    --project ${BILLING_PROJECT} \
    --provider google-v2 \
    --use-private-address \
    --regions us-central1 us-east1 us-west1 \
    --disk-type pd-standard \
    --disk-size 10 \
    --min-cores 2 \
    --min-ram 4 \
    --image ${DOCKER_IMAGE} \
    --preemptible \
    --retries ${MAX_PREEMPTION} \
    --skip \
    --wait \
    --logging ${PROJECT_PATH}/dsub-logs \
    --mount MOUNT_IMPUTED="gs://fc-7d5088b4-7673-45b5-95c2-17ae00a04183" \
    --tasks tasks.tsv \
    --name PRS-${NAME} \
    --script worker.sh; then
   exit 0
else
    :
fi

echo "Launching BOLT again but without preemption, to complete any unfinished tasks"
dsub \
    --project ${BILLING_PROJECT} \
    --provider google-v2 \
    --use-private-address \
    --regions us-central1 us-east1 us-west1 \
    --disk-type pd-standard \
    --disk-size 10 \
    --min-cores 2 \
    --min-ram 4 \
    --image ${DOCKER_IMAGE} \
    --skip \
    --wait \
    --logging ${PROJECT_PATH}/dsub-logs \
    --mount MOUNT_IMPUTED="gs://fc-7d5088b4-7673-45b5-95c2-17ae00a04183" \
    --tasks tasks.tsv \
    --name PRS-${NAME} \
    --script worker.sh
