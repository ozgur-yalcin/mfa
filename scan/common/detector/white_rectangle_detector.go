package detector

import (
	"github.com/ozgur-yalcin/mfa/scan"
	"github.com/ozgur-yalcin/mfa/scan/common/util"
)

const (
	whiteRectangleDetector_INIT_SIZE = 10
	whiteRectangleDetector_CORR      = 1
)

type WhiteRectangleDetector struct {
	image     *scan.BitMatrix
	height    int
	width     int
	leftInit  int
	rightInit int
	downInit  int
	upInit    int
}

func NewWhiteRectangleDetectorFromImage(image *scan.BitMatrix) (*WhiteRectangleDetector, error) {
	return NewWhiteRectangleDetector(
		image, whiteRectangleDetector_INIT_SIZE, image.GetWidth()/2, image.GetHeight()/2)
}

func NewWhiteRectangleDetector(image *scan.BitMatrix, initSize, x, y int) (*WhiteRectangleDetector, error) {
	halfsize := initSize / 2
	d := &WhiteRectangleDetector{
		image:     image,
		height:    image.GetHeight(),
		width:     image.GetWidth(),
		leftInit:  x - halfsize,
		rightInit: x + halfsize,
		upInit:    y - halfsize,
		downInit:  y + halfsize,
	}
	if d.upInit < 0 || d.leftInit < 0 || d.downInit >= d.height || d.rightInit >= d.width {
		return nil, scan.NewNotFoundException()
	}
	return d, nil
}

func (this *WhiteRectangleDetector) Detect() ([]scan.ResultPoint, error) {
	left := this.leftInit
	right := this.rightInit
	up := this.upInit
	down := this.downInit
	sizeExceeded := false
	aBlackPointFoundOnBorder := true

	atLeastOneBlackPointFoundOnRight := false
	atLeastOneBlackPointFoundOnBottom := false
	atLeastOneBlackPointFoundOnLeft := false
	atLeastOneBlackPointFoundOnTop := false

	for aBlackPointFoundOnBorder {

		aBlackPointFoundOnBorder = false

		rightBorderNotWhite := true
		for (rightBorderNotWhite || !atLeastOneBlackPointFoundOnRight) && right < this.width {
			rightBorderNotWhite = this.containsBlackPoint(up, down, right, false)
			if rightBorderNotWhite {
				right++
				aBlackPointFoundOnBorder = true
				atLeastOneBlackPointFoundOnRight = true
			} else if !atLeastOneBlackPointFoundOnRight {
				right++
			}
		}

		if right >= this.width {
			sizeExceeded = true
			break
		}

		bottomBorderNotWhite := true
		for (bottomBorderNotWhite || !atLeastOneBlackPointFoundOnBottom) && down < this.height {
			bottomBorderNotWhite = this.containsBlackPoint(left, right, down, true)
			if bottomBorderNotWhite {
				down++
				aBlackPointFoundOnBorder = true
				atLeastOneBlackPointFoundOnBottom = true
			} else if !atLeastOneBlackPointFoundOnBottom {
				down++
			}
		}

		if down >= this.height {
			sizeExceeded = true
			break
		}

		leftBorderNotWhite := true
		for (leftBorderNotWhite || !atLeastOneBlackPointFoundOnLeft) && left >= 0 {
			leftBorderNotWhite = this.containsBlackPoint(up, down, left, false)
			if leftBorderNotWhite {
				left--
				aBlackPointFoundOnBorder = true
				atLeastOneBlackPointFoundOnLeft = true
			} else if !atLeastOneBlackPointFoundOnLeft {
				left--
			}
		}

		if left < 0 {
			sizeExceeded = true
			break
		}

		topBorderNotWhite := true
		for (topBorderNotWhite || !atLeastOneBlackPointFoundOnTop) && up >= 0 {
			topBorderNotWhite = this.containsBlackPoint(left, right, up, true)
			if topBorderNotWhite {
				up--
				aBlackPointFoundOnBorder = true
				atLeastOneBlackPointFoundOnTop = true
			} else if !atLeastOneBlackPointFoundOnTop {
				up--
			}
		}

		if up < 0 {
			sizeExceeded = true
			break
		}

	}

	if !sizeExceeded {

		maxSize := right - left

		var z scan.ResultPoint
		for i := 1; z == nil && i < maxSize; i++ {
			z = this.getBlackPointOnSegment(left, down-i, left+i, down)
		}

		if z == nil {
			return nil, scan.NewNotFoundException("no black point on left-down")
		}

		var t scan.ResultPoint
		//go down right
		for i := 1; t == nil && i < maxSize; i++ {
			t = this.getBlackPointOnSegment(left, up+i, left+i, up)
		}

		if t == nil {
			return nil, scan.NewNotFoundException("no black point on left-up")
		}

		var x scan.ResultPoint
		//go down left
		for i := 1; x == nil && i < maxSize; i++ {
			x = this.getBlackPointOnSegment(right, up+i, right-i, up)
		}

		if x == nil {
			return nil, scan.NewNotFoundException("no black point on right-up")
		}

		var y scan.ResultPoint
		//go up left
		for i := 1; y == nil && i < maxSize; i++ {
			y = this.getBlackPointOnSegment(right, down-i, right-i, down)
		}

		if y == nil {
			return nil, scan.NewNotFoundException("no black point on right-down")
		}

		return this.centerEdges(y, z, x, t), nil
	}

	return nil, scan.NewNotFoundException()
}

func (this *WhiteRectangleDetector) getBlackPointOnSegment(aX, aY, bX, bY int) scan.ResultPoint {
	dist := util.MathUtils_Round(util.MathUtils_DistanceInt(aX, aY, bX, bY))
	xStep := float64(bX-aX) / float64(dist)
	yStep := float64(bY-aY) / float64(dist)

	for i := 0; i < dist; i++ {
		x := util.MathUtils_Round(float64(aX) + float64(i)*xStep)
		y := util.MathUtils_Round(float64(aY) + float64(i)*yStep)
		if this.image.Get(x, y) {
			return scan.NewResultPoint(float64(x), float64(y))
		}
	}
	return nil
}

func (this *WhiteRectangleDetector) centerEdges(y, z, x, t scan.ResultPoint) []scan.ResultPoint {
	yi := y.GetX()
	yj := y.GetY()
	zi := z.GetX()
	zj := z.GetY()
	xi := x.GetX()
	xj := x.GetY()
	ti := t.GetX()
	tj := t.GetY()

	if yi < float64(this.width)/2.0 {
		return []scan.ResultPoint{
			scan.NewResultPoint(ti-whiteRectangleDetector_CORR, tj+whiteRectangleDetector_CORR),
			scan.NewResultPoint(zi+whiteRectangleDetector_CORR, zj+whiteRectangleDetector_CORR),
			scan.NewResultPoint(xi-whiteRectangleDetector_CORR, xj-whiteRectangleDetector_CORR),
			scan.NewResultPoint(yi+whiteRectangleDetector_CORR, yj-whiteRectangleDetector_CORR),
		}
	} else {
		return []scan.ResultPoint{
			scan.NewResultPoint(ti+whiteRectangleDetector_CORR, tj+whiteRectangleDetector_CORR),
			scan.NewResultPoint(zi+whiteRectangleDetector_CORR, zj-whiteRectangleDetector_CORR),
			scan.NewResultPoint(xi-whiteRectangleDetector_CORR, xj+whiteRectangleDetector_CORR),
			scan.NewResultPoint(yi-whiteRectangleDetector_CORR, yj-whiteRectangleDetector_CORR),
		}
	}
}

func (this *WhiteRectangleDetector) containsBlackPoint(a, b, fixed int, horizontal bool) bool {

	if horizontal {
		for x := a; x <= b; x++ {
			if this.image.Get(x, fixed) {
				return true
			}
		}
	} else {
		for y := a; y <= b; y++ {
			if this.image.Get(fixed, y) {
				return true
			}
		}
	}

	return false
}
