package overlay

import (
	"image"

	"github.com/theodesp/unionfind"
)

// Following the guide at
// http://aishack.in/tutorials/connected-component-labelling/

type Connected struct {
	// Each point represents a label ID
	PixelLabelIDs [][]uint8

	// Each point represents a connected component ID
	PixelConnectedComponentIDs [][]uint32

	// Each component of each label is tracked, including bounding box and pixel count
	LabeledConnectedComponents map[uint8]map[uint32]ConnectedComponent
}

func NewConnected(img image.Image) (*Connected, error) {
	out := &Connected{}

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

	out.PixelLabelIDs = imgData
	out.PixelConnectedComponentIDs = labels

	return out, nil
}

// Count evaluates the number of pixels for each label, the number of components
// for each label, and then also performs the same operations on a subset of the
// data that requires that a threshold be met (in terms of number of pixels
// within a contiguous component) in order to be counted. This permits you to,
// e.g., ignore single-pixel noisy blips that appear and which shouldn't be
// counted towards area measurements.
func (c *Connected) Count(l LabelMap, threshold int) (rawPixels, rawCounts, thresholdedPixels, thresholdedCounts map[Label]int, err error) {
	uf := unionfind.NewThreadSafeUnionFind(25000)

	var nextLabel uint32 = 1
	for y, row := range c.PixelLabelIDs {
		for x := range row {

			// See if we already labeled an adjacent pixel with a connected component ID.
			found := false
			foundUp, valUp := c.LabelAbove(x, y)
			if foundUp {
				found = true
				c.PixelConnectedComponentIDs[y][x] = valUp
			}

			foundLeft, val := c.LabelLeft(x, y)
			if foundLeft {
				found = true
				// This overrides the pixel if it was set by foundUp, but since
				// we will find their shared root in the second pass, there is
				// no reason to check or abort this override step.
				c.PixelConnectedComponentIDs[y][x] = val

				if foundUp {
					// If the left pixel's label differs from that of the top
					// pixel, now we know these two labels need to be joined.
					// Keeping the lower value here seems to be recommended but,
					// as far as I can tell, is irrelevant because we check all
					// roots in the second pass.
					uf.Union(int(val), int(valUp))
				}
			}

			if found {
				continue
			}

			// If neither neighbor is a candidate root label, then this pixel
			// gets its own label.
			c.PixelConnectedComponentIDs[y][x] = nextLabel
			nextLabel++
		}
	}

	// Now reconcile the adjacent connected component labels
	for y, row := range c.PixelConnectedComponentIDs {
		for x, v := range row {

			root := uf.Root(int(v))
			if root < 0 {
				// No adjacent connected component labels
				continue
			}

			// If a root exists, replace this connected component label with the
			// root connected component label. The root isn't guaranteed to be
			// numerically the smallest entry in its tree, so no value checks
			// are done here.
			c.PixelConnectedComponentIDs[y][x] = uint32(root)
		}
	}

	// Count number of components per labelID

	// For each truth label, give it a map that lists all of the connected
	// component labels that are assigned to it. Each separate connected
	// component label will also get a pixel count.
	c.LabeledConnectedComponents = make(map[uint8]map[uint32]ConnectedComponent) //map[labelID] => map[(label, connected component)]ConnectedComponent
	for _, v := range l.Sorted() {
		c.LabeledConnectedComponents[uint8(v.ID)] = make(map[uint32]ConnectedComponent)
	}

	// Each pixel is now marked with its corresponding [label, connected]
	// component tuple.
	for y, row := range c.PixelConnectedComponentIDs {
		for x, connectedComponentID := range row {
			// Original label ID:
			id := c.PixelLabelIDs[y][x]

			// Connected component ID:
			m, exists := c.LabeledConnectedComponents[id]
			if !exists {
				m = make(map[uint32]ConnectedComponent)
				c.LabeledConnectedComponents[id] = m
			}

			component, exists := m[connectedComponentID]

			// Update our bounding box for this component
			if !exists {
				// This is the first we have seen of the component
				component.LabelID = id
				component.ComponentID = connectedComponentID
				component.Bounds.TopLeft = Coord{X: x, Y: y}
				component.Bounds.BottomRight = Coord{X: x, Y: y}
			} else {
				// Since we are scanning y = 0 on up, no need to check for
				// TopLeft Y
				if x < component.Bounds.TopLeft.X {
					component.Bounds.TopLeft.X = x
				}
				if x > component.Bounds.BottomRight.X {
					component.Bounds.BottomRight.X = x
				}
				if y > component.Bounds.BottomRight.Y {
					component.Bounds.BottomRight.Y = y
				}
			}

			component.PixelCount++
			m[connectedComponentID] = component
			c.LabeledConnectedComponents[id] = m
			// m[v]++
			// labels[id] = m
		}
	}

	// Iterate over the full list of possible labels so we can be sure that each
	// one has a representation in our output
	rawPixels = make(map[Label]int)
	rawCounts = make(map[Label]int)
	thresholdedPixels = make(map[Label]int)
	thresholdedCounts = make(map[Label]int)
	for _, label := range l.Sorted() {

		subcomponents := c.LabeledConnectedComponents[uint8(label.ID)]

		// The main connected components output just tallies the number of
		// components seen for each labelID, regardless of how many pixels they
		// encompass.
		rawCounts[label] = len(subcomponents)

		// For pixel counts and the thresholded component count, only increment if the
		// number of pixels of a subcomponent is greater than the defined
		// threshold. The idea is that blips of noise can be ignored.
		for _, subcount := range subcomponents {

			// Always increment the raw pixel count
			rawPixels[label] += subcount.PixelCount

			// Only increment the thresholded pixel count and thresholded
			// component count if this component surpasses the threshold
			if subcount.PixelCount < threshold {
				continue
			}

			thresholdedPixels[label] += subcount.PixelCount

			thresholdedCounts[label]++
		}
	}

	return rawPixels, rawCounts, thresholdedPixels, thresholdedCounts, nil
}

func (c Connected) LabelAbove(x, y int) (bool, uint32) {
	if y == 0 {
		return false, 0
	}

	if c.PixelLabelIDs[y][x] != c.PixelLabelIDs[y-1][x] {
		return false, 0
	}

	return true, c.PixelConnectedComponentIDs[y-1][x]
}

func (c Connected) LabelLeft(x, y int) (bool, uint32) {
	if x == 0 {
		return false, 0
	}

	if c.PixelLabelIDs[y][x] != c.PixelLabelIDs[y][x-1] {
		return false, 0
	}

	return true, c.PixelConnectedComponentIDs[y][x-1]
}
