package main

import (
	"fmt"
	"strings"

	"github.com/carbocation/vcfgo"
)

var sampleFields = []string{
	"RS",

	// ALLELEID: the Allele ID for the variant
	"ALLELEID",

	//  CLNSIG: a clinical significance.  E.g., "Uncertain_significance"
	"CLNSIG",

	// CLNVC: the type of variation (e.g., single_nucleotide_variant)
	"CLNVC",

	// CLNVI: reports identifiers for the variant in other databases, e.g. OMIM
	// Allelic variant IDs.
	"CLNVI",

	// CLNREVSTAT: Not defined in the README. Appears to describe the submitter
	// and the degree of clinical review of the variant.
	"CLNREVSTAT",

	//  CLNDISDB: associated disease. Links out to multiple databases.
	//  Comma-delimited list of key:value entries where the key is the database
	//  name and the value is the disease ID in that database.
	"CLNDISDB",

	//  CLNDN: associated disease. May be a list, in which case it is pipe
	//  delimited.
	"CLNDN",

	// GENEINFO: Not defined in the README. Appears to show the HUGO name of a
	// gene followed by a gene ID (e.g., FBN1:2200)
	"GENEINFO",

	// MC: report the predicted molecular consequence of the variant. It is
	// reported as pairs of the Sequence Ontology (SO) identifier and the
	// molecular consequence term joined by a vertical bar. Multiple values are
	// separated by a comma.
	"MC",

	// ORIGIN: reports an integer representing the allele origins that have been
	// observed for the variant and reported to ClinVar. One or more of the
	// values may be added: 0 - unknown; 1 - germline; 2 - somatic; 4 -
	// inherited; 8 - paternal; 16 - maternal; 32 - de-novo; 64 - biparental;
	// 128 - uniparental; 256 - not-tested; 512 - tested-inconclusive;
	// 1073741824 - other. This tag replaces CLNORIGIN in the old format.
	"ORIGIN",

	// Interpretations may be made on a single variant or a set of variants,
	// such as a haplotype. Variants that have only been interpreted as part of
	// a set of variants (i.e. no direct interpretation for the variant itself)
	// are considered "included" variants. The VCF files include both variants
	// with a direct interpretation and included variants. Included variants do
	// not have an associated disease (CLNDN, CLNDISDB) or a clinical
	// significance (CLNSIG). Instead there are three tags are specific to the
	// included variants - CLNDNINCL, CLNDISDBINCL, and CLNSIGINCL (see below).

	// CLNDNINCL: reports ClinVar's preferred disease name for an interpretation
	// for a haplotype or genotype that includes this variant.
	"CLNDNINCL",

	// CLNSIGINCL: reports the clinical significance of a haplotype or genotype
	// that includes this variant. It is reported as pairs of Variation ID for
	// the haplotype or genotype and the corresponding clininical significance.
	"CLNSIGINCL",

	// CLNDISDBINCL: reports the database name and identifier for the disease
	// name for an interpretation for a haplotype or genotype that includes this
	// variant. Multiples are separated by a pipe.
	"CLNDISDBINCL",
}

type Work struct {
	Chrom        string
	Pos          uint64
	Ref          string
	Alt          string
	SNP          string
	SampleFields []string
}

func (w Work) String() string {
	formatting := "%s\t%s\t%d\t%s\t%s\t%s\t%s" //+ strings.Repeat("%s\t", len(w.SampleFields)-1)

	siteID := fmt.Sprintf("%s:%d:%s:%s", w.Chrom, w.Pos, w.Ref, w.Alt)

	return fmt.Sprintf(formatting, siteID, w.Chrom, w.Pos, w.Ref, w.Alt, w.SNP, strings.Join(w.SampleFields, "\t"))
}

func header() string {
	head := make([]string, 0)
	head = append(head, []string{"siteid", "CHR", "BP", "REF", "ALT", "SNP"}...)
	head = append(head, sampleFields...)

	return fmt.Sprintf("%s", strings.Join(head, "\t"))
}

func worker(variant *vcfgo.Variant, alleleID int) Work {
	w := Work{
		Chrom: variant.Chrom(),
		Pos:   variant.Pos,
		Ref:   variant.Ref(),
		Alt:   variant.Alt()[alleleID],
		SNP:   variant.Id(),
	}

	// Add the sample fields in order, if requested
	for _, requestedField := range sampleFields {
		if requestedField == "FILTER" {
			// Field is FILTER?
			w.SampleFields = append(w.SampleFields, variant.Filter)
		} else if _, err := variant.Info().Get(strings.TrimPrefix(requestedField, "INFO_")); err == nil {
			// Field is at the variant-level (within the INFO field)?
			// Note that you can force a look at the INFO field by prefixing with INFO_

			// TODO: This does not correctly handle multiallelics (which yield an array, from which you should pick the alleleID'th allele)

			// Need to ensure that the INFO_ prefix is removed, if present
			reqField := strings.TrimPrefix(requestedField, "INFO_")

			// Prepend rs to the rsid.
			stringVal := string(vcfgo.NewInfoByte(variant.Info().Bytes(), nil).SGet(reqField))
			if reqField == "RS" {
				stringVal = "rs" + stringVal
			}

			w.SampleFields = append(w.SampleFields, stringVal) //fmt.Sprintf("%v", field)) //string(vcfgo.NewInfoByte(variant.Info().Bytes(), nil).SGet(reqField)))
		} else {
			w.SampleFields = append(w.SampleFields, "")
		}
	}

	return w

}
