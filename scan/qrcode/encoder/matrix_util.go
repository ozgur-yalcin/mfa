package encoder

import (
	"errors"
	"math/bits"

	"github.com/ozgur-yalcin/mfa/scan"
	"github.com/ozgur-yalcin/mfa/scan/qrcode/decoder"
)

var (
	matrixUtil_POSITION_DETECTION_PATTERN = [][]int8{
		{1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 1, 1, 1, 0, 1},
		{1, 0, 1, 1, 1, 0, 1},
		{1, 0, 1, 1, 1, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1},
	}

	matrixUtil_POSITION_ADJUSTMENT_PATTERN = [][]int8{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 1, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	}

	matrixUtil_POSITION_ADJUSTMENT_PATTERN_COORDINATE_TABLE = [][]int{
		{-1, -1, -1, -1, -1, -1, -1},
		{6, 18, -1, -1, -1, -1, -1},
		{6, 22, -1, -1, -1, -1, -1},
		{6, 26, -1, -1, -1, -1, -1},
		{6, 30, -1, -1, -1, -1, -1},
		{6, 34, -1, -1, -1, -1, -1},
		{6, 22, 38, -1, -1, -1, -1},
		{6, 24, 42, -1, -1, -1, -1},
		{6, 26, 46, -1, -1, -1, -1},
		{6, 28, 50, -1, -1, -1, -1},
		{6, 30, 54, -1, -1, -1, -1},
		{6, 32, 58, -1, -1, -1, -1},
		{6, 34, 62, -1, -1, -1, -1},
		{6, 26, 46, 66, -1, -1, -1},
		{6, 26, 48, 70, -1, -1, -1},
		{6, 26, 50, 74, -1, -1, -1},
		{6, 30, 54, 78, -1, -1, -1},
		{6, 30, 56, 82, -1, -1, -1},
		{6, 30, 58, 86, -1, -1, -1},
		{6, 34, 62, 90, -1, -1, -1},
		{6, 28, 50, 72, 94, -1, -1},
		{6, 26, 50, 74, 98, -1, -1},
		{6, 30, 54, 78, 102, -1, -1},
		{6, 28, 54, 80, 106, -1, -1},
		{6, 32, 58, 84, 110, -1, -1},
		{6, 30, 58, 86, 114, -1, -1},
		{6, 34, 62, 90, 118, -1, -1},
		{6, 26, 50, 74, 98, 122, -1},
		{6, 30, 54, 78, 102, 126, -1},
		{6, 26, 52, 78, 104, 130, -1},
		{6, 30, 56, 82, 108, 134, -1},
		{6, 34, 60, 86, 112, 138, -1},
		{6, 30, 58, 86, 114, 142, -1},
		{6, 34, 62, 90, 118, 146, -1},
		{6, 30, 54, 78, 102, 126, 150},
		{6, 24, 50, 76, 102, 128, 154},
		{6, 28, 54, 80, 106, 132, 158},
		{6, 32, 58, 84, 110, 136, 162},
		{6, 26, 54, 82, 110, 138, 166},
		{6, 30, 58, 86, 114, 142, 170},
	}

	matrixUtil_TYPE_INFO_COORDINATES = [][]int{
		{8, 0},
		{8, 1},
		{8, 2},
		{8, 3},
		{8, 4},
		{8, 5},
		{8, 7},
		{8, 8},
		{7, 8},
		{5, 8},
		{4, 8},
		{3, 8},
		{2, 8},
		{1, 8},
		{0, 8},
	}

	matrixUtil_VERSION_INFO_POLY = 0x1f25

	matrixUtil_TYPE_INFO_POLY         = 0x537
	matrixUtil_TYPE_INFO_MASK_PATTERN = 0x5412
)

func clearMatrix(matrix *ByteMatrix) {
	matrix.Clear(-1)
}

func MatrixUtil_buildMatrix(
	dataBits *scan.BitArray,
	ecLevel decoder.ErrorCorrectionLevel,
	version *decoder.Version,
	maskPattern int,
	matrix *ByteMatrix) error {

	clearMatrix(matrix)

	e := embedBasicPatterns(version, matrix)
	if e == nil {
		e = embedTypeInfo(ecLevel, maskPattern, matrix)
	}
	if e == nil {
		e = maybeEmbedVersionInfo(version, matrix)
	}
	if e == nil {
		e = embedDataBits(dataBits, maskPattern, matrix)
	}
	return e
}

func embedBasicPatterns(version *decoder.Version, matrix *ByteMatrix) scan.WriterException {
	e := embedPositionDetectionPatternsAndSeparators(matrix)
	if e != nil {
		return e
	}
	e = embedDarkDotAtLeftBottomCorner(matrix)
	if e != nil {
		return e
	}
	maybeEmbedPositionAdjustmentPatterns(version, matrix)

	embedTimingPatterns(matrix)

	return nil
}

func embedTypeInfo(ecLevel decoder.ErrorCorrectionLevel, maskPattern int, matrix *ByteMatrix) scan.WriterException {
	typeInfoBits := scan.NewEmptyBitArray()

	e := makeTypeInfoBits(ecLevel, maskPattern, typeInfoBits)
	if e != nil {
		return e
	}

	for i := 0; i < typeInfoBits.GetSize(); i++ {
		bit := typeInfoBits.Get(typeInfoBits.GetSize() - 1 - i)

		coordinates := matrixUtil_TYPE_INFO_COORDINATES[i]
		x1 := coordinates[0]
		y1 := coordinates[1]
		matrix.SetBool(x1, y1, bit)

		var x2, y2 int
		if i < 8 {
			x2 = matrix.GetWidth() - i - 1
			y2 = 8
		} else {
			x2 = 8
			y2 = matrix.GetHeight() - 7 + (i - 8)
			matrix.SetBool(x2, y2, bit)
		}
		matrix.SetBool(x2, y2, bit)
	}

	return nil
}

func maybeEmbedVersionInfo(version *decoder.Version, matrix *ByteMatrix) scan.WriterException {
	if version.GetVersionNumber() < 7 {
		return nil
	}
	versionInfoBits := scan.NewEmptyBitArray()
	e := makeVersionInfoBits(version, versionInfoBits)
	if e != nil {
		return e
	}

	bitIndex := 6*3 - 1
	for i := 0; i < 6; i++ {
		for j := 0; j < 3; j++ {
			bit := versionInfoBits.Get(bitIndex)
			bitIndex--
			matrix.SetBool(i, matrix.GetHeight()-11+j, bit)
			matrix.SetBool(matrix.GetHeight()-11+j, i, bit)
		}
	}
	return nil
}

func embedDataBits(dataBits *scan.BitArray, maskPattern int, matrix *ByteMatrix) scan.WriterException {
	bitIndex := 0
	direction := -1
	x := matrix.GetWidth() - 1
	y := matrix.GetHeight() - 1
	for x > 0 {
		if x == 6 {
			x -= 1
		}
		for y >= 0 && y < matrix.GetHeight() {
			for i := 0; i < 2; i++ {
				xx := x - i
				if !isEmpty(matrix.Get(xx, y)) {
					continue
				}
				var bit bool
				if bitIndex < dataBits.GetSize() {
					bit = dataBits.Get(bitIndex)
					bitIndex++
				} else {
					bit = false
				}

				if maskPattern != -1 {
					maskBit, e := MaskUtil_getDataMaskBit(maskPattern, xx, y)
					if e != nil {
						return scan.WrapWriterException(e)
					}
					if maskBit {
						bit = !bit
					}
				}
				matrix.SetBool(xx, y, bit)
			}
			y += direction
		}
		direction = -direction
		y += direction
		x -= 2
	}
	if bitIndex != dataBits.GetSize() {
		return scan.NewWriterException(
			"Not all bits consumed: %v/%v", bitIndex, dataBits.GetSize())
	}
	return nil
}

func findMSBSet(value int) int {
	return 32 - bits.LeadingZeros32(uint32(value))
}

//	x^2
//	__________________________________________________
//

// x^14 + x^13 + x^12 + x^11 + x^10 + x^7 + x^4 + x^2
// --------------------------------------------------
//
//	x^11 + x^10 + x^7 + x^4 + x^2
func calculateBCHCode(value, poly int) (int, error) {
	if poly == 0 {
		return 0, errors.New("IllegalArgumentException: 0 polynomial")
	}
	msbSetInPoly := findMSBSet(poly)

	value <<= uint(msbSetInPoly - 1)
	for findMSBSet(value) >= msbSetInPoly {
		value ^= poly << uint(findMSBSet(value)-msbSetInPoly)
	}
	return value, nil
}

func makeTypeInfoBits(ecLevel decoder.ErrorCorrectionLevel, maskPattern int, bits *scan.BitArray) scan.WriterException {
	if !QRCode_IsValidMaskPattern(maskPattern) {
		return scan.NewWriterException("Invalid mask pattern")
	}
	typeInfo := (ecLevel.GetBits() << 3) | maskPattern
	bits.AppendBits(typeInfo, 5)

	bchCode, _ := calculateBCHCode(typeInfo, matrixUtil_TYPE_INFO_POLY)
	bits.AppendBits(bchCode, 10)

	maskBits := scan.NewEmptyBitArray()
	maskBits.AppendBits(matrixUtil_TYPE_INFO_MASK_PATTERN, 15)
	bits.Xor(maskBits)

	if bits.GetSize() != 15 {
		return scan.NewWriterException(
			"should not happen but we got: %v", bits.GetSize())
	}

	return nil
}

func makeVersionInfoBits(version *decoder.Version, bits *scan.BitArray) scan.WriterException {
	bits.AppendBits(version.GetVersionNumber(), 6)
	bchCode, _ := calculateBCHCode(version.GetVersionNumber(), matrixUtil_VERSION_INFO_POLY)
	bits.AppendBits(bchCode, 12)

	if bits.GetSize() != 18 {
		return scan.NewWriterException(
			"should not happen but we got: %v", bits.GetSize())
	}

	return nil
}

func isEmpty(value int8) bool {
	return value == -1
}

func embedTimingPatterns(matrix *ByteMatrix) {
	for i := 8; i < matrix.GetWidth()-8; i++ {
		bit := int8((i + 1) % 2)
		if isEmpty(matrix.Get(i, 6)) {
			matrix.Set(i, 6, bit)
		}
		if isEmpty(matrix.Get(6, i)) {
			matrix.Set(6, i, bit)
		}
	}
}

func embedDarkDotAtLeftBottomCorner(matrix *ByteMatrix) scan.WriterException {
	if matrix.Get(8, matrix.GetHeight()-8) == 0 {
		return scan.NewWriterException("embedDarkDotAtLeftBottomCorner")
	}
	matrix.Set(8, matrix.GetHeight()-8, 1)
	return nil
}

func embedHorizontalSeparationPattern(xStart, yStart int, matrix *ByteMatrix) scan.WriterException {
	for x := 0; x < 8; x++ {
		if !isEmpty(matrix.Get(xStart+x, yStart)) {
			return scan.NewWriterException(
				"embedHorizontalSeparationPattern(%d, %d)", xStart, yStart)
		}
		matrix.Set(xStart+x, yStart, 0)
	}
	return nil
}

func embedVerticalSeparationPattern(xStart, yStart int, matrix *ByteMatrix) scan.WriterException {
	for y := 0; y < 7; y++ {
		if !isEmpty(matrix.Get(xStart, yStart+y)) {
			return scan.NewWriterException(
				"embedVerticalSeparationPattern(%d, %d)", xStart, yStart)
		}
		matrix.Set(xStart, yStart+y, 0)
	}
	return nil
}

func embedPositionAdjustmentPattern(xStart, yStart int, matrix *ByteMatrix) {
	for y := 0; y < 5; y++ {
		patternY := matrixUtil_POSITION_ADJUSTMENT_PATTERN[y]
		for x := 0; x < 5; x++ {
			matrix.Set(xStart+x, yStart+y, patternY[x])
		}
	}
}

func embedPositionDetectionPattern(xStart, yStart int, matrix *ByteMatrix) {
	for y := 0; y < 7; y++ {
		patternY := matrixUtil_POSITION_DETECTION_PATTERN[y]
		for x := 0; x < 7; x++ {
			matrix.Set(xStart+x, yStart+y, patternY[x])
		}
	}
}

func embedPositionDetectionPatternsAndSeparators(matrix *ByteMatrix) scan.WriterException {
	pdpWidth := len(matrixUtil_POSITION_DETECTION_PATTERN[0])
	embedPositionDetectionPattern(0, 0, matrix)
	embedPositionDetectionPattern(matrix.GetWidth()-pdpWidth, 0, matrix)
	embedPositionDetectionPattern(0, matrix.GetWidth()-pdpWidth, matrix)

	hspWidth := 8
	e := embedHorizontalSeparationPattern(0, hspWidth-1, matrix)
	if e == nil {
		e = embedHorizontalSeparationPattern(matrix.GetWidth()-hspWidth, hspWidth-1, matrix)
	}
	if e == nil {
		e = embedHorizontalSeparationPattern(0, matrix.GetWidth()-hspWidth, matrix)
	}

	vspSize := 7
	if e == nil {
		e = embedVerticalSeparationPattern(vspSize, 0, matrix)
	}
	if e == nil {
		e = embedVerticalSeparationPattern(matrix.GetHeight()-vspSize-1, 0, matrix)
	}
	if e == nil {
		e = embedVerticalSeparationPattern(vspSize, matrix.GetHeight()-vspSize, matrix)
	}

	return e
}

func maybeEmbedPositionAdjustmentPatterns(version *decoder.Version, matrix *ByteMatrix) {
	if version.GetVersionNumber() < 2 {
		return
	}
	index := version.GetVersionNumber() - 1
	coordinates := matrixUtil_POSITION_ADJUSTMENT_PATTERN_COORDINATE_TABLE[index]
	for _, y := range coordinates {
		if y >= 0 {
			for _, x := range coordinates {
				if x >= 0 && isEmpty(matrix.Get(x, y)) {
					embedPositionAdjustmentPattern(x-2, y-2, matrix)
				}
			}
		}
	}
}
