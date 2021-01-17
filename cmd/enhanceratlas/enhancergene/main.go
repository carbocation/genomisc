package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	var filename string

	flag.StringVar(&filename, "filename", "", "File containing EnhancerAtlas.org _EP_v2 style data, which is enhancer-gene data")
	flag.Parse()

	if filename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	r := csv.NewReader(f)
	r.Comma = '\t'
	recs, err := r.ReadAll()
	if err != nil {
		return err
	}

	fmt.Printf("tissue\tchr\tstart_grch37\tstop_grch37\tensembl_id\tgene_name\n")

	for _, row := range recs {
		if len(row) != 2 {
			continue
		}

		ep := EP{}
		ep.Tissue = strings.ReplaceAll(filepath.Base(filename), "_EP.txt", "")

		score, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return err
		}

		parts := strings.Split(row[0], "_")
		if x := len(parts); x != 2 {
			return fmt.Errorf("Expected 2 parts when splitting by _ but found %d", x)
		}

		chpos := strings.Split(parts[0], ":")
		startstop := strings.Split(chpos[1], "-")

		ep.Chromosome = strings.ReplaceAll(chpos[0], "chr", "")
		ep.Start = startstop[0]
		ep.Stop = startstop[1]
		ep.EnhancerGeneInteractionScore = score

		geneparts := strings.Split(parts[1], "$")

		ep.EnsemblID = geneparts[0]
		ep.GeneName = geneparts[1]

		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", ep.Tissue, ep.Chromosome, ep.Start, ep.Stop, ep.EnsemblID, ep.GeneName)
	}

	return nil
}

type EP struct {
	Tissue                       string
	Chromosome                   string
	Start                        string
	Stop                         string
	EnsemblID                    string
	GeneName                     string
	EnhancerGeneInteractionScore float64
}
