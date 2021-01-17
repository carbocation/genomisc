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
	var tall bool

	flag.StringVar(&filename, "filename", "", "File containing EnhancerAtlas.org _EP_v2 style data, which is enhancer-gene data")
	flag.BoolVar(&tall, "tall", false, "If true, creates one entry per chr-pos. If false, creates a summary entry indicating the pos start and stop positions.")
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

	if tall {
		fmt.Printf("tissue\tchr\tpos_grch37\tensembl_id\tgene_name\tscore\n")
	} else {
		fmt.Printf("tissue\tchr\tstart_grch37\tstop_grch37\tensembl_id\tgene_name\tscore\n")
	}

	for i, row := range recs {
		if len(row) != 2 {
			continue
		}

		ep := EP{}
		ep.Tissue = strings.ReplaceAll(filepath.Base(filename), "_EP.txt", "")

		score, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return err
		}

		parts := strings.SplitN(row[0], "_", 2)
		if x := len(parts); x != 2 {
			return fmt.Errorf("Line %d: Expected 2 parts when splitting by _ but found %d", i, x)
		}

		chpos := strings.Split(parts[0], ":")
		startstop := strings.Split(chpos[1], "-")

		start, err := strconv.ParseInt(startstop[0], 10, 64)
		if err != nil {
			return err
		}

		stop, err := strconv.ParseInt(startstop[1], 10, 64)
		if err != nil {
			return err
		}

		ep.Chromosome = strings.ReplaceAll(chpos[0], "chr", "")
		ep.Start = start
		ep.Stop = stop
		ep.EnhancerGeneInteractionScore = score

		geneparts := strings.Split(parts[1], "$")

		ep.EnsemblID = geneparts[0]
		ep.GeneName = geneparts[1]
		if tall {
			for i := ep.Start; i <= ep.Stop; i++ {
				fmt.Printf("%s\t%s\t%d\t%s\t%s\t%.3f\n", ep.Tissue, ep.Chromosome, i, ep.EnsemblID, ep.GeneName, ep.EnhancerGeneInteractionScore)
			}
		} else {
			fmt.Printf("%s\t%s\t%d\t%d\t%s\t%s\t%.3f\n", ep.Tissue, ep.Chromosome, ep.Start, ep.Stop, ep.EnsemblID, ep.GeneName, ep.EnhancerGeneInteractionScore)
		}
	}

	return nil
}

type EP struct {
	Tissue                       string
	Chromosome                   string
	Start                        int64
	Stop                         int64
	EnsemblID                    string
	GeneName                     string
	EnhancerGeneInteractionScore float64
}
