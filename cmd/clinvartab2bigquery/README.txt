# From https://ftp.ncbi.nlm.nih.gov/pub/clinvar/tab_delimited/README
Last updated February 10, 2020
README file for ftp://ftp.ncbi.nih.gov/pub/clinvar

This directory contains reports from the ClinVar dataset and documents about ClinVar development. Sections in this README are divided by type of content. 

This directory has a folder for documents related to the collaboration with ClinGen (http://www.clinicalgenome.org/). For more details, see http://www.ncbi.nlm.nih.gov/clinvar/docs/review_guidelines.

This README file also documents ClinVar-related data in other directories, such as 
        ftp://ftp.ncbi.nlm.nih.gov/pub/GTR/standard_terms, for terminology used by both GTR and ClinVar.
        ftp://ftp.ncbi.nlm.nih.gov/gene/DATA/mim2gene_medgen, for gene-disease relationships related to OMIM

http://www.ncbi.nlm.nih.gov/clinvar/


================================================================================
SUBMISSIONS
================================================================================
  Details about how to submit are provided at this site:
       http://www.ncbi.nlm.nih.gov/clinvar/docs/submit/

--------------------------------------------------
clinvar_submission.xsd
--------------------------------------------------
  
  If you have any questions about submitting data to ClinVar as an XML file, please contact us via clinvar@ncbi.nlm.nih.gov
  This is the .xsd file for validating records submitted to ClinVar as xml.

------------------------------------------------------------
history of the versions of clinvar_submission.xsd
subdirectory xsd_submission
------------------------------------------------------------
This subdirectory archives different versions of the submission xsd.  The current one is also accessible as clinvar_submission.xsd


--------------------------------------------------
submission_templates 
subdirectory ftp://ftp.ncbi.nih.gov/pub/clinvar/submission_templates
--------------------------------------------------
  This subdirectory contains excel spreadsheets that can be used to 
  submit data to ClinVar. 
 
    There are two templates in this directory:

         SubmissionTemplate.xlsx
         SubmissionTemplateLite.xlsx

    SubmissionTemplate.xlsx is the standard submission template.

    SubmissionTemplateLite.xlsx is simpler, and is designed for
    submissions with less supporting evidence.

================================================================================
EXTRACTS OF CLINVAR DATA
================================================================================
ClinVar data are provided for download as extracts in xml, vcf and tab-delimited 
formats in the directories described below.

-------------------------------------------------
Updates
-------------------------------------------------
Data on the ftp site are updated monthly and weekly.
Please note on each file how when the data are refreshed.

Monthly:  usually the first Thursday of the month.
          A copy is also made to the archive sub-directory.

Weekly:   usually on Mondays
          XML: Weekly releases are not archived, and accumulate in the weekly_release subdirectory until the next monthly release.
			 tab-delimited: The latest version is retained in the tab-delimited path.
			                The archive subdirectory retains the copy consistent with the monthly release.

-------------------------------------------------
clinvar_public.xsd
-------------------------------------------------
The schema for the export version of the XML 
clinvar_public.xsd		Link to the current version such as /xsd_public/clinvar_public_1.5.xsd

------------------------------------------------------------
history of the versions of clinvar_submission.xsd
subdirectory xsd_public
------------------------------------------------------------
This subdirectory archives different versions of xsd used to validate ClinVar's comprehensive export as xml.  The current one is also accessible as clinvar_submission.xsd

The version number is represented in the file name.

----------------------------------------
release_notes
----------------------------------------
The release_notes subdirectory  contains reports of the differences between versions of clinvar_public.xsd


================================================================================
NAMES OF PHENOTYPES
================================================================================

--------------------------------------------------
disease_names
--------------------------------------------------

  This document is updated daily, and is provided to report the preferred names and
  identifiers used in GTR and ClinVar. Please note there may be more than one
  line per condition, when a name is used by more than one source. This
  differs from the gene_condition_source_id file because it is comprehensive,
  and does not require knowledge of any gene-to-disease relationship.
    NOTE: in February, 2020, the scope of the sources being reported was modified.
	 In particular, specific submitters and historical references to GeneTests were removed.
	 Attributing to GARD and SNOMED CT were suspended until maintenance is more up to date.
	 
  Tab-delimited file with the following 7 fields:

DiseaseName:          The name preferred by GTR and ClinVar
SourceName:           Sources that also use this preferred name
ConceptID:            The identifier assigned to a disorder associated with this
                        gene. If the value starts with a C and is followed by digits,
                        the ConceptID is a value from UMLS; if a value begins with CN,
                        it was created by NCBI-based processing.
SourceID:             Identifier used by the source reported in column 2 (SourceName)
DiseaseMIM:           MIM number for the condition.
LastUpdated:          Last time this record was modified by NCBI staff
Category:             Category of disease (as reported in ClinVar's XML), one of:
                        - Blood group
                        - Disease
                        - Finding
                        - Named protein variant
                        - Pharmacological response
                        - phenotype instruction


--------------------------------------------------
gene_condition_source_id
--------------------------------------------------

  This document is updated daily, and is provided to report gene-disease relationships used in ClinVar, Gene, GTR and MedGen.
  The sources of information for the gene-disease relationship include OMIM, GeneReviews, and a limited amount of curation by NCBI staff.
  The scope of disorders reported in this file is a subset of the disease_names file because a gene-to-disease relationship is required.  
  Tab-delimited file with the following fields:

GeneID:               The NCBI GeneID
GeneSymbol:           The preferred symbol corresponding to the GeneID
ConceptID:            The identifier assigned to a disorder associated with this
                        gene. If the value starts with a C and is followed by digits,
                        the ConceptID is a value from UMLS; if a value begins with
                        CN, it was created by NCBI-based processing
DiseaseName:          Full name for the condition
SourceName:           Sources that use this name
SourceID:             The identifier used by this source
DiseaseMIM:           MIM number for the condition
LastUpdated:          Last time this record was modified by NCBI staff

--------------------------------------------------
ConceptID_history.txt 
--------------------------------------------------
  --Added to the directory October 24, 2012
  
  This document is updated daily, and is provided to help track changes in
identifiers assigned to phenotypes over time. The ConceptID values in the
first column are no longer active, and are either discontinued (the value in
column 2 is 'No longer reported', or replaced by a record with a different
identifier.  That replacement may result either because of a merge (one record
becoming secondary to another) or because of a change in numbering, usually
because an identifier assigned by NCBI (starting with CN) is now thought to be
represented by a ConceptID from UMLS (starting with C followed by numerals).

Previous ConceptID 			the outdated identifier
Current ConceptID				the current identifier
Date of Action					the date this change occurred


--------------------------------------------------
dbGaP_frequency_study_list 
--------------------------------------------------
  --Added to the directory September 29, 2015

Text and html files reporting the studies in dbGaP that were assessed for single nucleotide variants reported in ClinVar.
dbGaP_frequency_study_list.html
dbGaP_frequency_study_list.txt

================================================================================
SUBDIRECTORIES
================================================================================
community		 files generated in the initial design of ClinVar
presentations	 slides or other documents about ClinVar
submission_templates	  templates for submission by spreadsheet
tab_delimited			  flattened tabular data summaries of several types
-----------------------VCF------------------------
See the README specific to ClinVar's VCF files:
ftp://ftp.ncbi.nlm.nih.gov/pub/clinvar/vcf_GRCh37/README_VCF.txt
ftp://ftp.ncbi.nlm.nih.gov/pub/clinvar/vcf_GRCh38/README_VCF.txt
--------------------------------------------------
xml				 An extraction of data in ClinVar as xml
The xsd for the export version of the XML is clinvar_public.xsd
For more details about the files in the xml directory, please refer to 
ftp://ftp.ncbi.nlm.nih.gov/pub/clinvar/xml/_README


==================================================
tab_delimited sub-directory

ClinVar makes every attempt to retain backward compatibility in its information products. 
If changes are made to these tab-delimited files, new columns are usually added to the end. 
If, however, a column needs to be removed, the names of the remaining column headers will be stable. 

Most of the files in this directory are generated weekly, usually on Mondays.
Not all of the files in this directory are archived, so please note the for each file whether
  monthly versions are copies to the archive subdirectory the first of each month.

--------------------------------------------------------------------------------
1. gene_specific_summary
--------------------------------------------------------------------------------
Generated weekly
Archived monthly ( first Thurday of each month)

Although this report is generated each week, it is currently based on statistics that are captured the first day of each month.
Therefore there will be some discrepancies between what is reported in this file and what may be viewed interactively on the web.

A tab-delimited report, for each gene, of the number of submissions and the number of different variants (alleles).
Because some variant-gene relationships are submitted, and some are calculated from overlapping annotation, in January of 2015, the report was modified to indicate when the gene-variant relationship was submitted.

Symbol                 Gene symbol (if officially named, from HGNC, else from NCBI's Gene database)                
GeneID                 Unique identifier from  NCBI's Gene database
Total_submissions		  Total submissions to ClinVar with variants in/overlapping this gene
Total_alleles          Number of alleles submitted to ClinVar for this gene
Submissions_reporting_this_gene
                       Subset of the total submissions that also reported the gene
Alleles_reported_Pathogenic_Likely_pathogenic
                       Number of variants reported as pathogenic or likely pathogenic
                       Excludes structural variants that may overlap a gene
Gene_MIM_Number        The MIM number for this gene
Number_Uncertain       Submissions with an interpretation of 'Uncertain significance'
Number_with_conflicts  Number of VariationIDs for this gene with conflicting interpretations
--------------------------------------------------------------------------------
2. variant_summary.txt
--------------------------------------------------------------------------------
Generated weekly
Archived monthly (first Thurday of each month)


A tab-delimited report based on each variant at a location on the genome for which data have been submitted to ClinVar.  
The data for the variant are reported for each assembly, so most variants have a line for GRCh37 (hg19) and another line for GRCh38 (hg38).
 
Please note: Beginning in October 2016, this file was modified to restrict reporting to attributes of an AlleleID, not a mixture of AlleleID and VariationID.  The modifications were announced in our September 2016 release notes:  ftp://ftp.ncbi.nlm.nih.gov/pub/clinvar/release_notes/20160901_data_release_notes.pdf.

             The last file that reported VariationID was ftp://ftp.ncbi.nlm.nih.gov/pub/clinvar/tab_delimited/archive/variant_summary_2016-09.txt.gz
             The first with the new set of columns is ftp://ftp.ncbi.nlm.nih.gov/pub/clinvar/tab_delimited/archive/variant_summary_2016-10.txt.gz
             Content that used to be in this file may be found in 
                    ftp://ftp.ncbi.nlm.nih.gov/pub/clinvar/tab_delimited/hgvs4variation.txt.gz
                    ftp://ftp.ncbi.nlm.nih.gov/pub/clinvar/tab_delimited/variation_allele.txt.gz

Please note: Beginning in November 2019, the values for referenceAllele and alternateAllele are being written according to the VCF
             standard.  For single nucleotide variants there was no change in the value.

             See also the authoritative file for identifiers assigned to genes represented by NCBI, namely:
                    ftp://ftp.ncbi.nih.gov/gene/DATA/GENE_INFO/Mammalia/Homo_sapiens.gene_info.gz

AlleleID               integer value as stored in the AlleleID field in ClinVar  (//Measure/@ID in the XML)
Type                   character, the type of variant represented by the AlleleID
Name                   character, ClinVar's preferred name for the record with this AlleleID
GeneID                 integer, GeneID in NCBI's Gene database, reported if there is a single gene, otherwise reported as -1.
GeneSymbol             character, comma-separated list of GeneIDs overlapping the variant
HGNC_ID                string, of format HGNC:integer, reported if there is a single GeneID. Otherwise reported as '-'
ClinicalSignificance   character, comma-separated list of aggregate values of clinical significance calculated for this variant
                       If the value is 'no interpretation for the single variant', this allele was submitted to
                       ClinVar as part of a haplotype or genotype, and its specific significance has not been submitted to ClinVar.
                       NOTE: Now that the aggregate values of clinical significance give precedence to records
                             with assertion criteria and evidence, the values in this column may appear to be in
                             conflict with the value reported in ClinSigSimple. 
ClinSigSimple          integer, 0 = no current value of Likely pathogenic or Pathogenic
                                1 = at least one current record submitted with an interpretation of Likely pathogenic or
                                    Pathogenic (independent of whether that record includes assertion criteria and evidence).
                               -1 = no values for clinical significance at all for this variant or set of variants; used for
                                    the "included" variants that are only in ClinVar because they are included in a
                                    haplotype or genotype with an interpretation
                       NOTE: Now that the aggregate values of clinical significance give precedence to records with
                             assertion criteria and evidence, the values in this column may appear to be in conflict with the
                             value reported in ClinicalSignificance.  In other words, if a submission without assertion criteria and
                             evidence interpreted an allele as pathogenic, and those with assertion criteria and evidence interpreted
                             as benign, then ClinicalSignificance would be reported as Benign and ClinSigSimple as 1.
LastEvaluated          date, the latest date any submitter reported clinical significance
RS# (dbSNP)            integer, rs# in dbSNP, reported as -1 if missing
nsv/esv (dbVar)        character, the NSV identifier for the region in dbVar
RCVaccession           character, list of RCV accessions that report this variant
PhenotypeIDs           character, list of identifiers for phenotype(s) interpreted for this variant
PhenotypeList          character, list of names corresponding to PhenotypeIDs
Origin                 character, list of all allelic origins for this variant
OriginSimple           character, processed from Origin to make it easier to distinguish between germline and somatic
Assembly               character, name of the assembly on which locations are based  
ChromosomeAccession    Accession and version of the RefSeq sequence defining the position reported in the start and stop columns. 
                            Please note some of these accessions may be for sub-chromosomal regions.
Chromosome             character, chromosomal location
Start                  integer, starting location, in pter->qter orientation
Stop                   integer, end location, in pter->qter orientation
ReferenceAllele        The reference allele according to the vcf standard.
AlternateAllele        The alternate allele according to the vcf standard.
Cytogenetic            character, ISCN band
ReviewStatus           character, highest review status for reporting this measure. For the key to the terms, 
                           and their relationship to the star graphics ClinVar displays on its web pages, 
                           see http://www.ncbi.nlm.nih.gov/clinvar/docs/variation_report/#interpretation
									Note also that 'no interpretation for the single variant' is used for AlleleIDs in ClinVar
									that were submitted as part of the definition of a complex allele, but not interpreted
									individually.
NumberSubmitters       integer, number of submitters describing this variant
Guidelines             character, ACMG only right now, for the reporting of incidental variation in a Gene 
                       enumerates whether the guideline is from 2013 (ACMG2013, PubMed 23788249) or 2016 (ACMG2016, PubMed 27854360)
                               (NOTE: if ACMG, not a specific to the AlleleID but to the Gene in which the AlleleID is found)
TestedInGTR            character, Y/N for Yes/No if there is a test registered as specific to this variant 
                          in the NIH Genetic Testing Registry (GTR)
OtherIDs               character, list of other identifiers or sources of information about this variant
SubmitterCategories    coded value to indicate whether data were submitted by another resource (1), any other type of source (2) or both (3)
VariationID            The identifier ClinVar uses specific to the AlleleID.  Not all VariationIDS that may be related to
                           the AlleleID are reported in this file. For a comprehensive mapping of AlleleID to VariationID,
									please use ftp://ftp.ncbi.nlm.nih.gov/pub/clinvar/tab_delimited/variation_allele.txt.gz.
							      Note also that some of the values for VariationID are not supported in the current
									default web display, but will be reported from ClinVar 2.0 as current seen from our preview site.
							  
                       
--------------------------------------------------------------------------------
3. cross_references.txt
--------------------------------------------------------------------------------
Generated weekly
Not archived

A tab-delimited report based on each variant in ClinVar, providing identifiers in other databases and when those data were last updated. This file is based on AlleleID rather than VariationID for complex alleles, so it corresponds to a unique genomic location.
NOTE: this file is preliminary and currently includes only identifiers in dbSNP and dbVar. Identifiers from more databases will be added in the future.

AlleleID		 	 		  integer value as stored in the AlleleID field in ClinVar  (//Measure/@ID in the XML)
Database					  name of the database
ID							  identifier used by that database
last_updated			  date the identifier /AlleleID relationship was created or last updated

--------------------------------------------------------------------------------
4. var_citations.txt
--------------------------------------------------------------------------------
Generated weekly
Not archived

A tab-delimited report of citations associated with data in ClinVar, connected to the AlleleID, the VariationID, and either rs# from dbSNP or nsv in dbVar.

AlleleID                       integer value as stored in the AlleleID field in ClinVar  (//Measure/@ID in the XML)
VariationID                    The identifier ClinVar uses to anchor its default display. (in the XML,  //MeasureSet/@ID)
rs										 rs identifier from dbSNP, null if missing
nsv									 nsv identifier from dbVar, null if missing
citation_source					 The source of the citation, either PubMed, PubMedCentral, or the NCBI Bookshelf
citation_id							 The identifier used by that source

--------------------------------------------------------------------------------
5. summary_of_conflicting_interpretations.txt
--------------------------------------------------------------------------------
Generated weekly
Archived monthly ( first Thurday of each month)

This file first became available in January, 2016. It replaces summary_of_conflicting_data.txt (documented below) and differs in that
   a. It is limited to differences in interpretation (i.e. does not report differences in the phenotype being interpreted)
   b. Reports all pairwise differences, so that if submitter a differs from submitters b and c, and submitter b differs from c,  a-b, a-c, b-c will all be reported instead of just a-b and a-c
   c. Reports fewer columns

Gene_Symbol                          If in a gene, its symbol
NCBI_Variation_ID                    The identifier ClinVar uses to anchor its default display. (in the XML,  //MeasureSet/@ID)
ClinVar_Preferred                    The preferred description ClinVar uses for this VariationID
Submitter1                           Name of this submitter
Submitter1_SCV  							 Accession assigned to this submission
Submitter1_ClinSig                   Clinical signficance asserted by this submitter
Submitter1_LastEval                  Date last evaluated by this submitter
Submitter1_ReviewStatus              Review status of this submission
Submitter1_Sub_Condition             Submitted name of condition
Submitter1_Description  				 Description of the interpretation
Submitter2                           Name of this submitter
Submitter2_SCV                       Accession assigned to this submission 
Submitter2_ClinSig				       Clinical significance asserted by this submitter
Submitter2_LastEval                  Date last evaluated by this submitter
Submitter2_ReviewStatus              Review status of this submission
Submitter2_Sub_Condition             Submitted name of condition
Submitter2_Description               Description of the interpretation
Rank_diff                            Rank value assigned to the differences in interpretation:
                                        -1: one of the interpretations is not in the set of Pathogenic, Likely pathogenic, Uncertain significance, Likely benign, Benign
                                         0: difference in phenotype only
                                         1-4, difference when both interpretations are in the set of Pathogenic, Likely pathogenic, Uncertain significance, 
                                              Likely benign, Benign, where 4 is most divergent
Conflict_Reported                    yes or no.  Useful to supplement the Rank_diff column when Rank_diff = 1 but a conflict is still reported.
Variant_Type                         the type of variant being described
Submitter1_Method                    the collection method(s) reported by this submitter
Submitter2_Method                    the collection method(s) reported by this submitter
--------------------------------------------------------------------------------
6. hgvs4variation.txt.gz
--------------------------------------------------------------------------------
Updated weekly
Not archived

A compressed report of HGVS expressions ClinVar reports per VariationID and AlleleID. These are broadly categorized by type, based on the reference sequence (coding, genomic, non-coding, protein, RNA) and on the complexity of the submission represented by the VariationID (CompoundHeterozygote, Distinct chromosomes, Haplotype, Phase unknown). 

The header of the file explains the columns, which include the VariationID, the AlleleID, the type and the HGVS expression.  The NCBI GeneID and GeneSymbol are included for ready filtering of lines in the file by gene. The assembly is provided for the HGVS expressions based on chromosome sequences; otherwise the assembly is reported as 'na'.
HINT:  Please note that for human, the accession of the RefSeq representing each chromosome indicates the chromosome being represented. In other words, NC_000001 is for chromosome 1, NC_000002 is for chromosome 2, ... NC_000023 is for X, and NC_000024 is for Y.
In the December release, 3 columns were added to support those wishing to identify which HGVS expressions are used for naming, which HGVS expressions were provided explicitly by a submitter, and which are based on RefSeqs that are reference standards on RefSeqGenes.
--------------------------------------------------------------------------------
7. variation_allele.txt
--------------------------------------------------------------------------------
Updated weekly
Not archived

Mapping of ClinVar's VariationID (used to build the URL on the web site) and the AlleleIDs assigned to each simple variant.

1. VariationID:            the identifier assigned by ClinVar and used to build the URL, namely https://ncbi.nlm.nih.gov/clinvar/VariationID
2. Type:                   Types of VariationID include Variant (simple variant), Haplotype, CompoundHeterozygote, Complex, Phase unknown, Distinct chromosomes
3. AlleleID:               the integer identifier assigned by ClinVar to each simple allele
4. Interpreted:            _yes_ indicates an interpretation was submitted about the VariationID specifically,
                           _no_ indicates that information about the VariationID was submitted as a component of a different record.

--------------------------------------------------------------------------------
8. submission_summary.txt
--------------------------------------------------------------------------------
Generated weekly
Archived monthly (first Thurday of each month)

   Overview of interpretation, phenotypes, observations, and methods reported in each current submission 

1.  VariationID:              the identifier assigned by ClinVar and used to build the URL, namely https://ncbi.nlm.nih.gov/clinvar/VariationID
2.  ClinicalSignificance:     interpretation of the variation-condition relationship
3.  DateLastEvaluated:        the last date the variation-condition relationship was evaluated by this submitter
4.  Description:              an optional free text description of the basis of the interpretation
5.  SubmittedPhenotypeInfo:   the name(s) or identifier(s)  submitted for the condition that was interpreted relative to the variant
6.  ReportedPhenotypeInfo:    the MedGen identifier/name combinations ClinVar uses to report the condition that was interpreted. 'na' means there is no public identifier in MedGen for the condition.
7.  ReviewStatus:             the level of review for this submission, namely http://www.ncbi.nlm.nih.gov/clinvar/docs/variation_report/#review_status
8.  CollectionMethod:         the method by which the submitter obtained the information provided
9.  OriginCounts:             the reported origin and the number of observations for each origin
10. Submitter:                the submitter of this record
11. SCV:                      the accession and current version assigned by ClinVar to the submitted interpretation of the variation-condition relationship
12. SubmittedGeneSymbol:      the symbol provided by the submitter for the gene affected by the variant. May be null.

--------------------------------------------------------------------------------
9. allele_gene.txt
--------------------------------------------------------------------------------
Updated weekly
Not archived

Reports per ClinVar's AlleleID, the genes that are related to that gene and how they are related. The values for category are:

asserted, but not computed:          Submitted as related to a gene, but not within the location of that gene on the genome
genes overlapped by variant          The gene and variant overlap
near gene, downstream                Outside the location of the gene on the genome, within 5 kb
near gene, upstream                  Outside the location of the gene on the genome, within 5 kb
within multiple genes by overlap     The variant is within genes that overlap on the genome. Includes introns.
within single gene	 	 				 The variant is in only one gene.  Includes introns.

AlleleID:                            the integer identifier assigned by ClinVar to each simple allele
GeneID:                              integer, GeneID in NCBI's Gene database  
Symbol                               character, Symbol preferred in NCBI's Gene database. Is the symbol from HGNC when available
Name                                 character, full name of the gene
GenesPerAlleleID                     integer, number of genes related to the allele
Category        							 character, type of allele-gene relationship
Source                               character, was the relationship submitted or calculated?

10. organization_summary.txt
updated weekly
not archived

organization                         the name of the lab and the institution of which it is part	
organization ID                      the id used in ClinVar and GTR; often reported as OrgID;
                                     append to https://www.ncbi.nlm.nih.gov/clinvar/submitters to review more details
institution type                     type of organization
street address                       street address
country                              country
number of ClinVar submissions        number of submission to ClinVar
date last submitted                  last date on a public submission from this organization
maximum review status                the 'most stars' valid for any submission from this organization
collection methods                   comma-delimited list of methods used to determine information for the submission
novel and updates                    values are novel, novel and updates.  The latter indicates the submitter has provided updates.
clinical significance categories submitted
                                     list of types of interpretations from this organization
number of submissions from clinical testing
                                     number of submissions for the list of categories in 'collection methods'
number of submissions from research
                                     number of submissions for the list of categories in 'collection methods'
number of submissions from literature only
                                     number of submissions for the list of categories in 'collection methods'
number of submissions from curation
                                     number of submissions for the list of categories in 'collection methods'
number of submissions from phenotyping
                                     number of submissions for the list of categories in 'collection methods'
												 

11. special_requests subdirectory
This path contains files that were requested by more than one user. The reports are not kept current, but if the contents are considered useful, they may be converted to production level and updated regularly.  For more details, see the README.txt file in that path.

================================================================================
VALIDATING FILE DOWNLOADS
================================================================================
We are providing md5 checksum files to validate file downloads to ensure your ftp transfer is complete. If you are unfamiliar with md5 it is a string of letters and numbers that act as a fingerprint for a file. When you download the file generate an md5 hash and compare to the value in our md5 checksum file to ensure your download has the entire file. We are currently providing md5 files for some of the tab_delmited files and the ClinVarFullRelease XML file. After you download a file, use a utility to create a checksum value and compare it to the one we provide. In the linux enviroment, the utility is likely md5sum. There is freeware available for tools in other environments. 

================================================================================
RELATED SITES
================================================================================
--------------------------------------------------
Gene-disease relationships
--------------------------------------------------
ftp://ftp.ncbi.nlm.nih.gov/gene/DATA/mim2gene_medgen

This file, completely documented in ftp://ftp.ncbi.nih.gov/gene/README, maintains information about gene-disease relationships inferred from gene-variation and variation-disease relationships or reported by OMIM.

--------------------------------------------------
standard terms
--------------------------------------------------
ftp://ftp.ncbi.nlm.nih.gov/pub/GTR/standard_terms/

This directory contains terms used by ClinVar and GTR in specified categories.  Cross-references to the term in other databases may also be provided.

================================================================================
ClinGen 
================================================================================
ftp://ftp.ncbi.nlm.nih.gov/pub/clinvar/ClinGen/

ClinVar is an active participant in the ClinGen project. (http://clinicalgenome.org/)
As part of that collaboration, ClinVar requests that submitters wishing to be identified as an expert panel provide documentation about their their methods of determining Clinical Significance. The form to be completed is provided in this directory.


================================================================================
MISCELLANEOUS
================================================================================
--------------------------------------------------
2013.1-hgmd-public.tsv
--------------------------------------------------
  This file was removed June 19, 2013 on request from HGMD.

vcf
  This path was removed to make explicit the difference between VCF files on GRCh37 and GRCh38.
It was replaced with vcf_GRCh37 and vcf_GRCh38


======================================================================
Partial revision history
======================================================================
December 3, 2014:    Added the file /tab_delimited/summary_of_conflicting_data.txt
January 12, 2015:    Documented the changes to gene_specific_summary.txt and explained the new directory structure
                     for vcf files.
September 10, 2015:  Documented providing md5 files.
April 7, 2016:       Discontinued generating summary_of_conflicting_data.txt
                     The historical reports are maintained in the archive directory (ftp://ftp.ncbi.nlm.nih.gov/pub/clinvar/tab_delimited/archive/)


It was defined as below:
Generated monthly, the first Thursday of the month
A tab-delimited report based on variants in ClinVar, for which information has been provided by more than one submitter, and for which there is inconsistency in reporting phenotype or interpretation.  The file includes some basic information about the variant, and then describes what each submitter said for which a discrepancy was noted.  Although generated primarily for the submitter, or groups reviewing evidence supporting variant assessments, this file may be of interest to others as well. Please note *all* data conflicts are reported.  To focus on a subset of interest, we suggest you use either the Rank_diff or Conflict_Reported columns. For example, look for rows in the report where  Rank_diff >=2 for Conflict_Reported is 'yes'.

Gene_Symbol                          If in a gene, its symbol
NCBI_Variation_ID                    The identifier ClinVar uses to anchor its default display. (in the XML,  //MeasureSet/@ID)
NCBI_AlleleID                        The identifier ClinVar uses to define an allele. (in the XML,  //Measure/@ID)
HGVS                                 The default HGVS expression for this variant
Submitter1                           Name of this submitter
Submitter1_ID   							 Identifier for this variant provided by this submitter, or constructed by NCBI for the submitter
Submitter1_SCV  							 Accession assigned to this submission
Submitter1_Definition   				 Variant as defined by this submitter  (currently missing for those not submitted by HGVS)
Submitter1_ClinSig                   Clinical significance asserted by this submitter
Submitter1_LastEval                  Date last evaluated by this submitter
Submitter1_AssertionMethod           Assertion method
Submitter1_Sub_Condition             Submitted name of condition
Submitter1_Calc_Condition            Name ClinVar uses for this condition
Submitter1_Description  				 Description of the interpretation
Submitter2                           Name of this submitter
Submitter2_ID                        Identifier for this variant provided by this submitter, or constructed by NCBI for the submitter
Submitter2_SCV                       Accession assigned to this submission 
Submitter2_Definition                Variant as defined by this submitter  (currently missing for those not submitted by HGVS)
Submitter2_ClinSig				       Clinical significance asserted by this submitter
Submitter2_LastEval                  Date last evaluated by this submitter
Submitter2_AssertionMethod           Assertion method
Submitter2_Sub_Condition             Submitted name of condition
Submitter2_Calc_Condition            Name ClinVar uses for this condition
Submitter2_Description               Description of the interpretation
Rank_diff                            Rank value assigned to the differences in interpretation:
                                        -1: one of the interpretations is not in the set of Pathogenic, Likely pathogenic, Uncertain significance, Likely benign, Benign
                                         0: difference in phenotype only
                                         1-4, difference when both interpretations are in the set of Pathogenic, Likely pathogenic, Uncertain significance, 
                                              Likely benign, Benign, where 4 is most divergent
Conflict_Reported                    yes or no.  Useful to supplement the Rank_diff column when Rank_diff = 1 but a conflict is still reported.
Variant_Type                         the type of variant being described

August 4, 2016:        Added documentation for several new files in the tab-delimited subdirectory
                       * hgvs4variation.txt.gz
                       * submission_summary.txt
                       * variation_allele.txt
                       Added explanation of numbering systems for locations in VCF vs. all other reports
                                      
September 24, 2016:	  Added documentation for the allele_gene files in the tab-delimited directory
November   9, 2016:    Corrected documentation for variant_summary that should have been included in the October update.
December   9, 2016:    Added SubmittedGeneSymbol to the description of submission_summary.txt, and updated the description of hgvs4variation.txt.gz
January   19, 2017     Corrected inconsistent definition of NumberSubmitters in variant_summary.txt
November  30, 2017     Modification to summary_of_conflicting_interpretations.txt
February  20, 2019     Added documentation of organization_summary.txt
November  12, 2019     Noted change in reporting referenceAllele and alternateAllele in variant_summary.txt; added more information about releases.