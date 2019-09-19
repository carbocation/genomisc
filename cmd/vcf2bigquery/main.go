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
	"sort"
	"strings"
	"sync"

	"github.com/brentp/bix"
	"github.com/brentp/irelate/interfaces"
	"github.com/brentp/vcfgo"
)

// Special value that is to be set using ldflags
// E.g.: go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"
// Consider aliasing in .profile: alias gobuild='go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"'
var builddate string

var (
	BufferSize = 4096 * 8
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
)

var (
	// Because globals are the signature of really great, not-lazy-at-all
	// programming
	keepMissing bool
	keepAlt     bool
)

func main() {
	defer STDOUT.Flush()

	fmt.Fprintf(os.Stderr, "This vcf2bigquery binary was built at: %s\n", builddate)

	var sampleFields flagSlice
	var vcfFile, assembly, chromosome string
	var chunk, chunksize, startPos, endPos int
	flag.StringVar(&chromosome, "chromosome", "", "If set, only extracts from one specific chromosome.")
	flag.IntVar(&startPos, "start_pos", 0, "In kilobases. If set, only extracts from this position onward within the specified chromosome.")
	flag.IntVar(&endPos, "end_pos", 0, "In kilobases. If set, only extracts until this position within the specified chromosome.")
	flag.StringVar(&vcfFile, "vcf", "", "Path to VCF containing diploid genotype data to be linearized.")
	flag.StringVar(&assembly, "assembly", "", "Name of assembly. Must be grch37 or grch38.")
	flag.IntVar(&chunksize, "chunksize", 0, "Use this chunksize (in kilobases).")
	flag.IntVar(&chunk, "chunk", 0, "1-based: which chunk to iterate over (if --chunksize is set).")
	flag.BoolVar(&keepMissing, "missing", false, "Print missing genotypes? (Will disable printing of ref alleles)")
	flag.BoolVar(&keepAlt, "alt", false, "Print genotypes with at least one non-reference allele? (Will disable printing of ref alleles)")
	flag.Var(&sampleFields, "field", "Fields to keep (other than GT, which is automatically included). Pass once per additional field, e.g., --field DP --field TLOD")

	flag.Parse()

	var err error
	var chunks []TabixLocus

	if chunk == 0 && (startPos != 0 || endPos != 0) {
		log.Fatalln("--start_pos and --end_pos can only be set for chunked jobs (--chunk)")
	}

	if chunksize > 0 {
		// Kilobases (kibibases, I guess)
		chunks, err = SplitChrPos(chunksize*1000, assembly, chromosome, startPos*1000, endPos*1000)
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("Split genome into %d chunks of ~%d kilobases each\n", len(chunks), chunksize)

		if chunk == 0 {
			log.Printf("--chunk was not specified. Specify --chunk between 1 and %d (inclusive)\n", len(chunks))
			flag.PrintDefaults()
			os.Exit(1)
		} else {
			log.Printf("This job will process chunk #%d\n", chunk)
		}
	}

	if keepAlt || keepMissing {
		log.Println("Reference alleles will *not* be printed.")
	}
	if keepAlt {
		log.Println("Alt alleles will be printed.")
	}
	if keepMissing {
		log.Println("Missing alleles will be printed.")
	}

	if len(sampleFields) > 0 {
		sort.StringSlice(sampleFields).Sort()
		log.Printf("In addition to genotype, will also fetch these sample fields: %s\n", strings.Join(sampleFields, ","))
	}

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

	go printer(completedWork, &pool, sampleFields)

	log.Println("Limiting concurrent goroutines to", runtime.NumCPU()*2)
	concurrencyLimit := make(chan struct{}, runtime.NumCPU()*2)

	log.Println("Linearizing", vcfFile)

	if len(sampleFields) > 0 {
		fmt.Fprintf(STDOUT, "sample_id\tchromosome\tposition_%s\tref\talt\trsid\tgenotype\t%s\n", assembly, strings.Join(sampleFields, "\t"))
	} else {
		fmt.Fprintf(STDOUT, "sample_id\tchromosome\tposition_%s\tref\talt\trsid\tgenotype\n", assembly)
	}

	if chunksize == 0 {

		// Read the full VCF

		if err := ReadAllVCF(rdr, concurrencyLimit, &pool, completedWork, sampleFields); err != nil {
			log.Println(err)
		}
	} else {

		// Read only a subset via tabix
		locus := chunks[chunk-1]
		if err := ReadTabixVCF(rdr, vcfFile, []TabixLocus{locus}, concurrencyLimit, &pool, completedWork, sampleFields); err != nil {
			log.Println(err)
		}

	}

	// Make sure that the last variants have finished printing before the
	// program is allowed to exit.
	pool.Wait()
	log.Println("Completed")
}

func ReadTabixVCF(rdr *vcfgo.Reader, vcfFile string, loci []TabixLocus, concurrencyLimit chan struct{}, pool *sync.WaitGroup, completedWork chan Work, sampleFields []string) error {

	tbx, err := bix.New(vcfFile)
	if err != nil {
		return err
	}
	defer tbx.Close()

	summarized := false

	j := 0
	for _, locus := range loci {
		vals, err := tbx.Query(locus)
		if err != nil {
			return err
		}
		defer vals.Close()

		for i := 0; ; i++ {
			j++
			v, err := vals.Next()
			if err != nil && err != io.EOF {
				// True error
				return err
			} else if err == io.EOF {
				// Finished all data.
				break
			}

			// Unwrap multiple layers to get to vcfgo.Variant{}

			v2, ok := v.(interfaces.VarWrap)
			if !ok {
				return fmt.Errorf(v.Chrom(), v.Source(), v.End(), "Not valid VarWrap")
			}

			snp, ok := v2.IVariant.(*vcfgo.Variant)
			if !ok {
				return fmt.Errorf(v.Chrom(), v.Source(), v.End(), "Not valid IVariant")
			}

			if i == 0 && !summarized {
				// Prime the integer sample lookup (since previously, a lot of time
				// was spent in runtime.mapaccess1_faststr, indicating that map
				// lookups were taking a meaningful amount of time)
				snp.Header.ParseSamples(snp)

				log.Println(len(snp.Samples), "samples found in the VCF")

				summarized = true
			}

			if i%1000 == 0 {
				log.Printf("Processed %d variants. Last %s:%d\n", i, snp.Chrom(), snp.Pos)
			}

			ProcessVariant(snp, concurrencyLimit, pool, completedWork, sampleFields)
		}
	}
	log.Printf("Processed %d variants.\n", j)

	return nil
}

func ReadAllVCF(rdr *vcfgo.Reader, concurrencyLimit chan struct{}, pool *sync.WaitGroup, completedWork chan Work, sampleFields []string) error {
	if rdr == nil {
		panic("Nil reader")
	}

	i := 0
	for ; ; i++ {
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

		if i%1000 == 0 {
			log.Printf("Processed %d variants. Last %s:%d\n", i, variant.Chrom(), variant.Pos)
		}

		ProcessVariant(variant, concurrencyLimit, pool, completedWork, sampleFields)
	}
	log.Printf("Processed %d variants.\n", i)
	if err := rdr.Error(); err != nil {
		return err
	}

	return nil
}

func ProcessVariant(variant *vcfgo.Variant, concurrencyLimit chan struct{}, pool *sync.WaitGroup, completedWork chan Work, sampleFields []string) {
	// Iterating over the variant.Alt() allows for multiallelics.
	for alleleID := range variant.Alt() {
		concurrencyLimit <- struct{}{}
		pool.Add(1)
		go worker(variant, alleleID, completedWork, concurrencyLimit, pool, sampleFields)
	}
}
