package main

// Generate the String() function from the command line with: stringer -type=StatusEnum
type StatusEnum int

// Simplify replaces the LostToFollowUp code with NoDisease, since lost to
// follow-up is probably not an interesting status for most analyses.
func (s StatusEnum) Simplify() StatusEnum {
	if s == LostToFollowUp {
		return NoDisease
	}

	return s
}

const (
	Excluded StatusEnum = iota - 1
	NoDisease
	Disease
	Died
	LostToFollowUp
)
