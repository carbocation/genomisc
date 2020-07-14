package tasker

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/carbocation/genomisc"
	"github.com/carbocation/genomisc/prsparser"
)

func AddPRSParserLayout(ldp prsparser.Layout, name string) {
	prsparser.Layouts[name] = ldp
}

func CreateTasks(
	inputBucket string,
	outputBucket string,
	prsPath string,
	layout string,
	hasHeader bool,
	variantsPerJob int,
	overrideName string,
	customLayout string,
) error {

	parser, err := prsparser.New(layout)
	if err != nil {
		return fmt.Errorf("CreatePRSParserErr: %v", err)
	}

	// Open PRS file
	f, err := os.Open(prsPath)
	if err != nil {
		return fmt.Errorf("FileOpenErr: %v", err)
	}
	defer f.Close()

	fd, err := genomisc.MaybeDecompressReadCloserFromFile(f)
	if err != nil {
		return fmt.Errorf("CompressionDetectionErr: %v", err)
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
				return fmt.Errorf("OneByteErr: %v", err)
			}
			if oneByte[0] == '\n' {
				break
			}
		}
	}

	sourceFileName := path.Base(prsPath)

	layoutCol := "layout"
	if layout == "CUSTOM" {
		layoutCol = "custom_layout"
		layout = customLayout
	}

	fmt.Printf("--env chrom\t--env firstline\t--env lastline\t--input infile\t--output outfile\t--env %s\t--env source\n", layoutCol)

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
			return fmt.Errorf("ReaderErr: %v", err)
		}

		if i%250000 == 0 {
			log.Println("Observed entry ", i)
		}

		val, err := parser.ParseRow(row)
		if err != nil {
			return fmt.Errorf("ParseRowErr: %v", err)
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
				task.Print(inputBucket, outputBucket, layout, sourceFileName, overrideName)
				task.Clear()
				taskCount++
			}
		}

		if active := fileMap[val.Chromosome]; !active {
			return fmt.Errorf("Asked to process a variant on chr%s (%d), but we previously saw this chromosome already", val.Chromosome, val.Position)
		}

		// We're still on the same active chromosome

		// Is this task too full?
		if task.Count() >= variantsPerJob {
			task.Print(inputBucket, outputBucket, layout, sourceFileName, overrideName)
			task.Clear()
			taskCount++
		}

		task.AddLine(taskCount, i, val.Chromosome)
	}

	// Final cleanup of the last task
	if task.Active() {
		task.Print(inputBucket, outputBucket, layout, sourceFileName, overrideName)
		task.Clear()
		taskCount++
	}

	log.Printf("Processed %d entries into %d tasks\n", i, taskCount)

	return nil
}
