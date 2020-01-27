package overlay

type Coord struct {
	X, Y int
}

type ConnectedComponent struct {
	LabelID     uint8
	ComponentID uint32
	PixelCount  int
	Bounds      struct {
		TopLeft     Coord
		BottomRight Coord
	}
}
