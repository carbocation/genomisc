# Mendeloverlap

Mendel overlap helps assess the degree to which your GWAS loci overlap with a
Mendelian gene list beyond what would be expected by chance. It requires that
you provide a gene list (just a file with each gene name on its own line) and
the output from [SNPsnap](https://data.broadinstitute.org/mpg/snpsnap/). SNPsnap
is a tool that takes in your GWAS SNPs and returns similar SNPs (in terms of
allele frequency, gene density, etc) to make for a fair permutation. The tool
then tests how many genes are within a radius from your real loci and your
SNPsnap permutation loci, providing a P-value.

Once mendeloverlap is installed (`go install`), it can be run as follows:

```sh
mendeloverlap \
  -snpsnap snpsnap_output.txt \
  -radius 250 \
  -mendel cardiomyopathy_genes.csv
```

Run it without arguments to see the different toggles that are available.
