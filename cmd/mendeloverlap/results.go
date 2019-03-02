package main

import (
	"fmt"
	"sort"

	"github.com/glycerine/golang-fisher-exact"
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

func (e Results) Summarize() {
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
	mgenes = make([]Gene, 0, len(e.MendelianGenes))
	for _, v := range e.MendelianGenes {
		for _, locus := range e.Permutations[0].Loci {
			if locus.IsGeneWithinRadius(v, e.Radius) {
				mgenes = append(mgenes, v)
			}
		}
	}
	fmt.Printf("%d Mendelian genes overlapped with your original SNP list:\n", len(mgenes))
	sort.Slice(mgenes, func(i, j int) bool { return mgenes[i].Symbol < mgenes[j].Symbol })
	for i, v := range mgenes {
		fmt.Printf("%d) %s\n", i+1, v.Symbol)
	}

	fmt.Println()
	fmt.Println("Examined", len(e.Permutations), "permutations")
	fmt.Printf("N_Overlapping_Loci\tN_Permutations\tContains_Original_Dataset\n")
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
