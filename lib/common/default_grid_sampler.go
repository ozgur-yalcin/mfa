package common

import (
	"github.com/ozgur-yalcin/mfa/lib"
)

type DefaultGridSampler struct{}

func NewDefaultGridSampler() GridSampler {
	return DefaultGridSampler{}
}

func (s DefaultGridSampler) SampleGrid(image *lib.BitMatrix, dimensionX, dimensionY int,
	p1ToX, p1ToY, p2ToX, p2ToY, p3ToX, p3ToY, p4ToX, p4ToY float64,
	p1FromX, p1FromY, p2FromX, p2FromY, p3FromX, p3FromY, p4FromX, p4FromY float64) (*lib.BitMatrix, error) {

	transform := PerspectiveTransform_QuadrilateralToQuadrilateral(
		p1ToX, p1ToY, p2ToX, p2ToY, p3ToX, p3ToY, p4ToX, p4ToY,
		p1FromX, p1FromY, p2FromX, p2FromY, p3FromX, p3FromY, p4FromX, p4FromY)

	return s.SampleGridWithTransform(image, dimensionX, dimensionY, transform)
}

func (s DefaultGridSampler) SampleGridWithTransform(image *lib.BitMatrix,
	dimensionX, dimensionY int, transform *PerspectiveTransform) (*lib.BitMatrix, error) {

	if dimensionX <= 0 || dimensionY <= 0 {
		return nil, lib.NewNotFoundException("dimensions X, Y = %v, %v", dimensionX, dimensionY)
	}
	bits, _ := lib.NewBitMatrix(dimensionX, dimensionY)
	points := make([]float64, 2*dimensionX)
	for y := 0; y < dimensionY; y++ {
		max := len(points)
		iValue := float64(y) + 0.5
		for x := 0; x < max; x += 2 {
			points[x] = float64(x/2) + 0.5
			points[x+1] = iValue
		}
		transform.TransformPoints(points)
		e := GridSampler_checkAndNudgePoints(image, points)
		if e != nil {
			return nil, lib.WrapNotFoundException(e)
		}
		for x := 0; x < max; x += 2 {
			px := int(points[x])
			py := int(points[x+1])

			if px >= image.GetWidth() || py >= image.GetHeight() {
				return nil, lib.NewNotFoundException()
			}

			if image.Get(px, py) {
				bits.Set(x/2, y)
			}
		}
	}
	return bits, nil
}
