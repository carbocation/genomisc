package main

import (
	"fmt"
)

type ErrorInfo struct {
	Chromosome string
	Position   uint32
	Message    string
}

func (e ErrorInfo) Error() string {
	return fmt.Sprintf("Chromosome: %s, Position: %d, Message: %s", e.Chromosome, e.Position, e.Message)
}
