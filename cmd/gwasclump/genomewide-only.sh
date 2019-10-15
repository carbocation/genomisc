>&2 echo "Processing ${1}"

awk 'NR > 1 && $16 < 5e-8 {print $0}' $1