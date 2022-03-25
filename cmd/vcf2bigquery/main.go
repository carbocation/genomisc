package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/brentp/irelate/interfaces"
	"github.com/carbocation/bix"
	"github.com/carbocation/genomisc"
	_ "github.com/carbocation/genomisc/compileinfoprint"
	"github.com/carbocation/vcfgo"
)

// Special value that is to be set using ldflags
// E.g.: go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"
// Consider aliasing in .profile: alias gobuild='go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"'
var builddate string

var (
	BufferSize = 4096 * 32
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
	client     *storage.Client
)

var (
	// Because globals are the signature of really great, not-lazy-at-all
	// programming
	keepMissing    bool
	keepAlt        bool
	passFilterOnly bool
	siteID         bool
)

func main() {
	defer STDOUT.Flush()

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
	flag.Var(&sampleFields, "field", "Fields to keep (other than GT, which is automatically included). Pass once per additional field, e.g., --field DP --field TLOD. Note that you can force a look at the INFO field by prefixing with `INFO_` - useful in cases such as 'DP' which can exist per sample and across the study in the INFO field.")
	flag.BoolVar(&passFilterOnly, "passonly", false, "If true, will not print variants at sites that have values other than PASS or '.' in the FILTER field.")
	flag.BoolVar(&siteID, "siteid", false, "If true, will create a 'siteid' field which is 'CHR:POS:REF:ALT' and which should be a unique identifier for joins.")

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
			log.Fatalf("Error processing chunked VCF: %v\n", err)
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

	if chromosome != "" && chunksize == 0 {
		log.Printf("Chromosome was set, but chunksize was not. To enable per-chromosome variant extraction, set --chunk and --chunksize")
		flag.PrintDefaults()
		os.Exit(1)
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

	if passFilterOnly {
		log.Println("Will NOT print variants with FILTER values other than PASS or '.'")
	} else {
		log.Println("Will print variants regardless of FILTER field")
	}

	if siteID {
		log.Println("Will also create a 'siteid' field to uniquely identify this variant for joins")
	}

	if len(sampleFields) > 0 {
		sort.StringSlice(sampleFields).Sort()
		log.Printf("In addition to genotype, will also fetch these sample fields: %s\n", strings.Join(sampleFields, ","))
	}

	if vcfFile == "" || assembly == "" {
		log.Println("Please pass both -vcf and -assembly")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if strings.HasPrefix(vcfFile, "gs://") {
		var err error
		client, err = storage.NewClient(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}

	// Prime the VCF reader

	fraw, err := genomisc.MaybeOpenSeekerFromGoogleStorage(vcfFile, client)
	if err != nil {
		log.Fatalf("Error opening VCF: %s\n", err)
	}
	defer fraw.Close()

	var f io.Reader
	f, err = gzip.NewReader(fraw)
	if err != nil {
		fraw.Seek(0, 0)
		f = fraw
	}

	// Buffered reader
	buffRead := bufio.NewReaderSize(f, BufferSize)

	rdr, err := vcfgo.NewReader(buffRead, true) // Lazy genotype parsing, so we can avoid doing it on the main thread
	if err != nil {
		log.Println("Invalid VCF. Invalid features include:")
		log.Println(err)
		if rdr != nil {
			log.Println(rdr.Error())
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

	// Current structure is not safe for concurrent use because a vcfgo.Reader
	// is not safe for concurrent use. One approach to fix this would be to
	// launch N concurrent goroutines, each of which open the file and which are
	// assigned a modulus and then only process lines of that modulus. Whether
	// that is faster than one process isn't immediately obvious to me.
	log.Println("Limiting concurrent goroutines to 1")
	concurrencyLimit := make(chan struct{}, 1)

	log.Println("Linearizing", vcfFile)

	siteIDstring := ""
	if siteID {
		siteIDstring = "\tsiteid"
	}

	if len(sampleFields) > 0 {
		fmt.Fprintf(STDOUT, "sample_id\tchromosome\tposition_%s\tref\talt%s\trsid\tgenotype\t%s\n", assembly, siteIDstring, strings.Join(sampleFields, "\t"))
	} else {
		fmt.Fprintf(STDOUT, "sample_id\tchromosome\tposition_%s\tref\talt%s\trsid\tgenotype\n", assembly, siteIDstring)
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

	tbx, err := bix.NewGCP(vcfFile, client)
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

				if len(snp.Header.SampleFormats) > 0 {
					log.Println("SampleFormats for this file:")
					for _, v := range snp.Header.SampleFormats {
						log.Println(v.String())
					}
				}

				summarized = true
			}

			if i%1000 == 0 {
				log.Printf("Processed %d variants. Last %s:%d\n", i, snp.Chrom(), snp.Pos)
			}

			if passFilterOnly && (snp.Filter != "PASS" && snp.Filter != ".") {
				// Skip non-pass filter variants if toggled
				continue
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

			if len(variant.Header.SampleFormats) > 0 {
				log.Println("SampleFormats for this file:")
				for _, v := range variant.Header.SampleFormats {
					log.Println(v.String())
				}
			}
		}

		if i%1000 == 0 {
			log.Printf("Processed %d variants. Last %s:%d\n", i, variant.Chrom(), variant.Pos)
		}

		if passFilterOnly && (variant.Filter != "PASS" && variant.Filter != ".") {
			// Skip non-pass filter variants if toggled
			continue
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
