#!/bin/bash

# Import the worker
gsutil -m cp -n gs://ukbb_v2/projects/jamesp/bin/prsworker.exe ~/
chmod +x ~/prsworker.exe

~/prsworker.exe \
    -bgen-template "${MOUNT_IMPUTED}/imputed/ukb_imp_chr%s_v3.bgen" \
    -input ${infile} \
    -layout ${layout} \
    -chromosome ${chrom} \
    -first_line ${firstline} \
    -source ${source} \
    -last_line ${lastline} > ${outfile}