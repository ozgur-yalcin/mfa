package encoder

import (
	errors "golang.org/x/xerrors"
)

const (
	maskUtilN1 = 3
	maskUtilN2 = 3
	maskUtilN3 = 40
	maskUtilN4 = 10
)

func MaskUtil_applyMaskPenaltyRule1(matrix *ByteMatrix) int {
	return applyMaskPenaltyRule1Internal(matrix, true) + applyMaskPenaltyRule1Internal(matrix, false)
}

func MaskUtil_applyMaskPenaltyRule2(matrix *ByteMatrix) int {
	penalty := 0
	array := matrix.GetArray()
	width := matrix.GetWidth()
	height := matrix.GetHeight()
	for y := 0; y < height-1; y++ {
		arrayY := array[y]
		for x := 0; x < width-1; x++ {
			value := arrayY[x]
			if value == arrayY[x+1] && value == array[y+1][x] && value == array[y+1][x+1] {
				penalty++
			}
		}
	}
	return maskUtilN2 * penalty
}

func MaskUtil_applyMaskPenaltyRule3(matrix *ByteMatrix) int {
	numPenalties := 0
	array := matrix.GetArray()
	width := matrix.GetWidth()
	height := matrix.GetHeight()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			arrayY := array[y]
			if x+6 < width &&
				arrayY[x] == 1 &&
				arrayY[x+1] == 0 &&
				arrayY[x+2] == 1 &&
				arrayY[x+3] == 1 &&
				arrayY[x+4] == 1 &&
				arrayY[x+5] == 0 &&
				arrayY[x+6] == 1 &&
				(isWhiteHorizontal(arrayY, x-4, x) || isWhiteHorizontal(arrayY, x+7, x+11)) {
				numPenalties++
			}
			if y+6 < height &&
				array[y][x] == 1 &&
				array[y+1][x] == 0 &&
				array[y+2][x] == 1 &&
				array[y+3][x] == 1 &&
				array[y+4][x] == 1 &&
				array[y+5][x] == 0 &&
				array[y+6][x] == 1 &&
				(isWhiteVertical(array, x, y-4, y) || isWhiteVertical(array, x, y+7, y+11)) {
				numPenalties++
			}
		}
	}
	return numPenalties * maskUtilN3
}

func isWhiteHorizontal(rowArray []int8, from, to int) bool {
	if from < 0 {
		from = 0
	}
	if to > len(rowArray) {
		to = len(rowArray)
	}
	for i := from; i < to; i++ {
		if rowArray[i] == 1 {
			return false
		}
	}
	return true
}

func isWhiteVertical(array [][]int8, col, from, to int) bool {
	if from < 0 {
		from = 0
	}
	if to > len(array) {
		to = len(array)
	}
	for i := from; i < to; i++ {
		if array[i][col] == 1 {
			return false
		}
	}
	return true
}

func MaskUtil_applyMaskPenaltyRule4(matrix *ByteMatrix) int {
	numDarkCells := 0
	array := matrix.GetArray()
	width := matrix.GetWidth()
	height := matrix.GetHeight()
	for y := 0; y < height; y++ {
		arrayY := array[y]
		for x := 0; x < width; x++ {
			if arrayY[x] == 1 {
				numDarkCells++
			}
		}
	}
	numTotalCells := matrix.GetHeight() * matrix.GetWidth()
	distance := numDarkCells*2 - numTotalCells
	if distance < 0 {
		distance = -distance
	}
	fivePercentVariances := distance * 10 / numTotalCells
	return fivePercentVariances * maskUtilN4
}

func MaskUtil_getDataMaskBit(maskPattern, x, y int) (bool, error) {
	var intermediate int
	switch maskPattern {
	case 0:
		intermediate = (y + x) & 0x1
		break
	case 1:
		intermediate = y & 0x1
		break
	case 2:
		intermediate = x % 3
		break
	case 3:
		intermediate = (y + x) % 3
		break
	case 4:
		intermediate = ((y / 2) + (x / 3)) & 0x1
		break
	case 5:
		temp := y * x
		intermediate = (temp & 0x1) + (temp % 3)
		break
	case 6:
		temp := y * x
		intermediate = ((temp & 0x1) + (temp % 3)) & 0x1
		break
	case 7:
		temp := y * x
		intermediate = ((temp % 3) + ((y + x) & 0x1)) & 0x1
		break
	default:
		return false, errors.Errorf("IllegalArgumentException: Invalid mask pattern: %d", maskPattern)
	}
	return (intermediate == 0), nil
}

func applyMaskPenaltyRule1Internal(matrix *ByteMatrix, isHorizontal bool) int {
	penalty := 0
	iLimit := matrix.GetWidth()
	jLimit := matrix.GetHeight()
	if isHorizontal {
		iLimit, jLimit = jLimit, iLimit
	}
	array := matrix.GetArray()
	for i := 0; i < iLimit; i++ {
		numSameBitCells := 0
		prevBit := -1
		for j := 0; j < jLimit; j++ {
			var bit int
			if isHorizontal {
				bit = int(array[i][j])
			} else {
				bit = int(array[j][i])
			}
			if bit == prevBit {
				numSameBitCells++
			} else {
				if numSameBitCells >= 5 {
					penalty += maskUtilN1 + (numSameBitCells - 5)
				}
				numSameBitCells = 1
				prevBit = bit
			}
		}
		if numSameBitCells >= 5 {
			penalty += maskUtilN1 + (numSameBitCells - 5)
		}
	}
	return penalty
}
