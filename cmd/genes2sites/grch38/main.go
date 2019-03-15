// MendelOverlap performs permutation to quantify surprise at the number of GWAS
// loci that overlap with Mendelian genes for your trait of interest. GRCh37
// only.
package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"
)

const BioMartFilename = "ensembl.grch38.p12.genes"

func main() {
	var (
		mendelianGeneFile string
		overrideMissing   bool
	)

	log.Println("This program uses GRCh38")
	flag.StringVar(&mendelianGeneFile, "genes", "", "Filename containing one gene symbol per line representing your genes.")
	flag.BoolVar(&overrideMissing, "overridemissing", false, "If not every gene on your gene list can be mapped, proceed anyway?")
	flag.Parse()

	if mendelianGeneFile == "" {
		flag.PrintDefaults()
		return
	}

	mendelianGeneList, err := ReadMendelianGeneFile(mendelianGeneFile)
	if err != nil {
		log.Fatalln(err)
	}

	if len(mendelianGeneList) < 1 {
		log.Println("No genes were parsed from your gene file")
	}

	mendelianTranscripts, err := SimplifyTranscripts(mendelianGeneList)
	if err != nil && !(strings.Contains(err.Error(), "ERR1:") && overrideMissing) {
		log.Println(err)
		log.Fatalln("You may re-run with --overridemissing if you are confident that missing these genes is acceptable")
	}

	geneSlice := make([]Gene, 0, len(mendelianTranscripts))
	for _, v := range mendelianTranscripts {
		geneSlice = append(geneSlice, v)
	}

	sort.Slice(geneSlice, func(i, j int) bool {
		if geneSlice[i].Symbol < geneSlice[j].Symbol {
			return true
		}

		return false
	})

	if len(geneSlice) < 1 {
		log.Println("No genes identified")
		return
	}

	for _, v := range geneSlice {
		fmt.Printf("%s\t%d\t%d\t%s\n", v.Chromosome, v.EarliestTranscriptStart, v.LatestTranscriptEnd, v.Symbol)
	}
}
