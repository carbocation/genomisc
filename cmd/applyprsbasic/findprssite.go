package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/carbocation/bgen"
	"github.com/carbocation/genomisc/prsparser"
	"github.com/carbocation/pfx"
)

// FindPRSSiteInBGI reads any variant(s) in the BGI whose positions match that
// of a single variant in the PRS, so that the location in the BGEN binary can
// be extracted for genotype extraction.
func FindPRSSiteInBGI(bgi *bgen.BGIIndex, siteScore prsparser.PRS) ([]bgen.VariantIndex, error) {

	prsUsesCHR := false
	if strings.HasPrefix(siteScore.Chromosome, "chr") {
		prsUsesCHR = true
	}

	// Note: BGENIX stores chromosome 1 as "01", etc., in the UK Biobank.
	sitesTemp := make([]bgen.VariantIndex, 0, len(currentVariantScoreLookup))
	if err := bgi.DB.Select(&sitesTemp, "SELECT * FROM Variant WHERE position=?", siteScore.Position); err != nil {
		return nil, pfx.Err(err)
	}

	// Does the bgi use a chr prefix? Just test the first one.
	bgiUsesCHR := false
	for _, v := range sitesTemp {
		if strings.HasPrefix(v.Chromosome, "chr") {
			bgiUsesCHR = true
		}

		break
	}

	// For each variant, we now need to do 2 things: First, eliminate leading
	// zeroes (in the UK Biobank, BGENIX stores chromosomes as strings with
	// leading zeroes!). Then, eliminate any variant that cannot contribute to
	// the score because it is not in the PRS.
	sites := make([]bgen.VariantIndex, 0, len(sitesTemp))
	for _, site := range sitesTemp {
		// Chromosomes are so contentious
		switch {
		case bgiUsesCHR && !prsUsesCHR:
			site.Chromosome = strings.ReplaceAll(site.Chromosome, "chr", "")
		case !bgiUsesCHR && prsUsesCHR:
			site.Chromosome = "chr" + site.Chromosome
		default:
			// By default, they either both do or both don't use chr
		}

		// Eliminate leading zero (UK Biobank-specific problem)
		if strings.HasPrefix(site.Chromosome, "0") {
			chrInt, err := strconv.Atoi(site.Chromosome)
			if err != nil {
				return nil, pfx.Err(err)
			}
			site.Chromosome = strconv.Itoa(chrInt)
		}

		// Don't keep the variant if it's not part of the risk score
		if LookupPRS(site.Chromosome, site.Position) == nil {
			continue
		}
		sites = append(sites, site)
	}

	// log.Printf("There are %d variants in the range of %s:%d-%d. Of these, %d map to the PRS lookup file)\n", len(sitesTemp), siteScore.Chromosome, siteScore.Position, siteScore.Position, len(sites))

	if len(sites) < 1 && len(sitesTemp) > 0 {
		// The BGEN has sites here, but the PRS doesn't cover any variants in
		// this region. There is nothing to process. So, why did you create a
		// job for this?
		return nil, pfx.Err(fmt.Errorf("There are no PRS sites to process at site range %s:%d-%d", siteScore.Chromosome, siteScore.Position, siteScore.Position))
	} else if len(sites) < 1 {
		// The BGEN has no sites here, which is probably an error.
		return nil, pfx.Err(fmt.Errorf("There are no BGEN sites in range %s:%d-%d (file %s)", siteScore.Chromosome, siteScore.Position, siteScore.Position, bgi.Metadata.Filename))
	}

	// There are some sites
	return sites, nil

}
