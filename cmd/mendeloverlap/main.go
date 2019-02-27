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
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/carbocation/pfx"
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
	Permutations   []Permutation
	MendelianGenes map[string]Gene
	Radius         float64
}

type Hist struct {
	Value    int
	Count    int
	Original bool
}

func (e Results) Summarize() {
	fmt.Println("Examining", len(e.Permutations), "Permutations")

	origValue := e.Permutations[0].MendelianGenesNearLoci(e.MendelianGenes, e.Radius)

	mendelianCounts := make(map[int]Hist)
	for _, permutation := range e.Permutations {
		val := permutation.MendelianGenesNearLoci(e.MendelianGenes, e.Radius)
		hist := mendelianCounts[val]
		hist.Value = val
		hist.Count++
		mendelianCounts[val] = hist
	}

	histslice := make([]Hist, 0, len(mendelianCounts))
	for _, v := range mendelianCounts {
		if v.Value == origValue {
			v.Original = true
		}
		histslice = append(histslice, v)
	}
	sort.Slice(histslice, func(i, j int) bool {
		if histslice[i].Value < histslice[j].Value {
			return true
		}

		return false
	})

	fmt.Printf("N_Overlapping_Loci\tPermutations\tOriginal_Dataset\n")
	for _, v := range histslice {
		fmt.Printf("%v\t%v\t%v\n", v.Value, v.Count, v.Original)
	}
}

func (e Results) FisherExactTest(nAllGenes int) float64 {
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

	nAllMendelianGenes := len(e.MendelianGenes)

	// Mendelian genes that we did not find in this permutation
	notfoundMendelianGenes := nAllMendelianGenes - e.MendelianGeneCount()

	residualGenes := nAllGenes - e.NonMendelianGeneCount() - e.MendelianGeneCount()

	_, _, _, twop := fet.FisherExactTest(e.MendelianGeneCount(), notfoundMendelianGenes, e.NonMendelianGeneCount(), residualGenes)

	return twop
}

func (e Results) NonMendelianGeneCount() int {
	n := 0
	for _, v := range e.Permutations {
		n += v.NonMendelianGenesNearLoci(e.MendelianGenes, e.Radius)
	}

	return n
}

func (e Results) MendelianGeneCount() int {
	n := 0
	for _, v := range e.Permutations {
		n += v.MendelianGenesNearLoci(e.MendelianGenes, e.Radius)
	}

	return n
}

type Permutation struct {
	Loci []Locus
}

func (p Permutation) NonMendelianGenesNearLoci(mendelian map[string]Gene, radius float64) int {
	n := 0

	return n
}

func (p Permutation) MendelianGenesNearLoci(mendelian map[string]Gene, radius float64) int {
	n := 0

	for _, locus := range p.Loci {
		for _, gene := range mendelian {
			if locus.IsGeneWithinRadius(gene, radius) {
				n++
			}
		}
	}

	return n
}

type Locus struct {
	Index      int
	Chromosome string
	Position   int
}

func (l Locus) IsGeneWithinRadius(gene Gene, radius float64) bool {
	if gene.Chromosome != l.Chromosome {
		return false
	}

	if math.Abs(float64(gene.EarliestTranscriptStart)-float64(l.Position)) < radius*1000 ||
		math.Abs(float64(gene.LatestTranscriptEnd)-float64(l.Position)) < radius*1000 {
		return true
	}

	return false
}

func main() {
	var (
		mendelianGeneFile string
		SNPsnapFile       string
		radius            float64
		overrideMissing   bool
	)

	flag.StringVar(&mendelianGeneFile, "mendel", "", "Filename containing one gene symbol per line representing your Mendelian disease genes.")
	flag.StringVar(&SNPsnapFile, "snpsnap", "", "Filename containing SNPsnap output.")
	flag.Float64Var(&radius, "radius", 250, "Radius, in kilobases, to define whether part of a transcript is 'within' a given locus.")
	flag.BoolVar(&overrideMissing, "overridemissing", false, "If not every gene on your gene list can be mapped, proceed anyway?")
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
	if err != nil && !(strings.Contains(err.Error(), "ERR1:") && overrideMissing) {
		log.Fatalln(err)
	}

	for _, v := range mendelianTranscripts {
		fmt.Print(v, " ")
	}
	fmt.Println()

	permutations, err := ReadSNPsnap(SNPsnapFile)
	if err != nil {
		log.Fatalln(err)
	}

	permutations.MendelianGenes = mendelianTranscripts
	permutations.Radius = radius

	permutations.Summarize()
}

func ReadSNPsnap(fileName string) (Results, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return Results{}, err
	}
	defer f.Close()

	cr := csv.NewReader(f)
	cr.Comment = '#'
	cr.Comma = '\t'
	lines, err := cr.ReadAll()
	if err != nil {
		return Results{}, err
	}

	if len(lines) < 1 {
		return Results{}, fmt.Errorf("0 lines in SNPsnap file")
	}
	if len(lines[0]) < 1 {
		return Results{}, fmt.Errorf("First line of SNPsnap file has no entries")
	}

	// Make one permutation per column
	permutations := make([]Permutation, len(lines[0]))
	for k := range permutations {
		// Initialize
		permutations[k].Loci = make([]Locus, len(lines)-1)
	}

	// Each row is a SNP
Outer:
	for i, row := range lines {
		if len(row) < 1 {
			continue
		}

		if i == 0 {
			continue
		}

		// Each column is a different permutation
		for j, col := range row {
			parts := strings.Split(col, ":")
			if len(parts) < 2 {
				log.Printf("%s had %d parts, expected 2. Skipping\n", col, len(parts))
				continue Outer
			}
			intpos, err := strconv.Atoi(parts[1])
			if err != nil {
				log.Println("Len of bad col:", len(col))
				return Results{}, pfx.Err(err)
			}
			locus := Locus{Index: j, Chromosome: parts[0], Position: intpos}
			permutations[j].Loci[i-1] = locus
		}
	}

	return Results{Permutations: permutations}, nil
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
		return keepers, fmt.Errorf("ERR1: Your Mendelian set contained %d genes, but we could only map %d of them. Missing: %v", len(geneNames), len(keepers), missing)
	}

	return keepers, nil
}

func AbsInt(a1, a2 int) int {
	if a1 > a2 {
		return a1 - a2
	}

	return a2 - a1
}
