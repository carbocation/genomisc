package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/carbocation/bgen"
	"github.com/carbocation/genomisc/applyprsgcp/prsworker"
)

func BGENPreprocessor(bgenPath, chromosome string) ([]bgen.VariantIndex, error) {

	bgi, b, err := prsworker.OpenBGIAndBGEN(bgenPath, bgenPath+".bgi")
	if err != nil {
		log.Fatalln(err)
	}
	defer bgi.Close()
	defer b.Close()

	// Fetch variants within our range
	firstPos, lastPos := FirstLastPos(currentVariantScoreLookup)

	// Note: BGENIX stores chromosome 1 as "01", etc.(!)
	sitesTemp := make([]bgen.VariantIndex, 0, len(currentVariantScoreLookup))
	if err := bgi.DB.Select(&sitesTemp, "SELECT * FROM Variant WHERE chromosome=? AND position>=? AND position<=? ORDER BY chromosome ASC, position ASC", fmt.Sprintf("%02s", chromosome), firstPos, lastPos); err != nil {
		return nil, err
	}

	// For each variant, we now need to do 2 things: First, eliminate leading
	// zeroes (BGENIX stores chromosomes as strings with leading zeroes!). Then,
	// eliminate any variant that cannot contribute to the score because it is
	// not in the PRS.
	sites := make([]bgen.VariantIndex, 0, len(sitesTemp))
	for _, site := range sitesTemp {

		// Eliminate leading zero
		chrInt, err := strconv.Atoi(site.Chromosome)
		if err != nil {
			return nil, err
		}
		site.Chromosome = strconv.Itoa(chrInt)

		// Don't keep the variant if it's not part of the risk score
		if LookupPRS(site.Chromosome, site.Position) == nil {
			continue
		}
		sites = append(sites, site)
	}

	log.Printf("There are %d variants in the range of %s:%d-%d. Of these, %d map to the PRS lookup file)\n", len(sitesTemp), chromosome, firstPos, lastPos, len(sites))

	if len(sites) < 1 && len(sitesTemp) > 0 {
		// The BGEN has sites here, but the PRS doesn't cover any variants in
		// this region. There is nothing to process. So, why did you create a
		// job for this?
		return nil, fmt.Errorf("There are no PRS sites to process in range %s:%d-%d", chromosome, firstPos, lastPos)
	} else if len(sites) < 1 {
		// The BGEN has no sites here, which is probably an error.
		return nil, fmt.Errorf("There are no BGEN sites in range %s:%d-%d", chromosome, firstPos, lastPos)
	}

	// There are some sites
	return sites, nil

}
