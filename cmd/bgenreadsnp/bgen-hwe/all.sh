# Computes HWE for everything
/bin/ls ~/shortcuts/ukbb/imputed_v3/ukb_imp_chr[0-9]*_v3.bgen | xargs -P 22 -I {} bash -c 'bgen-hwe.exe -bgen {} > hwe.`basename {}`.tsv'

# Combine the first 8800 values into one file for processing in R
head -n 1 hwe.ukb_imp_chr13_v3.bgen.tsv > subset.hwe.tsv
/bin/ls *.tsv | xargs -I {} bash -c 'tail -n +2 {} | head -n 400 ' | sort -k 2 -g >> subset.hwe.tsv
