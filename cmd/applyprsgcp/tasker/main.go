package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/carbocation/genomisc"
	"github.com/carbocation/genomisc/prsparser"
)

type Task struct {
	ID         int
	Chromosome string
	FirstLine  int
	LastLine   int

	touched bool
}

func (t *Task) Active() bool {
	return t.touched
}

func (t *Task) AddLine(id, line int, chromosome string) {
	if !t.touched {
		t.touched = true
		t.ID = id
		t.Chromosome = chromosome
		t.FirstLine = line
	}

	t.LastLine = line
}

func (t *Task) Count() int {
	return t.LastLine - t.FirstLine
}

func (t *Task) Clear() {
	t.ID = 0
	t.Chromosome = ""
	t.FirstLine = 0
	t.LastLine = 0
	t.touched = false
}

func (t *Task) Print(inputBucket, outputBucket, layout, sourceFileName string) {
	if t == nil {
		return
	}

	fmt.Printf("%s\t%d\t%d\t%s\t%s\t%s\t%s\n", t.Chromosome, t.FirstLine, t.LastLine, inputBucket, fmt.Sprintf("%s/%d.tsv", outputBucket, t.ID), layout, sourceFileName)
}

func init() {
	// Add a tab-delimited LDPred processor in addition to the space-delimited
	// one
	ldp := prsparser.Layouts["LDPRED"]
	ldp.Delimiter = '\t'
	prsparser.Layouts["LDPREDTAB"] = ldp
}

func main() {
	var (
		inputBucket    string
		outputBucket   string
		prsPath        string
		layout         string
		hasHeader      bool
		variantsPerJob int
	)
	flag.StringVar(&inputBucket, "input", "", "Google Storage bucket path where the PRS file will be found")
	flag.StringVar(&outputBucket, "output", "", "Google Storage bucket path where output files should go")
	flag.StringVar(&prsPath, "prs", "", "Path to text file containing your polygenic risk score. Must be sorted by chromosome.")
	flag.StringVar(&layout, "layout", "LDPRED", fmt.Sprint("Layout of your prs file. Currently, options include: ", prsparser.LayoutNames()))
	flag.BoolVar(&hasHeader, "header", true, "Does the input file have a header that needs to be skipped?")
	flag.IntVar(&variantsPerJob, "variants_per_job", 0, "Maximum number of variants to be processed by each individual worker")
	flag.Parse()

	// PRS
	if prsPath == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide the path to a file with your polygenic risk score effects.")
	}

	if outputBucket == "" || inputBucket == "" {
		flag.PrintDefaults()
		log.Fatalln("Please provide the path to a google bucket folder where the input can be found and the output can be placed")
	}

	if variantsPerJob <= 0 {
		flag.PrintDefaults()
		log.Fatalln("Variants per job cannot be set to 0 or fewer")
	}

	parser, err := prsparser.New(layout)
	if err != nil {
		log.Fatalln("CreatePRSParserErr:", err)
	}

	// Open PRS file
	f, err := os.Open(prsPath)
	if err != nil {
		log.Fatalln("FileOpenErr:", err)
	}
	defer f.Close()

	fd, err := genomisc.MaybeDecompressReadCloserFromFile(f)
	if err != nil {
		log.Fatalln("CompressionDetectionErr:", err)
	}
	defer fd.Close()

	if hasHeader {
		// Skip the header. Fun fact: some ldpred files have the wrong
		// number of header fields, so if you do this at the CSV step, you
		// get an error
		var oneByte [1]byte
		for {
			_, err := fd.Read(oneByte[:])
			if err != nil {
				log.Fatalln("OneByteErr:", err)
			}
			if oneByte[0] == '\n' {
				break
			}
		}
	}

	sourceFileName := path.Base(prsPath)

	fmt.Printf("--env chrom\t--env firstline\t--env lastline\t--input infile\t--output outfile\t--env layout\t--env source\n")

	reader := csv.NewReader(fd)
	reader.Comma = parser.CSVReaderSettings.Comma
	reader.Comment = parser.CSVReaderSettings.Comment
	reader.TrimLeadingSpace = parser.CSVReaderSettings.TrimLeadingSpace

	// Parse and define first/last variant for each task
	fileMap := make(map[string]bool)

	// Track our output
	task := &Task{}
	taskCount := 0

	i := 0
	if hasHeader {
		i = 1
	}
	for ; ; i++ {
		row, err := reader.Read()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln("ReaderErr:", err)
		}

		if i%250000 == 0 {
			log.Println("Observed entry ", i)
		}

		val, err := parser.ParseRow(row)
		if err != nil {
			log.Fatalln("ParseRowErr:", err)
		}

		if _, exists := fileMap[val.Chromosome]; !exists {
			// First variant for this chromosome.
			// 1) Inactivate other chromosomes.
			for k := range fileMap {
				fileMap[k] = false
			}

			// 2) Create active entry for the new one
			fileMap[val.Chromosome] = true

			// 3) The task is complete
			if task.Active() {
				task.Print(inputBucket, outputBucket, layout, sourceFileName)
				task.Clear()
				taskCount++
			}
		}

		if active := fileMap[val.Chromosome]; !active {
			log.Fatalf("Asked to process a variant on chr%s (%d), but we previously saw this chromosome already.\n", val.Chromosome, val.Position)
		}

		// We're still on the same active chromosome

		// Is this task too full?
		if task.Count() >= variantsPerJob {
			task.Print(inputBucket, outputBucket, layout, sourceFileName)
			task.Clear()
			taskCount++
		}

		task.AddLine(taskCount, i, val.Chromosome)
	}

	// Final cleanup of the last task
	if task.Active() {
		task.Print(inputBucket, outputBucket, layout, sourceFileName)
		task.Clear()
		taskCount++
	}

	log.Printf("Processed %d entries into %d tasks\n", i, taskCount)

}
