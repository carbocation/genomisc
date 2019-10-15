/bin/ls *.tsv | head -n 1 | xargs -I {} head -n1 {} > genomewide.loci
/bin/ls *.tsv | xargs -I {} ./genomewide-only.sh {} >> genomewide.loci
