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

// MergeConnectedComponents brings two ConnectedComponents together. In doing
// so, it assumes that the LabelID is the same between the two (and sets it to
// the LabelID of the first, regardless). It unsets the component ID - this
// merged value is now only useful in the context of doing a summary level
// evaluation on the whole label.
func MergeConnectedComponentsSameLabel(c1, c2 ConnectedComponent) ConnectedComponent {
	c1.ComponentID = 0
	c1.PixelCount += c2.PixelCount
	if b := c2.Bounds.TopLeft.X; b < c1.Bounds.TopLeft.X {
		c1.Bounds.TopLeft.X = b
	}
	if b := c2.Bounds.TopLeft.Y; b < c1.Bounds.TopLeft.Y {
		c1.Bounds.TopLeft.Y = b
	}
	if b := c2.Bounds.BottomRight.X; b > c1.Bounds.BottomRight.X {
		c1.Bounds.BottomRight.X = b
	}
	if b := c2.Bounds.BottomRight.Y; b > c1.Bounds.BottomRight.Y {
		c1.Bounds.BottomRight.Y = b
	}

	return c1
}
