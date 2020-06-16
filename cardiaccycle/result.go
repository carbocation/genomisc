package cardiaccycle

type Result struct {
	Identifier          string
	Column              string
	MaxOneStepShift     float64 // Biggest change in pixel area between two adjacent steps
	InstanceNumberAtMin uint16
	Min                 float64
	SmoothedMin         float64
	InstanceNumberAtMax uint16
	Max                 float64
	SmoothedMax         float64
	Window              int
	Discards            int
}
