package main

import "strconv"

type sortable struct {
	series     string
	position   float64
	lineNumber int
}

func newSortable(line []string, series, position, lineNumber int) (sortable, error) {
	s := sortable{}

	s.lineNumber = lineNumber
	s.series = line[series]
	pos, err := strconv.ParseFloat(line[position], 64)
	if err != nil {
		return s, err
	}
	s.position = pos

	return s, nil
}

func linesToSortables(input [][]string, series, position int) ([]sortable, error) {
	output := make([]sortable, 0, len(input))

	for k, line := range input {
		l, err := newSortable(line, series, position, k)
		if err != nil {
			return nil, err
		}
		output = append(output, l)
	}

	return output, nil
}
