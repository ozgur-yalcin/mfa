package detector

import (
	"math"

	"github.com/ozgur-yalcin/mfa/lib"
)

type AlignmentPatternFinder struct {
	image                *lib.BitMatrix
	possibleCenters      []*AlignmentPattern
	startX               int
	startY               int
	width                int
	height               int
	moduleSize           float64
	crossCheckStateCount []int
	resultPointCallback  lib.ResultPointCallback
}

func NewAlignmentPatternFinder(image *lib.BitMatrix, startX, startY, width, height int, moduleSize float64, resultPointCallback lib.ResultPointCallback) *AlignmentPatternFinder {
	return &AlignmentPatternFinder{
		image:                image,
		possibleCenters:      make([]*AlignmentPattern, 0),
		startX:               startX,
		startY:               startY,
		width:                width,
		height:               height,
		moduleSize:           moduleSize,
		crossCheckStateCount: make([]int, 3),
		resultPointCallback:  resultPointCallback,
	}
}

func (this *AlignmentPatternFinder) Find() (*AlignmentPattern, lib.NotFoundException) {
	startX := this.startX
	height := this.height
	maxJ := startX + this.width
	middleI := this.startY + (this.height / 2)
	stateCount := make([]int, 3)
	for iGen := 0; iGen < height; iGen++ {
		i := middleI
		if iGen&1 == 0 {
			i += (iGen + 1) / 2
		} else {
			i -= (iGen + 1) / 2
		}
		stateCount[0] = 0
		stateCount[1] = 0
		stateCount[2] = 0
		j := startX
		for j < maxJ && !this.image.Get(j, i) {
			j++
		}
		currentState := 0
		for j < maxJ {
			if this.image.Get(j, i) {
				if currentState == 1 {
					stateCount[1]++
				} else {
					if currentState == 2 {
						if this.foundPatternCross(stateCount) {
							confirmed := this.handlePossibleCenter(stateCount, i, j)
							if confirmed != nil {
								return confirmed, nil
							}
						}
						stateCount[0] = stateCount[2]
						stateCount[1] = 1
						stateCount[2] = 0
						currentState = 1
					} else {
						currentState++
						stateCount[currentState]++
					}
				}
			} else {
				if currentState == 1 {
					currentState++
				}
				stateCount[currentState]++
			}
			j++
		}
		if this.foundPatternCross(stateCount) {
			confirmed := this.handlePossibleCenter(stateCount, i, maxJ)
			if confirmed != nil {
				return confirmed, nil
			}
		}

	}

	if len(this.possibleCenters) > 0 {
		return this.possibleCenters[0], nil
	}

	return nil, lib.NewNotFoundException()
}

func AlignmentPatternFinder_centerFromEnd(stateCount []int, end int) float64 {
	return float64(end-stateCount[2]) - float64(stateCount[1])/2.0
}

func (this *AlignmentPatternFinder) foundPatternCross(stateCount []int) bool {
	moduleSize := this.moduleSize
	maxVariance := moduleSize / 2
	for i := 0; i < 3; i++ {
		if math.Abs(moduleSize-float64(stateCount[i])) >= maxVariance {
			return false
		}
	}
	return true
}

func (this *AlignmentPatternFinder) crossCheckVertical(startI, centerJ, maxCount, originalStateCountTotal int) float64 {
	image := this.image

	maxI := image.GetHeight()
	stateCount := this.crossCheckStateCount
	stateCount[0] = 0
	stateCount[1] = 0
	stateCount[2] = 0

	i := startI
	for i >= 0 && image.Get(centerJ, i) && stateCount[1] <= maxCount {
		stateCount[1]++
		i--
	}
	if i < 0 || stateCount[1] > maxCount {
		return math.NaN()
	}
	for i >= 0 && !image.Get(centerJ, i) && stateCount[0] <= maxCount {
		stateCount[0]++
		i--
	}
	if stateCount[0] > maxCount {
		return math.NaN()
	}

	i = startI + 1
	for i < maxI && image.Get(centerJ, i) && stateCount[1] <= maxCount {
		stateCount[1]++
		i++
	}
	if i == maxI || stateCount[1] > maxCount {
		return math.NaN()
	}
	for i < maxI && !image.Get(centerJ, i) && stateCount[2] <= maxCount {
		stateCount[2]++
		i++
	}
	if stateCount[2] > maxCount {
		return math.NaN()
	}

	stateCountTotal := stateCount[0] + stateCount[1] + stateCount[2]
	abs := stateCountTotal - originalStateCountTotal
	if abs < 0 {
		abs = -abs
	}
	if 5*abs >= 2*originalStateCountTotal {
		return math.NaN()
	}

	if this.foundPatternCross(stateCount) {
		return AlignmentPatternFinder_centerFromEnd(stateCount, i)
	}
	return math.NaN()
}

func (this *AlignmentPatternFinder) handlePossibleCenter(stateCount []int, i, j int) *AlignmentPattern {
	stateCountTotal := stateCount[0] + stateCount[1] + stateCount[2]
	centerJ := AlignmentPatternFinder_centerFromEnd(stateCount, j)
	centerI := this.crossCheckVertical(i, int(centerJ), 2*stateCount[1], stateCountTotal)
	if !math.IsNaN(centerI) {
		estimatedModuleSize := float64(stateCount[0]+stateCount[1]+stateCount[2]) / 3
		for _, center := range this.possibleCenters {
			if center.AboutEquals(estimatedModuleSize, centerI, centerJ) {
				return center.CombineEstimate(centerI, centerJ, estimatedModuleSize)
			}
		}
		point := NewAlignmentPattern(centerJ, centerI, estimatedModuleSize)
		this.possibleCenters = append(this.possibleCenters, point)
		if this.resultPointCallback != nil {
			this.resultPointCallback(point)
		}
	}
	return nil
}
