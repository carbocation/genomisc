package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/carbocation/genomisc"
	"github.com/carbocation/genomisc/prsparser"
	"github.com/carbocation/pfx"
)

var currentVariantScoreLookup map[ChrPos]prsparser.PRS

type ChrPos struct {
	Chromosome string
	Position   uint32
	SNP        string
}

type PRSSitesOnChrom struct {
	Chrom    string
	PRSSites []prsparser.PRS
}

// LoadPRS is ***not*** safe for concurrent access from multiple goroutines
func LoadPRS(prsPath, layout string, alwaysIncrement bool) error {
	parser, err := prsparser.New(layout)
	if err != nil {
		return fmt.Errorf("CreatePRSParserErr: %s", err.Error())
	}

	// Open PRS file
	f, err := os.Open(prsPath)
	if err != nil {
		return pfx.Err(err)
	}
	defer f.Close()

	fd, err := genomisc.MaybeDecompressReadCloserFromFile(f)
	if err != nil {
		return pfx.Err(err)
	}
	defer fd.Close()

	reader := csv.NewReader(fd)
	reader.Comma = parser.CSVReaderSettings.Comma
	reader.Comment = parser.CSVReaderSettings.Comment
	reader.TrimLeadingSpace = parser.CSVReaderSettings.TrimLeadingSpace

	currentVariantScoreLookup = nil
	currentVariantScoreLookup = make(map[ChrPos]prsparser.PRS)
	for i := 0; ; i++ {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else if err, ok := err.(*csv.ParseError); ok && err.Err == csv.ErrFieldCount {
				// We actually permit this
				log.Printf("Recovering from parsing error which may be caused by jagged files with missing entries and proceeding with this variant. Error: %s", err.Error())
			} else {
				return pfx.Err(err)
			}
		}

		val, err := parser.ParseRow(row)
		if err != nil && i == 0 {
			// Permit a header and skip it
			continue
		} else if err != nil {
			return pfx.Err(err)
		}

		p := prsparser.PRS{
			Chromosome:   val.Chromosome,
			Position:     val.Position,
			EffectAllele: val.EffectAllele,
			Allele1:      val.Allele1,
			Allele2:      val.Allele2,
			Score:        val.Score,
			SNP:          val.SNP,
		}

		if p.EffectAllele != p.Allele1 && p.EffectAllele != p.Allele2 {
			return fmt.Errorf("Effect Allele (%v) is neither equal to Allele 1 (%v) nor Allele 2 (%v)", p.EffectAllele, p.Allele1, p.Allele2)
		}

		// Ensure that all scores will be positive. If the effect size is
		// negative, swap the effect and alt alleles and the effect sign.
		if alwaysIncrement && p.Score < 0 {
			p.Score *= -1
			if p.EffectAllele == p.Allele1 {
				p.EffectAllele = p.Allele2
			} else {
				p.EffectAllele = p.Allele1
			}
		}

		currentVariantScoreLookup[ChrPos{p.Chromosome, uint32(p.Position), p.SNP}] = p
	}

	return nil
}

func LookupPRS(chromosome string, position uint32, snp string) *prsparser.PRS {
	if prs, exists := currentVariantScoreLookup[ChrPos{chromosome, position, snp}]; exists {
		return &prs
	} else {
		// See if we have missed a leading zero
		chrInt, err := strconv.Atoi(chromosome)
		if err != nil {
			// If it cannot be parsed as an integer, then it was a sex
			// chromosome and it truly didn't match.
			return nil
		}

		// We parsed as an integer. Now recheck without the leading zero to see
		// if we can match.
		if prs, exists := currentVariantScoreLookup[ChrPos{strconv.Itoa(chrInt), position, snp}]; exists {
			return &prs
		}
	}

	return nil
}

// ChromosomalPRS creates a map containing each chromosome and the PRS variants
// on that chromosome.
func ChromosomalPRS(currentVariantScoreLookup map[ChrPos]prsparser.PRS) map[string][]prsparser.PRS {
	output := make(map[string][]prsparser.PRS)

	for _, v := range currentVariantScoreLookup {
		if _, exists := output[v.Chromosome]; !exists {
			output[v.Chromosome] = make([]prsparser.PRS, 0)
		}

		output[v.Chromosome] = append(output[v.Chromosome], v)
	}

	return output
}

// splitter parallelizes the computation. Each separate goroutine will process a
// set of sites. Each goroutine will have its own BGEN and BGI filehandles.
func splitter(chromosomalPRS map[string][]prsparser.PRS, chunkSize int) []PRSSitesOnChrom {
	out := make([]PRSSitesOnChrom, 0)

	var subChunk PRSSitesOnChrom

	// For each chromosome
	for chrom, prsSites := range chromosomalPRS {
		if subChunk.Chrom == "" {
			// We are initializing from scratch
			subChunk.Chrom = chrom
			subChunk.PRSSites = make([]prsparser.PRS, 0, chunkSize)
		}

		if subChunk.Chrom != chrom {
			// We have moved to a new chromosome
			out = append(out, subChunk)
			subChunk = PRSSitesOnChrom{
				Chrom:    chrom,
				PRSSites: make([]prsparser.PRS, 0, chunkSize),
			}
		}

		// Within this chromosome, add sites into different chunks
		for k, prsSite := range prsSites {
			subChunk.PRSSites = append(subChunk.PRSSites, prsSite)

			// But if we have iterated to our chunk size, then create a new
			// chunk
			if k > 0 && k%chunkSize == 0 {
				out = append(out, subChunk)
				subChunk = PRSSitesOnChrom{
					Chrom:    chrom,
					PRSSites: make([]prsparser.PRS, 0, chunkSize),
				}
			}
		}
	}

	// Clean up at the end
	if subChunk.Chrom != "" {
		out = append(out, subChunk)
	}

	return out
}
