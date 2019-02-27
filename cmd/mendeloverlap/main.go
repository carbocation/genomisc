// MendelOverlap performs permutation to quantify surprise at the number of GWAS
// loci that overlap with Mendelian genes for your trait of interest. GRCh37
// only.
package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/glycerine/golang-fisher-exact"
	"github.com/gobuffalo/packr"
)

const (
	GeneStableID int = iota
	TranscriptStableID
	ProteinStableID
	Chromosome
	GeneStartOneBased
	GeneEndOneBased
	Strand
	TranscriptStartOneBased
	TranscriptEndOneBased
	TranscriptLengthIncludingUTRAndCDS
	GeneName
)

type Gene struct {
	Symbol                  string
	Chromosome              string
	EarliestTranscriptStart int
	LatestTranscriptEnd     int
}

type Results struct {
	Permutations []Permutation
}

func (e Results) FisherExactTest(nAllGenes, nAllMendelianGenes int) float64 {
	// FisherExactTest computes Fisher's Exact Test for
	//  contigency tables. Nomenclature:
	//
	//    n11  n12  | n1_
	//    n21  n22  | n2_
	//   -----------+----
	//    n_1  n_2  | n
	//
	// Returned values:
	//
	//  probOfCurrentTable = probability of the current table
	//  leftp = the left sided alternative's p-value  (h0: odds-ratio is less than 1)
	//  rightp = the right sided alternative's p-value (h0: odds-ratio is greater than 1)
	//  twop = the two-sided p-value for the h0: odds-ratio is different from 1
	//

	// Mendelian genes that we did not find in this permutation
	notfoundMendelianGenes := nAllMendelianGenes - e.MendelianGeneCount()

	residualGenes := nAllGenes - e.NonMendelianGeneCount() - e.MendelianGeneCount()

	_, _, _, twop := fet.FisherExactTest(e.MendelianGeneCount(), notfoundMendelianGenes, e.NonMendelianGeneCount(), residualGenes)

	return twop
}

func (e Results) NonMendelianGeneCount() int {
	n := 0
	for _, v := range e.Permutations {
		n += v.NonMendelianGenesNearLoci
	}

	return n
}

func (e Results) MendelianGeneCount() int {
	n := 0
	for _, v := range e.Permutations {
		n += v.MendelianGenesNearLoci
	}

	return n
}

type Permutation struct {
	NonMendelianGenesNearLoci int
	MendelianGenesNearLoci    int
}

func main() {
	var (
		mendelianGeneFile string
		SNPsnapFile       string
		radius            float64
	)

	flag.StringVar(&mendelianGeneFile, "mendel", "", "Filename containing one gene symbol per line representing your Mendelian disease genes.")
	flag.StringVar(&SNPsnapFile, "snpsnap", "", "Filename containing SNPsnap output.")
	flag.Float64Var(&radius, "radius", 100, "Radius, in kilobases, to define whether part of a transcript is 'within' a given locus.")
	flag.Parse()

	if mendelianGeneFile == "" || SNPsnapFile == "" || radius < 0 {
		flag.PrintDefaults()
		return
	}

	mendelianGeneList, err := ReadMendelianGeneFile(mendelianGeneFile)
	if err != nil {
		log.Fatalln(err)
	}

	if len(mendelianGeneList) < 1 {
		log.Println("No genes were parsed from your Mendelian gene file")
	}

	mendelianTranscripts, err := SimplifyTranscripts(mendelianGeneList)
	if err != nil {
		log.Fatalln(err)
	}

	for _, v := range mendelianTranscripts {
		fmt.Print(v, " ")
	}
	fmt.Println()

	_ = mendelianTranscripts
}

func ReadMendelianGeneFile(fileName string) (map[string]struct{}, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cr := csv.NewReader(f)
	cr.Comment = '#'
	lines, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}

	genes := make(map[string]struct{})
	for _, cols := range lines {
		if len(cols) < 1 {
			continue
		}
		genes[cols[0]] = struct{}{}
	}

	return genes, nil
}

// Fetches all transcripts for a named symbol and simplifies them to the largest
// span of transcript start - transcript end, so there is just one entry per
// gene.
func SimplifyTranscripts(geneNames map[string]struct{}) (map[string]Gene, error) {
	lookups := packr.NewBox("./lookups")

	file := lookups.Bytes("ensembl.grch37.p13.genes")
	buf := bytes.NewBuffer(file)
	cr := csv.NewReader(buf)
	cr.Comma = '\t'
	cr.Comment = '#'

	results := make([]Gene, 0)

	var i int64
	for {
		rec, err := cr.Read()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		i++
		if i == 1 {
			continue
		}

		if _, exists := geneNames[rec[GeneName]]; !exists {
			continue
		}

		start, err := strconv.Atoi(rec[TranscriptStartOneBased])
		if err != nil {
			return nil, err
		}

		end, err := strconv.Atoi(rec[TranscriptEndOneBased])
		if err != nil {
			return nil, err
		}

		results = append(results, Gene{Symbol: rec[GeneName], Chromosome: rec[Chromosome], EarliestTranscriptStart: start, LatestTranscriptEnd: end})
	}

	keepers := make(map[string]Gene)
	for _, transcript := range results {
		keeper, exists := keepers[transcript.Symbol]
		if !exists {
			keepers[transcript.Symbol] = transcript
			continue
		}

		if transcript.EarliestTranscriptStart < keeper.EarliestTranscriptStart {
			temp := keepers[transcript.Symbol]
			temp.EarliestTranscriptStart = transcript.EarliestTranscriptStart
			keepers[transcript.Symbol] = temp
		}

		if transcript.LatestTranscriptEnd < keeper.LatestTranscriptEnd {
			temp := keepers[transcript.Symbol]
			temp.LatestTranscriptEnd = transcript.LatestTranscriptEnd
			keepers[transcript.Symbol] = temp
		}
	}

	if len(keepers) < len(geneNames) {
		missing := make([]string, 0)
		for key := range geneNames {
			if _, exists := keepers[key]; !exists {
				missing = append(missing, key)
			}
		}
		return keepers, fmt.Errorf("Your Mendelian set contained %d genes, but we could only map %d of them. Missing: %v", len(geneNames), len(keepers), missing)
	}

	return keepers, nil
}
