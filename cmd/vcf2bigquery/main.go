package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/brentp/vcfgo"
)

var (
	BufferSize = 4096 * 8
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
)

func main() {
	defer STDOUT.Flush()

	var vcfFile, assembly string
	flag.StringVar(&vcfFile, "vcf", "", "Path to VCF containing diploid genotype data to be linearized.")
	flag.StringVar(&assembly, "assembly", "", "Name of assembly. This will modify the name of the 'pos' column for your future reference.")
	flag.Parse()

	if vcfFile == "" || assembly == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Prime the VCF reader

	fraw, err := os.Open(vcfFile)
	if err != nil {
		log.Fatalln(err)
	}
	defer fraw.Close()

	var f io.Reader
	f, err = gzip.NewReader(fraw)
	if err != nil {
		f, _ = os.Open(vcfFile)
	}

	// Buffered reader
	buffRead := bufio.NewReaderSize(f, BufferSize)

	rdr, err := vcfgo.NewReader(buffRead, true) // Lazy genotype parsing, so we can avoid doing it on the main thread
	if err != nil {
		log.Println("Invalid VCF. Invalid features include:")
		log.Println(err)
		if rdr != nil {
			log.Println("Attempting to continue.")
			rdr.Clear()
		} else {
			log.Println("VCF reader could not be initialized due to the errors. Exiting.")
			return
		}
	}
	if err := rdr.Error(); err != nil {
		log.Println("Invalid VCF. Attempting to continue. Invalid features include:")
		log.Println(err)
		rdr.Clear()
	}

	// Receive work from goroutines over a channel with a large buffer
	completedWork := make(chan Work, runtime.NumCPU()*100)
	pool := sync.WaitGroup{}

	go printer(completedWork, &pool)

	log.Println("Limiting concurrent goroutines to", runtime.NumCPU()*2)
	concurrencyLimit := make(chan struct{}, runtime.NumCPU()*2)

	variantOrder := 0
	log.Println("Linearizing", vcfFile)

	fmt.Fprintf(STDOUT, "sample_id\tchromosome\tposition_%s\tref\talt\trsid\tgenotype\n", assembly)
	for i := 0; ; i++ {
		if rdr == nil {
			panic("Nil reader")
		}
		variant := rdr.Read()
		if variant == nil {
			log.Println("Finished")
			break
		}

		if i == 0 {
			// Prime the integer sample lookup (since previously, a lot of time
			// was spent in runtime.mapaccess1_faststr, indicating that map
			// lookups were taking a meaningful amount of time)
			variant.Header.ParseSamples(variant)

			log.Println(len(variant.Samples), "samples found in the VCF")
		}

		// Iterating over the variant.Alt() allows for multiallelics.
		for alleleID := range variant.Alt() {
			concurrencyLimit <- struct{}{}
			pool.Add(1)
			go worker(variant, alleleID, variantOrder, completedWork, concurrencyLimit, &pool)
			variantOrder++
		}
	}
	if err := rdr.Error(); err != nil {
		log.Println("Final errors:")
		log.Println(err)
	}

	// Make sure that the last variants have finished printing before the
	// program is allowed to exit.
	pool.Wait()

	// fmt.Fprintf(STDOUT, "%s\t%s\t%s\t%s\t%s\t%v\n", pheno.SampleID, pheno.FieldID, pheno.Instance, pheno.ArrayIDX, pheno.Value, pheno.CodingFileID)
}
