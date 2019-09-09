package tasker

import "fmt"

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

func (t *Task) Print(inputBucket, outputBucket, layout, sourceFileName, overrideName string) {
	if t == nil {
		return
	}

	if overrideName == "" {
		overrideName = sourceFileName
	}

	fmt.Printf("%s\t%d\t%d\t%s\t%s\t%s\t%s\n", t.Chromosome, t.FirstLine, t.LastLine, inputBucket, fmt.Sprintf("%s/%s-%d.tsv", outputBucket, sourceFileName, t.ID), layout, overrideName)
}
