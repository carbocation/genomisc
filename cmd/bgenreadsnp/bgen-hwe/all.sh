# Computes HWE for everything
#/bin/ls ~/shortcuts/ukbb/imputed_v3/ukb_imp_chr[0-9]*_v3.bgen | xargs -P 22 -I {} bash -c 'bgen-hwe.exe -bgen {} > hwe.`basename {}`.tsv'

# Computes the HWE for the lv20k v12 ~16.5k sample subset
/bin/ls /broad/ukbb/imputed_v3/ukb_imp_chr[0-9]*_v3.bgen | xargs -P 22 -I {} bash -c 'bgen-hwe.exe -bgen {} -sample /medpop/esp2/pradeep/UKBiobank/v3data/ukb7089_imp_chr15_v3_s487395.sample -sample_ids /medpop/esp/jamesp/projects/lv20k/data/kept.samples > hwe.`basename {}`.tsv'

# Just look up one SNP
bgen-hwe.exe -snp "17:70053954_AAGAGAGAGAGAGAGAG_A" -bgen /broad/ukbb/imputed_v3/ukb_imp_chr17_v3.bgen -sample /medpop/esp2/pradeep/UKBiobank/v3data/ukb7089_imp_chr15_v3_s487395.sample -sample_ids /medpop/esp/jamesp/projects/lv20k/data/kept.samples

# Combine the first N values into one file for processing in R
head -n 1 hwe.ukb_imp_chr13_v3.bgen.tsv > subset.hwe.tsv
/bin/ls hwe*.tsv | xargs -I {} bash -c 'tail -n +2 {} | head -n 450 ' | sort -k 2 -g >> subset.hwe.tsv
