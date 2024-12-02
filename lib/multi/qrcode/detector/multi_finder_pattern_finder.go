package detector

import (
	"math"
	"sort"

	"github.com/ozgur-yalcin/mfa/lib"
	"github.com/ozgur-yalcin/mfa/lib/qrcode/detector"
)

type MultiFinderPatternFinder struct {
	*detector.FinderPatternFinder
}

const (
	MAX_MODULE_COUNT_PER_EDGE   = 180
	MIN_MODULE_COUNT_PER_EDGE   = 9
	DIFF_MODSIZE_CUTOFF_PERCENT = 0.05
	DIFF_MODSIZE_CUTOFF         = 0.5
)

func ModuleSizeComparator(possibleCenters []*detector.FinderPattern) func(int, int) bool {
	return func(i, j int) bool {
		center1 := possibleCenters[i]
		center2 := possibleCenters[j]
		value := center2.GetEstimatedModuleSize() - center1.GetEstimatedModuleSize()
		return value < 0
	}
}

func NewMultiFinderPatternFinder(image *lib.BitMatrix, resultPointCallback lib.ResultPointCallback) *MultiFinderPatternFinder {
	return &MultiFinderPatternFinder{
		detector.NewFinderPatternFinder(image, resultPointCallback),
	}
}

func (this *MultiFinderPatternFinder) selectMultipleBestPatterns() ([][]*detector.FinderPattern, error) {
	possibleCenters := this.GetPossibleCenters()
	size := len(possibleCenters)

	if size < 3 {
		return nil, lib.NewNotFoundException("Couldn't find enough finder patterns (%d)", size)
	}

	if size == 3 {
		return [][]*detector.FinderPattern{
			{
				possibleCenters[0],
				possibleCenters[1],
				possibleCenters[2],
			},
		}, nil
	}

	sort.Slice(possibleCenters, ModuleSizeComparator(possibleCenters))
	results := make([][]*detector.FinderPattern, 0)

	for i1 := 0; i1 < (size - 2); i1++ {
		p1 := possibleCenters[i1]
		if p1 == nil {
			continue
		}

		for i2 := i1 + 1; i2 < (size - 1); i2++ {
			p2 := possibleCenters[i2]
			if p2 == nil {
				continue
			}

			vModSize12 := (p1.GetEstimatedModuleSize() - p2.GetEstimatedModuleSize()) /
				math.Min(p1.GetEstimatedModuleSize(), p2.GetEstimatedModuleSize())
			vModSize12A := math.Abs(p1.GetEstimatedModuleSize() - p2.GetEstimatedModuleSize())
			if vModSize12A > DIFF_MODSIZE_CUTOFF && vModSize12 >= DIFF_MODSIZE_CUTOFF_PERCENT {
				break
			}

			for i3 := i2 + 1; i3 < size; i3++ {
				p3 := possibleCenters[i3]
				if p3 == nil {
					continue
				}
				vModSize23 := (p2.GetEstimatedModuleSize() - p3.GetEstimatedModuleSize()) /
					math.Min(p2.GetEstimatedModuleSize(), p3.GetEstimatedModuleSize())
				vModSize23A := math.Abs(p2.GetEstimatedModuleSize() - p3.GetEstimatedModuleSize())
				if vModSize23A > DIFF_MODSIZE_CUTOFF && vModSize23 >= DIFF_MODSIZE_CUTOFF_PERCENT {
					break
				}

				bl, tl, tr := lib.ResultPoint_OrderBestPatterns(p1, p2, p3)
				test := []*detector.FinderPattern{
					bl.(*detector.FinderPattern), tl.(*detector.FinderPattern), tr.(*detector.FinderPattern),
				}

				info := detector.NewFinderPatternInfo(test[0], test[1], test[2])
				dA := lib.ResultPoint_Distance(info.GetTopLeft(), info.GetBottomLeft())
				dC := lib.ResultPoint_Distance(info.GetTopRight(), info.GetBottomLeft())
				dB := lib.ResultPoint_Distance(info.GetTopLeft(), info.GetTopRight())

				estimatedModuleCount := (dA + dB) / (p1.GetEstimatedModuleSize() * 2.0)
				if estimatedModuleCount > MAX_MODULE_COUNT_PER_EDGE ||
					estimatedModuleCount < MIN_MODULE_COUNT_PER_EDGE {
					continue
				}

				vABBC := math.Abs((dA - dB) / math.Min(dA, dB))
				if vABBC >= 0.1 {
					continue
				}

				dCpy := math.Sqrt(dA*dA + dB*dB)
				vPyC := math.Abs((dC - dCpy) / math.Min(dC, dCpy))

				if vPyC >= 0.1 {
					continue
				}
				results = append(results, test)
			}
		}
	}
	if len(results) > 0 {
		return results, nil
	}
	return nil, lib.NewNotFoundException()
}

func (this *MultiFinderPatternFinder) FindMulti(hints map[lib.DecodeHintType]interface{}) ([]*detector.FinderPatternInfo, error) {
	_, tryHarder := hints[lib.DecodeHintType_TRY_HARDER]
	image := this.GetImage()
	maxI := image.GetHeight()
	maxJ := image.GetWidth()
	iSkip := (3 * maxI) / (4 * detector.FinderPatternFinder_MAX_MODULES)
	if iSkip < detector.FinderPatternFinder_MIN_SKIP || tryHarder {
		iSkip = detector.FinderPatternFinder_MIN_SKIP
	}

	stateCount := make([]int, 5)
	for i := iSkip - 1; i < maxI; i += iSkip {
		detector.FinderPatternFinder_doClearCounts(stateCount)
		currentState := 0
		for j := 0; j < maxJ; j++ {
			if image.Get(j, i) {
				if (currentState & 1) == 1 {
					currentState++
				}
				stateCount[currentState]++
			} else {
				if (currentState & 1) == 0 {
					if currentState == 4 {
						if detector.FinderPatternFinder_foundPatternCross(stateCount) &&
							this.HandlePossibleCenter(stateCount, i, j) {
							currentState = 0
							detector.FinderPatternFinder_doClearCounts(stateCount)
						} else {
							detector.FinderPatternFinder_doShiftCounts2(stateCount)
							currentState = 3
						}
					} else {
						currentState++
						stateCount[currentState]++
					}
				} else {
					stateCount[currentState]++
				}
			}
		}

		if detector.FinderPatternFinder_foundPatternCross(stateCount) {
			this.HandlePossibleCenter(stateCount, i, maxJ)
		}
	}
	patternInfo, e := this.selectMultipleBestPatterns()
	if e != nil {
		return nil, e
	}
	result := make([]*detector.FinderPatternInfo, 0)
	for _, pattern := range patternInfo {
		bl, tl, tr := lib.ResultPoint_OrderBestPatterns(pattern[0], pattern[1], pattern[2])
		result = append(result,
			detector.NewFinderPatternInfo(
				bl.(*detector.FinderPattern), tl.(*detector.FinderPattern), tr.(*detector.FinderPattern)))
	}

	return result, nil
}
