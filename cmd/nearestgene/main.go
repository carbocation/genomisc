// nearestgene finds the single gene whose transcript start site is closest to a
// given chr:pos. O(N^2) so only use it for small data.
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
)

var BioMartFilename string

var assemblies = map[string]string{
	"37": "ensembl.grch37.p13.genes",
	"38": "ensembl.grch38.p12.genes",
}

func main() {
	var (
		sitesFile string
		assembly  string
		tss       bool
	)

	flag.StringVar(&sitesFile, "sites", "", "Filename containing one site per line (represented as chr:pos)")
	flag.StringVar(&assembly, "assembly", "37", fmt.Sprint("Version of genome assembly. Options:", assemblies))
	flag.BoolVar(&tss, "transcriptstart", true, "Measure distance from transcript start site (default true). Note that if a site is on a transcript, it will be assigned to that transcript (even if there is a closer TSS). If false, then distance will be measured from the start or the end (whichever is closer).")
	flag.Parse()

	BioMartFilename = assemblies[assembly]

	if sitesFile == "" {
		flag.PrintDefaults()
		return
	}

	log.Println("Using", BioMartFilename)
	if tss {
		log.Println("Measuring distance from the transcript start site")
	} else {
		log.Println("Measuring distance from any part of the transcript")
	}

	sitesList, err := ReadSitesFile(sitesFile)
	if err != nil {
		log.Fatalln(err)
	}

	if len(sitesList) < 1 {
		log.Fatalln("No sites were parsed from your sites file")
	}

	transcripts, err := FetchGenes()
	if err != nil && !(strings.Contains(err.Error(), "ERR1:")) {
		log.Fatalln(err)
	}

	type outputType struct {
		Site                    string
		Chromosome              string
		Position                int
		GeneName                string
		TranscriptStart         int
		TranscriptEnd           int
		DistanceTranscriptStart int
		DistanceTranscriptEnd   int
		Distance                int
		OnTranscript            bool
	}

	output := make([]outputType, 0)

	for site := range sitesList {
		parts := strings.Split(site, ":")
		if len(parts) != 2 {
			log.Fatalf("%s cannot be split into exactly 2 parts\n", site)
		}

		siteChr := parts[0]
		sitePosInt, err := strconv.Atoi(parts[1])
		if err != nil {
			log.Fatalln(err)
		}
		sitePos := float64(sitePosInt)

		sort.Slice(transcripts, func(i, j int) bool {
			if transcripts[i].Chromosome != siteChr {
				return false
			}

			if tss {
				// If it is on a transcript of one but not the other, keep the one where it's on the transcript
				onTranscriptI, onTranscriptJ := false, false
				if sitePosInt >= transcripts[i].TranscriptStart && sitePosInt <= transcripts[i].TranscriptEnd {
					onTranscriptI = true
				}
				if sitePosInt >= transcripts[j].TranscriptStart && sitePosInt <= transcripts[j].TranscriptEnd {
					onTranscriptJ = true
				}
				if onTranscriptI && !onTranscriptJ {
					return true
				} else if onTranscriptJ && !onTranscriptI {
					return false
				}

				// Otherwise (if it's on both or neither transcript), find the distance to the nearest TSS
				if math.Abs(sitePos-float64(transcripts[i].TranscriptStart)) < math.Abs(sitePos-float64(transcripts[j].TranscriptStart)) {
					return true
				}
			} else {
				closestI := math.Min(math.Abs(sitePos-float64(transcripts[i].TranscriptStart)), math.Abs(sitePos-float64(transcripts[i].TranscriptEnd)))
				closestJ := math.Min(math.Abs(sitePos-float64(transcripts[j].TranscriptStart)), math.Abs(sitePos-float64(transcripts[j].TranscriptEnd)))
				if closestI < closestJ {
					return true
				}
			}

			return false
		})

		onTranscript := false
		if sitePosInt >= transcripts[0].TranscriptStart && sitePosInt <= transcripts[0].TranscriptEnd {
			onTranscript = true
		}

		output = append(output, outputType{
			Site:                    site,
			Chromosome:              siteChr,
			Position:                sitePosInt,
			GeneName:                transcripts[0].Symbol,
			TranscriptStart:         transcripts[0].TranscriptStart,
			TranscriptEnd:           transcripts[0].TranscriptEnd,
			DistanceTranscriptStart: int(math.Abs(float64(transcripts[0].TranscriptStart) - sitePos)),
			DistanceTranscriptEnd:   int(math.Abs(float64(transcripts[0].TranscriptEnd) - sitePos)),
			Distance:                int(math.Abs(float64(transcripts[0].TranscriptStart) - sitePos)),
			OnTranscript:            onTranscript,
		})
		if !tss {
			output[len(output)-1].Distance = int(math.Min(float64(output[len(output)-1].DistanceTranscriptStart), float64(output[len(output)-1].DistanceTranscriptEnd)))
			if onTranscript {
				output[len(output)-1].Distance = 0
			}
		}
	}

	if x, y := len(output), len(sitesList); x < y {
		log.Fatalf("Identified genes for %d fewer sites than expected\n", y-x)
	}

	fmt.Printf("Site\tChromosome\tPosition\tGeneName\tTranscriptStart\tTranscriptEnd\tDistanceTranscriptStart\tDistanceTranscriptEnd\tDistance\tOnTranscript\n")
	for _, v := range output {
		fmt.Printf("%s\t%s\t%d\t%s\t%d\t%d\t%d\t%d\t%d\t%t\n",
			v.Site, v.Chromosome, v.Position,
			v.GeneName, v.TranscriptStart, v.TranscriptEnd,
			v.DistanceTranscriptStart, v.DistanceTranscriptEnd,
			v.Distance, v.OnTranscript)
	}
}
