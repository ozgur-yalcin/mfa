package lib

const (
	BLOCK_SIZE_POWER  = 3
	BLOCK_SIZE        = 1 << BLOCK_SIZE_POWER
	BLOCK_SIZE_MASK   = BLOCK_SIZE - 1
	MINIMUM_DIMENSION = BLOCK_SIZE * 5
	MIN_DYNAMIC_RANGE = 24
)

type HybridBinarizer struct {
	*GlobalHistogramBinarizer
	matrix *BitMatrix
}

func NewHybridBinarizer(source LuminanceSource) Binarizer {
	return &HybridBinarizer{
		NewGlobalHistgramBinarizer(source).(*GlobalHistogramBinarizer),
		nil,
	}
}

func (this *HybridBinarizer) GetBlackMatrix() (*BitMatrix, error) {
	if this.matrix != nil {
		return this.matrix, nil
	}
	source := this.GetLuminanceSource()
	width := source.GetWidth()
	height := source.GetHeight()
	if width >= MINIMUM_DIMENSION && height >= MINIMUM_DIMENSION {
		luminances := source.GetMatrix()
		subWidth := width >> BLOCK_SIZE_POWER
		if (width & BLOCK_SIZE_MASK) != 0 {
			subWidth++
		}
		subHeight := height >> BLOCK_SIZE_POWER
		if (height & BLOCK_SIZE_MASK) != 0 {
			subHeight++
		}
		blackPoints := this.calculateBlackPoints(luminances, subWidth, subHeight, width, height)

		newMatrix, _ := NewBitMatrix(width, height)
		this.calculateThresholdForBlock(luminances, subWidth, subHeight, width, height, blackPoints, newMatrix)
		this.matrix = newMatrix
	} else {
		newMatrix, e := this.GlobalHistogramBinarizer.GetBlackMatrix()
		if e != nil {
			return nil, e
		}
		this.matrix = newMatrix
	}
	return this.matrix, nil
}

func (this *HybridBinarizer) CreateBinarizer(source LuminanceSource) Binarizer {
	return NewHybridBinarizer(source)
}

func (this *HybridBinarizer) calculateThresholdForBlock(
	luminances []byte, subWidth, subHeight, width, height int, blackPoints [][]int, matrix *BitMatrix) {
	maxYOffset := height - BLOCK_SIZE
	maxXOffset := width - BLOCK_SIZE
	for y := 0; y < subHeight; y++ {
		yoffset := y << BLOCK_SIZE_POWER
		if yoffset > maxYOffset {
			yoffset = maxYOffset
		}
		top := this.cap(y, 2, subHeight-3)
		for x := 0; x < subWidth; x++ {
			xoffset := x << BLOCK_SIZE_POWER
			if xoffset > maxXOffset {
				xoffset = maxXOffset
			}
			left := this.cap(x, 2, subWidth-3)
			sum := 0
			for z := -2; z <= 2; z++ {
				blackRow := blackPoints[top+z]
				sum += blackRow[left-2] + blackRow[left-1] + blackRow[left] + blackRow[left+1] + blackRow[left+2]
			}
			average := sum / 25
			this.thresholdBlock(luminances, xoffset, yoffset, average, width, matrix)
		}
	}
}

func (this *HybridBinarizer) cap(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (this *HybridBinarizer) thresholdBlock(luminances []byte, xoffset, yoffset, threshold, stride int, matrix *BitMatrix) {
	for y, offset := 0, yoffset*stride+xoffset; y < BLOCK_SIZE; y, offset = y+1, offset+stride {
		for x := 0; x < BLOCK_SIZE; x++ {
			if int(luminances[offset+x]&0xFF) <= threshold {
				matrix.Set(xoffset+x, yoffset+y)
			}
		}
	}
}

func (this *HybridBinarizer) calculateBlackPoints(luminances []byte, subWidth, subHeight, width, height int) [][]int {
	maxYOffset := height - BLOCK_SIZE
	maxXOffset := width - BLOCK_SIZE
	blackPoints := make([][]int, subHeight)
	for y := 0; y < subHeight; y++ {
		blackPoints[y] = make([]int, subWidth)
		yoffset := y << BLOCK_SIZE_POWER
		if yoffset > maxYOffset {
			yoffset = maxYOffset
		}
		for x := 0; x < subWidth; x++ {
			xoffset := x << BLOCK_SIZE_POWER
			if xoffset > maxXOffset {
				xoffset = maxXOffset
			}
			sum := 0
			min := 0xFF
			max := 0
			for yy, offset := 0, yoffset*width+xoffset; yy < BLOCK_SIZE; yy, offset = yy+1, offset+width {
				for xx := 0; xx < BLOCK_SIZE; xx++ {
					pixel := int(luminances[offset+xx] & 0xFF)
					sum += pixel
					if pixel < min {
						min = pixel
					}
					if pixel > max {
						max = pixel
					}
				}

				if max-min > MIN_DYNAMIC_RANGE {
					for yy, offset = yy+1, offset+width; yy < BLOCK_SIZE; yy, offset = yy+1, offset+width {
						for xx := 0; xx < BLOCK_SIZE; xx++ {
							sum += int(luminances[offset+xx] & 0xFF)
						}
					}
				}
			}

			average := sum >> (BLOCK_SIZE_POWER * 2)
			if max-min <= MIN_DYNAMIC_RANGE {
				average = min / 2

				if y > 0 && x > 0 {

					averageNeighborBlackPoint :=
						(blackPoints[y-1][x] + (2 * blackPoints[y][x-1]) + blackPoints[y-1][x-1]) / 4
					if min < averageNeighborBlackPoint {
						average = averageNeighborBlackPoint
					}
				}
			}
			blackPoints[y][x] = average
		}
	}
	return blackPoints
}
