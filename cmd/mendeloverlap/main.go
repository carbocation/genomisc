// MendelOverlap performs permutation to quantify surprise at the number of GWAS
// loci that overlap with Mendelian genes for your trait of interest. GRCh37
// only.
package main

import (
	"flag"
	"log"
	"strings"
)

func main() {
	var (
		mendelianGeneFile string
		SNPsnapFile       string
		radius            float64
		overrideMissing   bool
	)

	log.Println("This program uses GRCh37")
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
		log.Println(err)
		log.Fatalln("You may re-run with --overridemissing if you are confident that missing these genes is acceptable")
	}

	permutations, err := ReadSNPsnap(SNPsnapFile)
	if err != nil {
		log.Fatalln(err)
	}

	permutations.MendelianGenes = mendelianTranscripts
	permutations.Radius = radius

	permutations.Summarize()
}
