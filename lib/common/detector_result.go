package common

import "github.com/ozgur-yalcin/mfa/lib"

type DetectorResult struct {
	bits   *lib.BitMatrix
	points []lib.ResultPoint
}

func NewDetectorResult(bits *lib.BitMatrix, points []lib.ResultPoint) *DetectorResult {
	return &DetectorResult{bits, points}
}

func (d *DetectorResult) GetBits() *lib.BitMatrix {
	return d.bits
}

func (d *DetectorResult) GetPoints() []lib.ResultPoint {
	return d.points
}
