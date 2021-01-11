// MendelOverlap performs permutation to quantify surprise at the number of GWAS
// loci that overlap with Mendelian genes for your trait of interest. GRCh37
// only.
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

func main() {
	var (
		mendelianGeneFile   string
		SNPsnapFile         string
		truthLociFile       string
		radius              float64
		overrideMissing     bool
		transcriptStartOnly bool
		repeat              int
	)

	fmt.Println("This program uses GRCh37")
	flag.StringVar(&mendelianGeneFile, "mendel", "", "Filename containing one gene symbol per line representing your Mendelian disease genes.")
	flag.StringVar(&SNPsnapFile, "snpsnap", "", "Filename containing SNPsnap output.")
	flag.StringVar(&truthLociFile, "truthloci", "", "Optional. Filename containing truth loci (one chr:pos per line). If set, overrides the first column in --snpsnap file as the representative of the real loci from your study. Must *not* contain a header.")
	flag.Float64Var(&radius, "radius", 250, "Radius, in kilobases, to define whether part of a transcript is 'within' a given locus.")
	flag.BoolVar(&overrideMissing, "overridemissing", false, "If not every gene on your gene list can be mapped, proceed anyway?")
	flag.BoolVar(&transcriptStartOnly, "transcriptstart", false, "Measure radius to the transcript start site only? If false, will measure radius to start or end of the transcript (whichever is closer).")
	flag.IntVar(&repeat, "repeat", 1, "Iterate over the SNPSnap input this many times.")
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
	} else if err != nil && overrideMissing {
		fmt.Printf("Because --overridemissing is enabled, proceeding despite the following error: %s\n", err.Error())
	}

	permutations, err := ReadSNPsnap(SNPsnapFile, true)
	if err != nil {
		log.Fatalln(err)
	}

	if truthLociFile != "" {
		yourPermutations, err := ReadSNPsnap(truthLociFile, false)
		if err != nil {
			log.Fatalln(err)
		}

		permutations.Permutations[0].Loci = yourPermutations.Permutations[0].Loci
	}

	if transcriptStartOnly {
		fmt.Println("Using a radius of", radius, "kilobases from the transcript start (or within the transcript itself) only")
	} else {
		fmt.Println("Using a radius of", radius, "kilobases")
	}

	permutations.MendelianGenes = mendelianTranscripts
	permutations.Radius = radius

	permutations.Summarize(repeat, transcriptStartOnly, mendelianGeneFile, SNPsnapFile, radius)
}
