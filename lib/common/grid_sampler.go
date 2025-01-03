package common

import (
	"github.com/ozgur-yalcin/mfa/lib"
)

type GridSampler interface {
	SampleGrid(image *lib.BitMatrix, dimensionX, dimensionY int,
		p1ToX, p1ToY, p2ToX, p2ToY, p3ToX, p3ToY, p4ToX, p4ToY float64,
		p1FromX, p1FromY, p2FromX, p2FromY, p3FromX, p3FromY, p4FromX, p4FromY float64) (*lib.BitMatrix, error)

	SampleGridWithTransform(image *lib.BitMatrix,
		dimensionX, dimensionY int, transform *PerspectiveTransform) (*lib.BitMatrix, error)
}

var gridSampler GridSampler = NewDefaultGridSampler()

func GridSampler_SetGridSampler(newGridSampler GridSampler) {
	gridSampler = newGridSampler
}

func GridSampler_GetInstance() GridSampler {
	return gridSampler
}

func GridSampler_checkAndNudgePoints(image *lib.BitMatrix, points []float64) error {
	width := image.GetWidth()
	height := image.GetHeight()
	nudged := true
	maxOffset := len(points) - 1
	for offset := 0; offset < maxOffset && nudged; offset += 2 {
		x := int(points[offset])
		y := int(points[offset+1])
		if x < -1 || x > width || y < -1 || y > height {
			return lib.NewNotFoundException(
				"(w, h) = (%v, %v),  (x, y) = (%v, %v)", width, height, x, y)
		}
		nudged = false
		if x == -1 {
			points[offset] = 0.0
			nudged = true
		} else if x == width {
			points[offset] = float64(width - 1)
			nudged = true
		}
		if y == -1 {
			points[offset+1] = 0.0
			nudged = true
		} else if y == height {
			points[offset+1] = float64(height)
			nudged = true
		}
	}
	nudged = true
	for offset := len(points) - 2; offset >= 0 && nudged; offset -= 2 {
		x := int(points[offset])
		y := int(points[offset+1])
		if x < -1 || x > width || y < -1 || y > height {
			return lib.NewNotFoundException(
				"(w, h) = (%v, %v),  (x, y) = (%v, %v)", width, height, x, y)
		}
		nudged = false
		if x == -1 {
			points[offset] = 0.0
			nudged = true
		} else if x == width {
			points[offset] = float64(width - 1)
			nudged = true
		}
		if y == -1 {
			points[offset+1] = 0.0
			nudged = true
		} else if y == height {
			points[offset+1] = float64(height - 1)
			nudged = true
		}
	}
	return nil
}
