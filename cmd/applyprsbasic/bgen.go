package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/carbocation/bgen"
	"github.com/carbocation/genomisc/applyprsgcp"
	"github.com/carbocation/genomisc/applyprsgcp/prsworker"
	"github.com/carbocation/genomisc/prsparser"
	"github.com/carbocation/pfx"
)

func init() {
	log.Println(bgen.WhichSQLiteDriver())
}

func scoreBGEN(chromosome string, chromosomalSites []prsparser.PRS, bgenTemplatePath, bgiTemplatePath string) ([]Sample, error) {
	// Place to accumulate scores
	var score []Sample
	var err error

	// Load the BGEN Index for this chromosome
	bgenPath := fmt.Sprintf(bgenTemplatePath, chromosome)
	if !strings.Contains(bgenTemplatePath, "%s") {
		// Permit explicit paths (e.g., when all data is in one BGEN)
		bgenPath = bgenTemplatePath
	}

	bgiPath := fmt.Sprintf(bgiTemplatePath, chromosome) //bgenPath + ".bgi"
	if !strings.Contains(bgiTemplatePath, "%s") {
		// Permit explicit paths (e.g., when all data is in one BGEN)
		bgiPath = bgiTemplatePath
	}

	// Open the BGI
	var bgi *bgen.BGIIndex
	var b *bgen.BGEN

	// Repeatedly reading SQLite files over-the-wire is slow. So localize
	// them.
	if strings.HasPrefix(bgiPath, "gs://") {
		bgiFilePath, newDownload, err := applyprsgcp.ImportBGIFromGoogleStorageLocked(bgiPath, client)
		if err != nil {
			log.Fatalln(err)
		}

		if newDownload {
			log.Printf("Copied BGI file from %s to %s\n", bgiPath, bgiFilePath)
		}

		bgiPath = bgiFilePath
	}

	bgi, b, err = prsworker.OpenBGIAndBGEN(bgenPath, bgiPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer bgi.Close()
	defer b.Close()

	// if len(score) == 0 {
	// 	// If we haven't initialized the score slice yet, do so now by
	// 	// reading a random variant and creating a slice that has one sample
	// 	// per sample found in that variant.
	// 	vr := b.NewVariantReader()
	// 	randomVariant := vr.Read()
	// 	score = make([]Sample, len(randomVariant.SampleProbabilities))
	// }

	// Iterate over each chromosomal position on this current chromosome
	for _, oneSite := range chromosomalSites {

		atomic.AddUint64(&nSitesProcessed, 1)

		var sites []bgen.VariantIndex

		// Find the site in the BGEN index file. Here we also handle a
		// specific filesystem error (input/output error) and retry in that
		// case. Namely, exiting early is superior to producing incorrect
		// output in the face of i/o errors. But, since we expect other
		// "error" messages, we permit those to just be printed and ignored.
		for loadAttempts, maxLoadAttempts := 1, 10; loadAttempts <= maxLoadAttempts; loadAttempts++ {
			sites, err = FindPRSSiteInBGI(bgi, oneSite)
			if err != nil && loadAttempts == maxLoadAttempts {
				// Ongoing failure at maxLoadAttempts is a terminal error
				log.Fatalln(err)
			} else if err != nil && strings.Contains(err.Error(), "input/output error") {
				// If we had an error, often due to an unreliable underlying
				// filesystem, wait for a substantial amount of time before
				// retrying.
				log.Println("FindPRSSiteInBGI: Sleeping 5s to recover from", err.Error(), "attempt", loadAttempts)
				time.Sleep(5 * time.Second)

				// We also seemingly need to reopen the handles.
				b.Close()
				bgi.Close()
				bgi, b, err = prsworker.OpenBGIAndBGEN(bgenPath, bgiPath)
				if err != nil {
					log.Fatalln(err)
				}

				continue
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "Skipping %v\n:", oneSite)
				log.Println(err)
			}

			// No errors? Don't retry.
			break
		}

		// Multiple variants may be present at each chromosomal position
		// (e.g., multiallelic sites, long deletions/insertions, etc) so
		// test each of them for being an exact match.
	SitesLoop:
		for _, site := range sites {

			// Process a variant from the BGEN file. Again, we also handle a
			// specific filesystem error (input/output error) and retry in
			// that case.
			for loadAttempts, maxLoadAttempts := 1, 10; loadAttempts <= maxLoadAttempts; loadAttempts++ {

				err = ProcessOneVariant(b, site, &oneSite, &score)
				if err != nil && loadAttempts == maxLoadAttempts {
					// Ongoing failure at maxLoadAttempts is a terminal error
					log.Fatalln(err)
				} else if err != nil && strings.Contains(err.Error(), "input/output error") {
					// If we had an error, often due to an unreliable underlying
					// filesystem, wait for a substantial amount of time before retrying.
					log.Println("ProcessOneVariant: Sleeping 5s to recover from", err.Error(), "attempt", loadAttempts)
					time.Sleep(5 * time.Second)

					// We also seemingly need to reopen the handles.
					b.Close()
					bgi.Close()
					bgi, b, err = prsworker.OpenBGIAndBGEN(bgenPath, bgiPath)
					if err != nil {
						log.Fatalln(err)
					}

					continue
				} else if err != nil {
					// Non i/o errors usually indicate that this is the wrong allele of a multiallelic. Print and move on.
					log.Println(err)
					continue SitesLoop
				}

				// No errors? Don't retry.
				break
			}
		}
	}

	// Clean up our file handles
	bgi.Close()
	b.Close()

	return score, nil
}

func ProcessOneVariant(b *bgen.BGEN, vi bgen.VariantIndex, prs *prsparser.PRS, scores *[]Sample) error {

	nonNilErr := ErrorInfo{Message: "", Chromosome: vi.Chromosome, Position: vi.Position}

	if prs == nil || b == nil {
		nonNilErr.Message = "prs was nil"
		return nonNilErr
	}
	if b == nil {
		nonNilErr.Message = "b was nil"
		return nonNilErr
	}

	if vi.NAlleles > 2 {
		nonNilErr.Message = fmt.Sprintf("%s:%d is multiallelic, skipping", vi.Chromosome, vi.Position)
		return nonNilErr
	}

	// Check whether there is allelic match (we assume same strand) between PRS
	// and the genetic data. Do this in case-insensitive fashion.
	if (strings.EqualFold(string(prs.Allele1), string(vi.Allele1)) && strings.EqualFold(string(prs.Allele2), string(vi.Allele2))) ||
		(strings.EqualFold(string(prs.Allele1), string(vi.Allele2)) && strings.EqualFold(string(prs.Allele2), string(vi.Allele1))) {
		// Must be the case
	} else {
		nonNilErr.Message = fmt.Sprintf("At %s:%d, PRS Alleles were %s,%s but variant alleles were %s,%s", vi.Chromosome, vi.Position, prs.Allele1, prs.Allele2, vi.Allele1, vi.Allele2)
		return nonNilErr
	}

	vr := b.NewVariantReader()

	variant := vr.ReadAt(int64(vi.FileStartPosition))
	if err := vr.Error(); err != nil {
		nonNilErr.Message = err.Error()
		return nonNilErr
	}

	// If it turns out that we are initializing the slice...
	if scores == nil || len(*scores) < 1 {
		results := make([]Sample, len(variant.SampleProbabilities))
		*scores = results
	}

	results := *scores
	for i := 0; i < len(results); i++ {
		results[i].SumScore += ComputeScore(variant.SampleProbabilities[i], variant, prs)
		results[i].NIncremented++
	}

	return nil
}

func ComputeScore(sampleProb bgen.SampleProbability, v *bgen.Variant, prs *prsparser.PRS) float64 {
	if sampleProb.Ploidy != 2 || len(sampleProb.Probabilities) != 3 {
		return 0.0
	}

	if strings.EqualFold(string(prs.EffectAllele), string(v.Alleles[0])) {
		return prs.Score * (2.0*sampleProb.Probabilities[0] + sampleProb.Probabilities[1])
	}

	if strings.EqualFold(string(prs.EffectAllele), string(v.Alleles[1])) {
		return prs.Score * (sampleProb.Probabilities[1] + 2.0*sampleProb.Probabilities[2])
	}

	return 0.0
}

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
	if siteScore.UseSNP() {
		// Our PRS does not contain position information. Explicitly fetch every
		// field except "position", since we need to *not* know that in order to
		// match based on how UseSNP() is defined.
		if err := bgi.DB.Select(&sitesTemp, "SELECT chromosome, rsid, number_of_alleles, allele1, allele2, file_start_position, size_in_bytes FROM Variant WHERE rsid=?", siteScore.SNP); err != nil {
			return nil, pfx.Err(err)
		}
	} else if siteScore.AllowSNP() {
		// Our PRS contains both SNP and position information, so we can fetch
		// all fields. Will choose based on position, but this is arbitrary.
		if err := bgi.DB.Select(&sitesTemp, "SELECT * FROM Variant WHERE position=?", siteScore.Position); err != nil {
			return nil, pfx.Err(err)
		}
	} else {
		// Our PRS does not contain SNP information, just position information.
		// Explicitly exclude the "rsid" field, since we need to *not* know that
		// in order to match based on how UseSNP() is defined.
		if err := bgi.DB.Select(&sitesTemp, "SELECT chromosome, position, number_of_alleles, allele1, allele2, file_start_position, size_in_bytes FROM Variant WHERE position=?", siteScore.Position); err != nil {
			return nil, pfx.Err(err)
		}
	}

	// Does the bgi use a chr prefix? Just test the first one.
	bgiUsesCHR := false
	for _, v := range sitesTemp {
		if strings.HasPrefix(v.Chromosome, "chr") {
			bgiUsesCHR = true
		}

		// Just check the first entry then terminate the loop
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
		if LookupPRS(site.Chromosome, site.Position, site.RSID) == nil {
			continue
		}
		sites = append(sites, site)
	}

	// log.Printf("There are %d variants in the range of %s:%d-%d. Of these, %d map to the PRS lookup file)\n", len(sitesTemp), siteScore.Chromosome, siteScore.Position, siteScore.Position, len(sites))

	if len(sites) < 1 && len(sitesTemp) > 0 {
		// The BGEN has sites here, but the PRS doesn't cover any variants in
		// this region. There is nothing to process. So, why did you create a
		// job for this?
		return nil, pfx.Err(fmt.Errorf("There are no PRS sites to process at site range %s:%d-%d (SNP %s)", siteScore.Chromosome, siteScore.Position, siteScore.Position, siteScore.SNP))
	} else if len(sites) < 1 {
		// The BGEN has no sites here, which is probably an error.
		return nil, pfx.Err(fmt.Errorf("There are no BGEN sites in range %s:%d-%d (SNP %s; file %s)", siteScore.Chromosome, siteScore.Position, siteScore.Position, siteScore.SNP, bgi.Metadata.Filename))
	}

	// There are some sites
	return sites, nil

}
