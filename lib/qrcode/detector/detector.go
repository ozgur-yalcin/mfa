package detector

import (
	"math"

	"github.com/ozgur-yalcin/mfa/lib"
	"github.com/ozgur-yalcin/mfa/lib/common"
	"github.com/ozgur-yalcin/mfa/lib/common/util"
	"github.com/ozgur-yalcin/mfa/lib/qrcode/decoder"
)

type Detector struct {
	image               *lib.BitMatrix
	resultPointCallback lib.ResultPointCallback
}

func NewDetector(image *lib.BitMatrix) *Detector {
	return &Detector{image, nil}
}

func (this *Detector) GetImage() *lib.BitMatrix {
	return this.image
}

func (this *Detector) GetResultPointCallback() lib.ResultPointCallback {
	return this.resultPointCallback
}

func (this *Detector) DetectWithoutHints() (*common.DetectorResult, error) {
	return this.Detect(nil)
}

func (this *Detector) Detect(hints map[lib.DecodeHintType]interface{}) (*common.DetectorResult, error) {
	if hints != nil {
		if cb, ok := hints[lib.DecodeHintType_NEED_RESULT_POINT_CALLBACK]; ok {
			this.resultPointCallback, _ = cb.(lib.ResultPointCallback)
		}
	}

	finder := NewFinderPatternFinder(this.image, this.resultPointCallback)
	info, e := finder.Find(hints)
	if e != nil {
		return nil, e
	}

	return this.ProcessFinderPatternInfo(info)
}

func (this *Detector) ProcessFinderPatternInfo(info *FinderPatternInfo) (*common.DetectorResult, error) {
	topLeft := info.GetTopLeft()
	topRight := info.GetTopRight()
	bottomLeft := info.GetBottomLeft()

	moduleSize := this.calculateModuleSize(topLeft, topRight, bottomLeft)
	if moduleSize < 1.0 {
		return nil, lib.NewNotFoundException("moduleSize = %v", moduleSize)
	}
	dimension, e := this.computeDimension(topLeft, topRight, bottomLeft, moduleSize)
	if e != nil {
		return nil, e
	}
	provisionalVersion, e := decoder.Version_GetProvisionalVersionForDimension(dimension)
	if e != nil {
		return nil, lib.WrapFormatException(e)
	}
	modulesBetweenFPCenters := provisionalVersion.GetDimensionForVersion() - 7

	var alignmentPattern *AlignmentPattern
	if len(provisionalVersion.GetAlignmentPatternCenters()) > 0 {
		bottomRightX := topRight.GetX() - topLeft.GetX() + bottomLeft.GetX()
		bottomRightY := topRight.GetY() - topLeft.GetY() + bottomLeft.GetY()

		correctionToTopLeft := 1.0 - 3.0/float64(modulesBetweenFPCenters)
		estAlignmentX := int(topLeft.GetX() + correctionToTopLeft*(bottomRightX-topLeft.GetX()))
		estAlignmentY := int(topLeft.GetY() + correctionToTopLeft*(bottomRightY-topLeft.GetY()))

		for i := 4; i <= 16; i <<= 1 {
			alignmentPattern, e = this.findAlignmentInRegion(moduleSize,
				estAlignmentX,
				estAlignmentY,
				float64(i))
			if e == nil {
				break
			} else if _, ok := e.(lib.NotFoundException); !ok {
				return nil, e
			}
		}
	}

	transform := Detector_createTransform(topLeft, topRight, bottomLeft, alignmentPattern, dimension)

	bits, e := Detector_sampleGrid(this.image, transform, dimension)
	if e != nil {
		return nil, lib.WrapNotFoundException(e)
	}

	var points []lib.ResultPoint
	if alignmentPattern == nil {
		points = []lib.ResultPoint{bottomLeft, topLeft, topRight}
	} else {
		points = []lib.ResultPoint{bottomLeft, topLeft, topRight, alignmentPattern}
	}
	return common.NewDetectorResult(bits, points), nil
}

func Detector_createTransform(topLeft, topRight, bottomLeft lib.ResultPoint, alignmentPattern *AlignmentPattern, dimension int) *common.PerspectiveTransform {
	dimMinusThree := float64(dimension) - 3.5
	var bottomRightX float64
	var bottomRightY float64
	var sourceBottomRightX float64
	var sourceBottomRightY float64
	if alignmentPattern != nil {
		bottomRightX = alignmentPattern.GetX()
		bottomRightY = alignmentPattern.GetY()
		sourceBottomRightX = dimMinusThree - 3.0
		sourceBottomRightY = sourceBottomRightX
	} else {
		bottomRightX = (topRight.GetX() - topLeft.GetX()) + bottomLeft.GetX()
		bottomRightY = (topRight.GetY() - topLeft.GetY()) + bottomLeft.GetY()
		sourceBottomRightX = dimMinusThree
		sourceBottomRightY = dimMinusThree
	}

	return common.PerspectiveTransform_QuadrilateralToQuadrilateral(
		3.5,
		3.5,
		dimMinusThree,
		3.5,
		sourceBottomRightX,
		sourceBottomRightY,
		3.5,
		dimMinusThree,
		topLeft.GetX(),
		topLeft.GetY(),
		topRight.GetX(),
		topRight.GetY(),
		bottomRightX,
		bottomRightY,
		bottomLeft.GetX(),
		bottomLeft.GetY())
}

func Detector_sampleGrid(image *lib.BitMatrix, transform *common.PerspectiveTransform, dimension int) (*lib.BitMatrix, error) {
	sampler := common.GridSampler_GetInstance()
	return sampler.SampleGridWithTransform(image, dimension, dimension, transform)
}

func (this *Detector) computeDimension(topLeft, topRight, bottomLeft lib.ResultPoint, moduleSize float64) (int, error) {
	tltrCentersDimension := util.MathUtils_Round(lib.ResultPoint_Distance(topLeft, topRight) / moduleSize)
	tlblCentersDimension := util.MathUtils_Round(lib.ResultPoint_Distance(topLeft, bottomLeft) / moduleSize)
	dimension := ((tltrCentersDimension + tlblCentersDimension) / 2) + 7
	switch dimension % 4 {
	default:
		break
	case 0:
		dimension++
		break
	case 2:
		dimension--
		break
	case 3:
		return 0, lib.NewNotFoundException("dimension = %v", dimension)
	}
	return dimension, nil
}

func (this *Detector) calculateModuleSize(topLeft, topRight, bottomLeft lib.ResultPoint) float64 {
	return (this.calculateModuleSizeOneWay(topLeft, topRight) +
		this.calculateModuleSizeOneWay(topLeft, bottomLeft)) / 2
}

func (this *Detector) calculateModuleSizeOneWay(pattern, otherPattern lib.ResultPoint) float64 {
	moduleSizeEst1 := this.sizeOfBlackWhiteBlackRunBothWays(int(pattern.GetX()),
		int(pattern.GetY()),
		int(otherPattern.GetX()),
		int(otherPattern.GetY()))
	moduleSizeEst2 := this.sizeOfBlackWhiteBlackRunBothWays(int(otherPattern.GetX()),
		int(otherPattern.GetY()),
		int(pattern.GetX()),
		int(pattern.GetY()))
	if math.IsNaN(moduleSizeEst1) {
		return moduleSizeEst2 / 7.0
	}
	if math.IsNaN(moduleSizeEst2) {
		return moduleSizeEst1 / 7.0
	}
	return (moduleSizeEst1 + moduleSizeEst2) / 14.0
}

func (this *Detector) sizeOfBlackWhiteBlackRunBothWays(fromX, fromY, toX, toY int) float64 {

	result := this.sizeOfBlackWhiteBlackRun(fromX, fromY, toX, toY)

	scale := float64(1.0)
	otherToX := fromX - (toX - fromX)
	if otherToX < 0 {
		scale = float64(fromX) / float64(fromX-otherToX)
		otherToX = 0
	} else if otherToX >= this.image.GetWidth() {
		scale = float64(this.image.GetWidth()-1-fromX) / float64(otherToX-fromX)
		otherToX = this.image.GetWidth() - 1
	}
	otherToY := int(float64(fromY) - float64(toY-fromY)*scale)

	scale = 1.0
	if otherToY < 0 {
		scale = float64(fromY) / float64(fromY-otherToY)
		otherToY = 0
	} else if otherToY >= this.image.GetHeight() {
		scale = float64(this.image.GetHeight()-1-fromY) / float64(otherToY-fromY)
		otherToY = this.image.GetHeight() - 1
	}
	otherToX = int(float64(fromX) + float64(otherToX-fromX)*scale)

	result += this.sizeOfBlackWhiteBlackRun(fromX, fromY, otherToX, otherToY)

	return result - 1.0
}

func (this *Detector) sizeOfBlackWhiteBlackRun(fromX, fromY, toX, toY int) float64 {
	steep := false
	dx := toX - fromX
	if dx < 0 {
		dx = -dx
	}
	dy := toY - fromY
	if dy < 0 {
		dy = -dy
	}
	if dy > dx {
		steep = true
		fromX, fromY = fromY, fromX
		toX, toY = toY, toX
		dx, dy = dy, dx
	}

	error := -dx / 2
	xstep := 1
	if fromX >= toX {
		xstep = -1
	}
	ystep := 1
	if fromY >= toY {
		ystep = -1
	}
	state := 0
	xLimit := toX + xstep
	for x, y := fromX, fromY; x != xLimit; x += xstep {
		realX := x
		realY := y
		if steep {
			realX = y
			realY = x
		}

		if (state == 1) == this.image.Get(realX, realY) {
			if state == 2 {
				return util.MathUtils_DistanceInt(x, y, fromX, fromY)
			}
			state++
		}

		error += dy
		if error > 0 {
			if y == toY {
				break
			}
			y += ystep
			error -= dx
		}
	}
	if state == 2 {
		return util.MathUtils_DistanceInt(toX+xstep, toY, fromX, fromY)
	}
	return math.NaN()
}

func (this *Detector) findAlignmentInRegion(overallEstModuleSize float64, estAlignmentX, estAlignmentY int, allowanceFactor float64) (*AlignmentPattern, error) {
	allowance := int(allowanceFactor * overallEstModuleSize)
	alignmentAreaLeftX := estAlignmentX - allowance
	if alignmentAreaLeftX < 0 {
		alignmentAreaLeftX = 0
	}
	alignmentAreaRightX := estAlignmentX + allowance
	if a := this.image.GetWidth() - 1; a < alignmentAreaRightX {
		alignmentAreaRightX = a
	}
	if x := float64(alignmentAreaRightX - alignmentAreaLeftX); x < overallEstModuleSize*3 {
		return nil, lib.NewNotFoundException("x = %v, moduleSize = %v", x, overallEstModuleSize)
	}

	alignmentAreaTopY := estAlignmentY - allowance
	if alignmentAreaTopY < 0 {
		alignmentAreaTopY = 0
	}
	alignmentAreaBottomY := estAlignmentY + allowance
	if a := this.image.GetHeight() - 1; a < alignmentAreaBottomY {
		alignmentAreaBottomY = a
	}

	if y := float64(alignmentAreaBottomY - alignmentAreaTopY); y < overallEstModuleSize*3 {
		return nil, lib.NewNotFoundException("y = %v, moduleSize = %v", y, overallEstModuleSize)
	}

	alignmentFinder := NewAlignmentPatternFinder(
		this.image,
		alignmentAreaLeftX,
		alignmentAreaTopY,
		alignmentAreaRightX-alignmentAreaLeftX,
		alignmentAreaBottomY-alignmentAreaTopY,
		overallEstModuleSize,
		this.resultPointCallback)
	return alignmentFinder.Find()
}
