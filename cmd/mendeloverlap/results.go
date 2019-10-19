package main

import (
	"fmt"
	"sort"

	fet "github.com/glycerine/golang-fisher-exact"
)

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

type geneWithLoci struct {
	Gene
	Loci Loci
}

func (e Results) Summarize(repeat int, transcriptStartOnly bool) {
	origValue := e.Permutations[0].MendelianGenesNearLoci(e.MendelianGenes, e.Radius, transcriptStartOnly)

	mendelianCounts := make(map[int]Hist)
	for i := 0; i < repeat; i++ {
		for j, permutation := range e.Permutations {
			if i > 0 && j == 0 {
				// Only visit the truth case once
				continue
			}
			val := permutation.MendelianGenesNearLoci(e.MendelianGenes, e.Radius, transcriptStartOnly)
			hist := mendelianCounts[val]
			hist.Value = val
			hist.Count++
			mendelianCounts[val] = hist
		}
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

	// All Mendelian genes
	fmt.Println(len(e.MendelianGenes), "Mendelian genes were tested in the gene panel:")
	mgenes := make([]Gene, 0, len(e.MendelianGenes))
	for _, v := range e.MendelianGenes {
		mgenes = append(mgenes, v)
	}
	sort.Slice(mgenes, func(i, j int) bool { return mgenes[i].Symbol < mgenes[j].Symbol })
	for i, v := range mgenes {
		if i > 0 && i%15 == 0 {
			fmt.Println()
		}
		fmt.Print(v.Symbol, " ")
	}
	fmt.Println()
	fmt.Println()

	// Mendelian genes that overlap with your loci
	seengenemap := make(map[string]int)
	yourMgenes := make([]geneWithLoci, 0, len(e.MendelianGenes))
	for _, v := range e.MendelianGenes {
		for _, locus := range e.Permutations[0].Loci {
			if locus.IsGeneWithinRadius(v, e.Radius, transcriptStartOnly) {
				if seenGeneID, exists := seengenemap[v.Symbol]; exists {
					geneLoc := yourMgenes[seenGeneID]
					geneLoc.Loci = append(geneLoc.Loci, locus)
					yourMgenes[seenGeneID] = geneLoc
					continue
				}

				geneLoc := geneWithLoci{
					Gene: v,
					Loci: []Locus{locus},
				}

				yourMgenes = append(yourMgenes, geneLoc)
				seengenemap[v.Symbol] = len(yourMgenes) - 1
			}
		}
	}
	fmt.Printf("%d Mendelian genes overlapped with your original SNP list:\n", len(yourMgenes))
	sort.Slice(yourMgenes, func(i, j int) bool { return yourMgenes[i].Symbol < yourMgenes[j].Symbol })
	for i, v := range yourMgenes {
		fmt.Printf("%d) %s %s\n", i+1, v.Symbol, v.Loci)
	}

	fmt.Println()
	fmt.Println("Examined", repeat*len(e.Permutations), "permutations")
	fmt.Printf("N_Overlapping_Loci\tN_Permutations\tContains_Original_Dataset\n")
	equallyOrMoreExtreme := 0
	for _, v := range histslice {
		fmt.Printf("%v\t%v\t%v\n", v.Value, v.Count, v.Original)
		if v.Original {
			equallyOrMoreExtreme += v.Count
		} else if equallyOrMoreExtreme > 0 {
			// Once we have hit our original value, we add the count from every
			// subsequent entry with that much overlap or higher.
			equallyOrMoreExtreme += v.Count
		}
	}
	fmt.Println()
	fmt.Printf("Approximate one-tailed P-value: P < %.1e\n", float64(equallyOrMoreExtreme)/float64(repeat*len(e.Permutations)))
}

func (e Results) FisherExactTest(nAllGenes int, transcriptStartOnly bool) float64 {
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
	notfoundMendelianGenes := nAllMendelianGenes - e.MendelianGeneCount(transcriptStartOnly)

	residualGenes := nAllGenes - e.NonMendelianGeneCount() - e.MendelianGeneCount(transcriptStartOnly)

	_, _, _, twop := fet.FisherExactTest(e.MendelianGeneCount(transcriptStartOnly), notfoundMendelianGenes, e.NonMendelianGeneCount(), residualGenes)

	return twop
}

func (e Results) NonMendelianGeneCount() int {
	n := 0
	for _, v := range e.Permutations {
		n += v.NonMendelianGenesNearLoci(e.MendelianGenes, e.Radius)
	}

	return n
}

func (e Results) MendelianGeneCount(transcriptStartOnly bool) int {
	n := 0
	for _, v := range e.Permutations {
		n += v.MendelianGenesNearLoci(e.MendelianGenes, e.Radius, transcriptStartOnly)
	}

	return n
}
