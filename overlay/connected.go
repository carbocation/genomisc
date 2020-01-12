package overlay

import (
	"image"

	"github.com/carbocation/unionfind"
)

// Following the guide at
// http://aishack.in/tutorials/connected-component-labelling/

type Connected struct {
	imgData [][]uint8
	labels  [][]uint32
}

func (l LabelMap) CountConnectedRegions(bmpImage image.Image) (map[Label]int, error) {
	conn, err := NewConnected(bmpImage)
	if err != nil {
		return nil, err
	}

	conn.Count(l)

	return nil, nil
}

func NewConnected(img image.Image) (Connected, error) {
	out := Connected{}

	imgData := make([][]uint8, 0, img.Bounds().Max.Y)
	labels := make([][]uint32, 0, img.Bounds().Max.Y)

	for y := 0; y < img.Bounds().Max.Y; y++ {
		pixelData := make([]uint8, 0, img.Bounds().Max.X)
		labelData := make([]uint32, img.Bounds().Max.X, img.Bounds().Max.X)

		for x := 0; x < img.Bounds().Max.X; x++ {
			c := img.At(x, y)

			id, err := LabeledPixelToID(c)
			if err != nil {
				return out, err
			}

			pixelData = append(pixelData, uint8(id))
		}

		imgData = append(imgData, pixelData)
		labels = append(labels, labelData)
	}

	out.imgData = imgData
	out.labels = labels

	return out, nil
}

func (c Connected) Count(l LabelMap) (map[Label]int, error) {
	uf := unionfind.NewThreadSafeUnionFind(25000)

	var nextLabel uint32 = 1
	for y, row := range c.imgData {
		for x := range row {

			// See if we already labeled an adjacent pixel.
			found := false
			foundUp, valUp := c.LabelAbove(x, y)
			if foundUp {
				found = true
				c.labels[y][x] = valUp
			}

			// If the left pixel's label is lower than that of the top pixel,
			// use the left pixel instead - plus now we know these two labels
			// need to be joined.
			foundLeft, val := c.LabelLeft(x, y)
			if foundLeft {
				found = true
				if foundUp {

					if val < c.labels[y][x] {
						c.labels[y][x] = val

						uf.Union(int(val), int(valUp))
					} else {
						uf.Union(int(valUp), int(val))
					}
				} else {
					c.labels[y][x] = val
				}
			}

			if found {
				continue
			}

			// If not, it gets its own label
			c.labels[y][x] = nextLabel
			nextLabel++
		}
	}

	// Now reconcile the adjacent labels
	for y, row := range c.labels {
		for x, v := range row {
			if root := uf.Root(int(v)); root < 0 {
				// No adjacent labels
				continue
			} else {
				if root32 := uint32(root); root32 < v {
					c.labels[y][x] = root32
				}
			}
		}
	}

	// Count number of components per label
	labels := make(map[uint8]map[uint32]struct{})
	mapper := make(map[uint8]Label)
	for _, v := range l.Sorted() {
		labels[uint8(v.ID)] = make(map[uint32]struct{})
		mapper[uint8(v.ID)] = v
	}
	for y, row := range c.labels {
		for x, v := range row {
			// Original label ID:
			id := c.imgData[y][x]

			// Connected component ID:
			m := labels[id]
			m[v] = struct{}{}
			labels[id] = m
		}
	}

	// Iterate over the full list of possible labels so we can be sure that each
	// one has a representation in our output
	out := make(map[Label]int)
	for _, label := range l.Sorted() {
		out[label] = len(labels[uint8(label.ID)])
	}

	return out, nil
}

func (c Connected) LabelAbove(x, y int) (bool, uint32) {
	if y == 0 {
		return false, 0
	}

	if c.imgData[y][x] != c.imgData[y-1][x] {
		return false, 0
	}

	return true, c.labels[y-1][x]
}

func (c Connected) LabelLeft(x, y int) (bool, uint32) {
	if x == 0 {
		return false, 0
	}

	if c.imgData[y][x] != c.imgData[y][x-1] {
		return false, 0
	}

	return true, c.labels[y][x-1]
}
