package common

import (
	"github.com/ozgur-yalcin/mfa/scan"
)

type DetectorResult struct {
	bits   *scan.BitMatrix
	points []scan.ResultPoint
}

func NewDetectorResult(bits *scan.BitMatrix, points []scan.ResultPoint) *DetectorResult {
	return &DetectorResult{bits, points}
}

func (d *DetectorResult) GetBits() *scan.BitMatrix {
	return d.bits
}

func (d *DetectorResult) GetPoints() []scan.ResultPoint {
	return d.points
}
