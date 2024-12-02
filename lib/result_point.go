package lib

import (
	"github.com/ozgur-yalcin/mfa/lib/common/util"
)

type ResultPoint interface {
	GetX() float64
	GetY() float64
}

type ResultPointBase struct {
	x float64
	y float64
}

func NewResultPoint(x, y float64) ResultPoint {
	return ResultPointBase{x, y}
}

func (rp ResultPointBase) GetX() float64 {
	return rp.x
}

func (rp ResultPointBase) GetY() float64 {
	return rp.y
}

func ResultPoint_OrderBestPatterns(pattern0, pattern1, pattern2 ResultPoint) (pointA, pointB, pointC ResultPoint) {
	zeroOneDistance := ResultPoint_Distance(pattern0, pattern1)
	oneTwoDistance := ResultPoint_Distance(pattern1, pattern2)
	zeroTwoDistance := ResultPoint_Distance(pattern0, pattern2)

	if oneTwoDistance >= zeroOneDistance && oneTwoDistance >= zeroTwoDistance {
		pointB = pattern0
		pointA = pattern1
		pointC = pattern2
	} else if zeroTwoDistance >= oneTwoDistance && zeroTwoDistance >= zeroOneDistance {
		pointB = pattern1
		pointA = pattern0
		pointC = pattern2
	} else {
		pointB = pattern2
		pointA = pattern0
		pointC = pattern1
	}

	if crossProductZ(pointA, pointB, pointC) < 0.0 {
		pointA, pointC = pointC, pointA
	}

	return pointA, pointB, pointC
}

func ResultPoint_Distance(pattern1, pattern2 ResultPoint) float64 {
	return util.MathUtils_DistanceFloat(pattern1.GetX(), pattern1.GetY(), pattern2.GetX(), pattern2.GetY())
}

func crossProductZ(pointA, pointB, pointC ResultPoint) float64 {
	bX := pointB.GetX()
	bY := pointB.GetY()
	return ((pointC.GetX() - bX) * (pointA.GetY() - bY)) - ((pointC.GetY() - bY) * (pointA.GetX() - bX))
}
