# From https://ftp.ncbi.nih.gov/pub/clinvar/README_VCF.txt
Last updated October 13, 2017

README file for: 
ftp://ftp.ncbi.nih.gov/pub/clinvar/vcf_GRCh37/
ftp://ftp.ncbi.nih.gov/pub/clinvar/vcf_GRCh38/

These directories contain VCF files in the new format (2.0) for the ClinVar dataset.
The files use version 4.1 of the VCF specification (http://www.internationalgenome.org/wiki/Analysis/variant-call-format).
These files are generated monthly as part of ClinVar's regular release.

These files are archived in the following directories, with subdirectories for each year:
     - ftp://ftp.ncbi.nih.gov/pub/clinvar/vcf_GRCh37/archive_2.0
     - ftp://ftp.ncbi.nih.gov/pub/clinvar/vcf_GRCh38/archive_2.0

We welcome your feedback. Email us at clinvar@ncbi.nlm.nih.gov.


VCF files in the old format (1.0) for the Sept 2017 release and all prior releases are found in the archives:
ftp://ftp.ncbi.nih.gov/pub/clinvar/vcf_GRCh37/archive_1.0
ftp://ftp.ncbi.nih.gov/pub/clinvar/vcf_GRCh38/archive_1.0

VCF files for the dbSNP data set are found in the following directories, for GRCh37 and GRCh38 respectively:
ftp://ftp.ncbi.nlm.nih.gov/snp/organisms/human_9606_b150_GRCh37p13/VCF/
ftp://ftp.ncbi.nlm.nih.gov/snp/organisms/human_9606/VCF/

======================================================================================================================
FILES
======================================================================================================================
clinvar_vcf.GRCh37.vcf.gz	  ClinVar variants with precise endpoints. Each variant is represented by a single 
				  location in the reference assembly for GRCh37. This file pairs with the file 
				  clinvar_vcf.GRCh37.vcf_papu.gz below.

clinvar_vcf.GRCh37_papu.vcf.gz	  ClinVar variants with precise endpoints. Each variant is represented by all other mapped 
				  locations in GRCh37 not reported in the file above. This includes the pseudoautosomal region 
				  (PAR), alternate loci, patch sequences and unlocalized or unplaced contigs (papu). This file 
				  pairs with the  file clinvar_vcf.GRCh37.vcf.gz above.

clinvar_vcf.GRCh38.vcf.gz	  ClinVar variants with precise endpoints. Each variant is represented by a single 
				  location in the reference assembly for GRCh38. This file pairs with the file 
				  clinvar_vcf.GRCh38.vcf_papu.gz below.

clinvar_vcf.GRCh38_papu.vcf.gz	  ClinVar variants with precise endpoints. Each variant is represented by all other mapped 
				  locations in GRCh38 not reported in the file above. This includes the pseudoautosomal region 
				  (PAR), alternate loci, patch sequences and unlocalized or unplaced contigs (papu). This file 
				  pairs with the  file clinvar_vcf.GRCh38.vcf.gz above.

======================================================================================================================


======================================================================================================================
CHANGES MADE IN THE NEW FORMAT (2.0)
======================================================================================================================
allele-specific		Each row represents a single allele at that position, rather than one row per rs number.

comprehensive		The new VCF files shall be comprehensive for all variants in ClinVar. There shall be one file 
			per assembly for variants with precise endpoints on the genome and a second file per assembly for
			variants with imprecise endpoints on the genome. Note the file with imprecise endpoints is not
			available yet.

ClinVar-specific	The new VCF files include only variants that have been reported to ClinVar. They do not include
			other alleles at the same location that have been reported to dbSNP.

ID column		The ID column (col 3) reports the ClinVar Variation ID, rather than the rs number.

ambiguous bases		ClinVar accepts all IUPAC ambiguity codes for nucleotides. However, the VCF specification 
	  		(https://samtools.github.io/hts-specs/VCFv4.2.pdf) only allows ambiguity code N. Thus ClinVar XML 
			will retain the actual ambiguous bases, but all ambiguous values will be converted to N in the VCF 
			files.

included variants	Interpretations may be made on a single variant or a set of variants, such as a haplotype. Variants
	 		that have only been interpreted as part of a set of variants (i.e. no direct interpretation for the
			variant itself) are considered "included" variants. The VCF files include both variants with a direct
			interpretation and included variants. Included variants do not have an associated disease (CLNDN, 
			CLNDISDB) or a clinical significance (CLNSIG). Instead there are three tags are specific to the 
			included variants - CLNDNINCL, CLNDISDBINCL, and CLNSIGINCL (see below).

INFO tags		Data reported in the INFO tags is aggregated by Variation ID. INFO tags that are retained from 
     			the old format are CLNDN, CLNDISDB, CLNSIG, GENEINFO, RS, SSR.

AF_ESP, AF_EXAC, AF_TGP    The former AF INFO tag has been split into three tags, one for each source of allele 
            	 	   frequency data. AF_ESP reports allele frequency from GO-ESP; AF_EXAC from the ExAC Consortium; and 
            		   AF_TGP from the 1000 Genomes Project.

ALLELEID		A new INFO tag, ALLELEID, reports the Allele ID for the variant.

CLNDNINCL		A new INFO tag, CLNDNINCL, reports ClinVar's preferred disease name for an interpretation for a
            		haplotype or genotype that includes this variant.

CLNDISDBINCL		A new INFO tag, CLNDNINCL, reports the database name and identifier for the disease name for an
			interpretation for a haplotype or genotype that includes this variant. Multiples are separated 
			by a pipe.

CLNHGVS			A new INFO tag, CLNHGVS, reports the top-level genomic HGVS expression for the variant. This may be on 
			an accession for the primary assembly or on an ALT LOCI.

CLNSIGINCL		A new INFO tag, CLNSIGINCL, reports the clinical significance of a haplotype or genotype that 
			includes this variant. It is reported as pairs of Variation ID for the haplotype or genotype 
			and the corresponding clininical significance.

CLNVI			A new INFO tag, CLNVI, reports identifiers for the variant in other databases, e.g. OMIM 
			Allelic variant IDs.

CLNVC			A new INFO tag, CLNVC, reports the type of variation.

MC			A new INFO tag, MC, report the predicted molecular consequence of the variant. It is reported as 
			pairs of the Sequence Ontology (SO) identifier and the molecular consequence term joined by a 
			vertical bar. Multiple values are separated by a comma. This tag replaces ASS, DSS, INT, NSF, NSM,
			NSN, R3, R5, SYN, U3, and U5 in the old format.

ORIGIN			A new INFO tag, ORIGIN, reports an integer representing the allele origins that have been observed 
			for the variant and reported to ClinVar. One or more of the values may be added: 0 - unknown; 
			1 - germline; 2 - somatic; 4 - inherited; 8 - paternal; 16 - maternal; 32 - de-novo; 64 - biparental; 
			128 - uniparental; 256 - not-tested; 512 - tested-inconclusive; 1073741824 - other. This tag replaces 
			CLNORIGIN in the old format. 



======================================================================================================================