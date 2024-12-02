package encoder

import (
	"fmt"
	"math"
	"strconv"
	"unicode/utf8"

	textencoding "golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"

	"github.com/ozgur-yalcin/mfa/scan"
	"github.com/ozgur-yalcin/mfa/scan/common"
	"github.com/ozgur-yalcin/mfa/scan/common/reedsolomon"
	"github.com/ozgur-yalcin/mfa/scan/qrcode/decoder"
)

var (
	Encoder_DEFAULT_BYTE_MODE_ENCODING textencoding.Encoding = unicode.UTF8
)

var alphanumericTable = []int{
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	36, -1, -1, -1, 37, 38, -1, -1, -1, -1, 39, 40, -1, 41, 42, 43,
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 44, -1, -1, -1, -1, -1,
	-1, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24,
	25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, -1, -1, -1, -1, -1,
}

func calculateMaskPenalty(matrix *ByteMatrix) int {
	return MaskUtil_applyMaskPenaltyRule1(matrix) +
		MaskUtil_applyMaskPenaltyRule2(matrix) +
		MaskUtil_applyMaskPenaltyRule3(matrix) +
		MaskUtil_applyMaskPenaltyRule4(matrix)
}

func Encoder_encodeWithoutHint(content string, ecLevel decoder.ErrorCorrectionLevel) (*QRCode, scan.WriterException) {
	return Encoder_encode(content, ecLevel, nil)
}

func Encoder_encode(content string, ecLevel decoder.ErrorCorrectionLevel, hints map[scan.EncodeHintType]interface{}) (*QRCode, scan.WriterException) {
	encoding := Encoder_DEFAULT_BYTE_MODE_ENCODING
	encodingHint, hasEncodingHint := hints[scan.EncodeHintType_CHARACTER_SET]
	if hasEncodingHint {
		if eci, ok := common.GetCharacterSetECIByName(fmt.Sprintf("%v", encodingHint)); ok {
			encoding = eci.GetCharset()
		} else {
			return nil, scan.NewWriterException(encodingHint)
		}
	}

	mode := chooseMode(content, encoding)

	headerBits := scan.NewEmptyBitArray()

	if mode == decoder.Mode_BYTE && hasEncodingHint {
		eci, ok := common.GetCharacterSetECI(encoding)
		if ok && eci != nil {
			appendECI(eci, headerBits)
		}
	}

	gs1FormatHint, hasGS1FormatHint := hints[scan.EncodeHintType_GS1_FORMAT]
	if hasGS1FormatHint {
		appendGS1, ok := gs1FormatHint.(bool)
		if !ok {
			s, ok := gs1FormatHint.(string)
			if ok {
				appendGS1, _ = strconv.ParseBool(s)
			}
		}
		if appendGS1 {
			appendModeInfo(decoder.Mode_FNC1_FIRST_POSITION, headerBits)
		}
	}

	appendModeInfo(mode, headerBits)

	dataBits := scan.NewEmptyBitArray()
	e := appendBytes(content, mode, dataBits, encoding)
	if e != nil {
		return nil, e
	}

	var version *decoder.Version
	if versionHint, ok := hints[scan.EncodeHintType_QR_VERSION]; ok {
		versionNumber, ok := versionHint.(int)
		if !ok {
			if s, ok := versionHint.(string); ok {
				versionNumber, _ = strconv.Atoi(s)
			}
		}
		var e error
		version, e = decoder.Version_GetVersionForNumber(versionNumber)
		if e != nil {
			return nil, scan.WrapWriterException(e)
		}
		bitsNeeded := calculateBitsNeeded(mode, headerBits, dataBits, version)
		if !willFit(bitsNeeded, version, ecLevel) {
			return nil, scan.NewWriterException("Data too big for requested version")
		}
	} else {
		version, e = recommendVersion(ecLevel, mode, headerBits, dataBits)
		if e != nil {
			return nil, e
		}
	}

	headerAndDataBits := scan.NewEmptyBitArray()
	headerAndDataBits.AppendBitArray(headerBits)
	numLetters := len(content)
	if mode == decoder.Mode_BYTE {
		numLetters = dataBits.GetSizeInBytes()
	} else if mode == decoder.Mode_KANJI {
		numLetters = utf8.RuneCountInString(content)
	}

	e = appendLengthInfo(numLetters, version, mode, headerAndDataBits)
	if e != nil {
		return nil, e
	}
	headerAndDataBits.AppendBitArray(dataBits)

	ecBlocks := version.GetECBlocksForLevel(ecLevel)
	numDataBytes := version.GetTotalCodewords() - ecBlocks.GetTotalECCodewords()

	e = terminateBits(numDataBytes, headerAndDataBits)
	if e != nil {
		return nil, e
	}

	finalBits, e := interleaveWithECBytes(
		headerAndDataBits, version.GetTotalCodewords(), numDataBytes, ecBlocks.GetNumBlocks())
	if e != nil {
		return nil, e
	}

	qrCode := NewQRCode()

	qrCode.SetECLevel(ecLevel)
	qrCode.SetMode(mode)
	qrCode.SetVersion(version)

	dimension := version.GetDimensionForVersion()
	matrix := NewByteMatrix(dimension, dimension)

	maskPattern := -1
	if hintMaskPattern, ok := hints[scan.EncodeHintType_QR_MASK_PATTERN]; ok {
		switch mask := hintMaskPattern.(type) {
		case int:
			maskPattern = mask
		case string:
			if m, e := strconv.Atoi(mask); e == nil {
				maskPattern = m
			}
		}
		if !QRCode_IsValidMaskPattern(maskPattern) {
			maskPattern = -1
		}
	}

	if maskPattern == -1 {
		maskPattern, e = chooseMaskPattern(finalBits, ecLevel, version, matrix)
		if e != nil {
			return nil, e
		}
	}
	qrCode.SetMaskPattern(maskPattern)

	_ = MatrixUtil_buildMatrix(finalBits, ecLevel, version, maskPattern, matrix)
	qrCode.SetMatrix(matrix)

	return qrCode, nil
}

func recommendVersion(ecLevel decoder.ErrorCorrectionLevel, mode *decoder.Mode,
	headerBits *scan.BitArray, dataBits *scan.BitArray) (*decoder.Version, scan.WriterException) {
	version1, _ := decoder.Version_GetVersionForNumber(1)
	provisionalBitsNeeded := calculateBitsNeeded(mode, headerBits, dataBits, version1)
	provisionalVersion, e := chooseVersion(provisionalBitsNeeded, ecLevel)
	if e != nil {
		return nil, e
	}

	bitsNeeded := calculateBitsNeeded(mode, headerBits, dataBits, provisionalVersion)
	return chooseVersion(bitsNeeded, ecLevel)
}

func calculateBitsNeeded(
	mode *decoder.Mode,
	headerBits *scan.BitArray,
	dataBits *scan.BitArray,
	version *decoder.Version) int {
	return headerBits.GetSize() + mode.GetCharacterCountBits(version) + dataBits.GetSize()
}

func getAlphanumericCode(code uint8) int {
	if int(code) < len(alphanumericTable) {
		return alphanumericTable[code]
	}
	return -1
}

func chooseMode(content string, encoding textencoding.Encoding) *decoder.Mode {
	if common.StringUtils_SHIFT_JIS_CHARSET == encoding && isOnlyDoubleByteKanji(content) {
		return decoder.Mode_KANJI
	}
	hasNumeric := false
	hasAlphanumeric := false
	for i := 0; i < len(content); i++ {
		c := content[i]
		if c >= '0' && c <= '9' {
			hasNumeric = true
		} else if getAlphanumericCode(c) != -1 {
			hasAlphanumeric = true
		} else {
			return decoder.Mode_BYTE
		}
	}
	if hasAlphanumeric {
		return decoder.Mode_ALPHANUMERIC
	}
	if hasNumeric {
		return decoder.Mode_NUMERIC
	}
	return decoder.Mode_BYTE
}

func isOnlyDoubleByteKanji(content string) bool {
	enc := common.StringUtils_SHIFT_JIS_CHARSET.NewEncoder()
	bytes, e := enc.Bytes([]byte(content))
	if e != nil {
		return false
	}

	length := len(bytes)
	if length%2 != 0 {
		return false
	}
	for i := 0; i < length; i += 2 {
		byte1 := bytes[i] & 0xFF
		if (byte1 < 0x81 || byte1 > 0x9F) && (byte1 < 0xE0 || byte1 > 0xEB) {
			return false
		}
	}
	return true
}

func chooseMaskPattern(bits *scan.BitArray, ecLevel decoder.ErrorCorrectionLevel,
	version *decoder.Version, matrix *ByteMatrix) (int, scan.WriterException) {

	minPenalty := math.MaxInt32
	bestMaskPattern := -1
	for maskPattern := 0; maskPattern < QRCode_NUM_MASK_PATERNS; maskPattern++ {
		e := MatrixUtil_buildMatrix(bits, ecLevel, version, maskPattern, matrix)
		if e != nil {
			return -1, scan.WrapWriterException(e)
		}
		penalty := calculateMaskPenalty(matrix)
		if penalty < minPenalty {
			minPenalty = penalty
			bestMaskPattern = maskPattern
		}
	}
	return bestMaskPattern, nil
}

func chooseVersion(numInputBits int, ecLevel decoder.ErrorCorrectionLevel) (*decoder.Version, scan.WriterException) {
	for versionNum := 1; versionNum <= 40; versionNum++ {
		version, _ := decoder.Version_GetVersionForNumber(versionNum)
		if willFit(numInputBits, version, ecLevel) {
			return version, nil
		}
	}
	return nil, scan.NewWriterException("Data too big")
}

func willFit(numInputBits int, version *decoder.Version, ecLevel decoder.ErrorCorrectionLevel) bool {
	numBytes := version.GetTotalCodewords()
	ecBlocks := version.GetECBlocksForLevel(ecLevel)
	numEcBytes := ecBlocks.GetTotalECCodewords()
	numDataBytes := numBytes - numEcBytes
	totalInputBytes := (numInputBits + 7) / 8
	return numDataBytes >= totalInputBytes
}

func terminateBits(numDataBytes int, bits *scan.BitArray) scan.WriterException {
	capacity := numDataBytes * 8
	if bits.GetSize() > capacity {
		return scan.NewWriterException(
			"data bits cannot fit in the QR Code %v > %v", bits.GetSize(), capacity)
	}
	for i := 0; i < 4 && bits.GetSize() < capacity; i++ {
		bits.AppendBit(false)
	}
	numBitsInLastByte := bits.GetSize() & 0x07
	if numBitsInLastByte > 0 {
		for i := numBitsInLastByte; i < 8; i++ {
			bits.AppendBit(false)
		}
	}
	numPaddingBytes := numDataBytes - bits.GetSizeInBytes()
	for i := 0; i < numPaddingBytes; i++ {
		v := 0x11
		if (i & 0x1) == 0 {
			v = 0xEC
		}
		_ = bits.AppendBits(v, 8)
	}
	if bits.GetSize() != capacity {
		return scan.NewWriterException("bits.GetSize()=%d, capacity=&d", bits.GetSize(), capacity)
	}
	return nil
}

func getNumDataBytesAndNumECBytesForBlockID(numTotalBytes, numDataBytes, numRSBlocks, blockID int) (int, int, scan.WriterException) {
	if blockID >= numRSBlocks {
		return 0, 0, scan.NewWriterException("Block ID too large")
	}
	numRsBlocksInGroup2 := numTotalBytes % numRSBlocks
	numRsBlocksInGroup1 := numRSBlocks - numRsBlocksInGroup2
	numTotalBytesInGroup1 := numTotalBytes / numRSBlocks
	numTotalBytesInGroup2 := numTotalBytesInGroup1 + 1
	numDataBytesInGroup1 := numDataBytes / numRSBlocks
	numDataBytesInGroup2 := numDataBytesInGroup1 + 1
	numEcBytesInGroup1 := numTotalBytesInGroup1 - numDataBytesInGroup1
	numEcBytesInGroup2 := numTotalBytesInGroup2 - numDataBytesInGroup2
	if numEcBytesInGroup1 != numEcBytesInGroup2 {
		return 0, 0, scan.NewWriterException("EC bytes mismatch")
	}
	if numRSBlocks != numRsBlocksInGroup1+numRsBlocksInGroup2 {
		return 0, 0, scan.NewWriterException("RS blocks mismatch")
	}
	if numTotalBytes !=
		((numDataBytesInGroup1+numEcBytesInGroup1)*numRsBlocksInGroup1)+
			((numDataBytesInGroup2+numEcBytesInGroup2)*numRsBlocksInGroup2) {
		return 0, 0, scan.NewWriterException("Total bytes mismatch")
	}

	if blockID < numRsBlocksInGroup1 {
		return numDataBytesInGroup1, numEcBytesInGroup1, nil
	}
	return numDataBytesInGroup2, numEcBytesInGroup2, nil
}

func interleaveWithECBytes(bits *scan.BitArray, numTotalBytes, numDataBytes, numRSBlocks int) (*scan.BitArray, scan.WriterException) {

	if bits.GetSizeInBytes() != numDataBytes {
		return nil, scan.NewWriterException("Number of bits and data bytes does not match")
	}

	dataBytesOffset := 0
	maxNumDataBytes := 0
	maxNumEcBytes := 0

	blocks := make([]*BlockPair, 0)

	for i := 0; i < numRSBlocks; i++ {
		numDataBytesInBlock, numEcBytesInBlock, e := getNumDataBytesAndNumECBytesForBlockID(
			numTotalBytes, numDataBytes, numRSBlocks, i)
		if e != nil {
			return nil, e
		}

		size := numDataBytesInBlock
		dataBytes := make([]byte, size)
		bits.ToBytes(8*dataBytesOffset, dataBytes, 0, size)
		ecBytes, e := generateECBytes(dataBytes, numEcBytesInBlock)
		if e != nil {
			return nil, e
		}
		blocks = append(blocks, NewBlockPair(dataBytes, ecBytes))

		if maxNumDataBytes < size {
			maxNumDataBytes = size
		}
		if maxNumEcBytes < len(ecBytes) {
			maxNumEcBytes = len(ecBytes)
		}
		dataBytesOffset += numDataBytesInBlock
	}
	if numDataBytes != dataBytesOffset {
		return nil, scan.NewWriterException("Data bytes does not match offset")
	}

	result := scan.NewEmptyBitArray()

	for i := 0; i < maxNumDataBytes; i++ {
		for _, block := range blocks {
			dataBytes := block.GetDataBytes()
			if i < len(dataBytes) {
				_ = result.AppendBits(int(dataBytes[i]), 8)
			}
		}
	}
	for i := 0; i < maxNumEcBytes; i++ {
		for _, block := range blocks {
			ecBytes := block.GetErrorCorrectionBytes()
			if i < len(ecBytes) {
				_ = result.AppendBits(int(ecBytes[i]), 8)
			}
		}
	}
	if numTotalBytes != result.GetSizeInBytes() {
		return nil, scan.NewWriterException(
			"Interleaving error: %v  and %v differ", numTotalBytes, result.GetSizeInBytes())
	}

	return result, nil
}

func generateECBytes(dataBytes []byte, numEcBytesInBlock int) ([]byte, scan.WriterException) {
	numDataBytes := len(dataBytes)
	toEncode := make([]int, numDataBytes+numEcBytesInBlock)
	for i := 0; i < numDataBytes; i++ {
		toEncode[i] = int(dataBytes[i]) & 0xFF
	}
	e := reedsolomon.NewReedSolomonEncoder(reedsolomon.GenericGF_QR_CODE_FIELD_256).Encode(toEncode, numEcBytesInBlock)
	if e != nil {
		return nil, scan.WrapWriterException(e)
	}

	ecBytes := make([]byte, numEcBytesInBlock)
	for i := 0; i < numEcBytesInBlock; i++ {
		ecBytes[i] = byte(toEncode[numDataBytes+i])
	}
	return ecBytes, nil
}

func appendModeInfo(mode *decoder.Mode, bits *scan.BitArray) {
	_ = bits.AppendBits(mode.GetBits(), 4)
}

func appendLengthInfo(numLetters int, version *decoder.Version, mode *decoder.Mode, bits *scan.BitArray) scan.WriterException {
	numBits := mode.GetCharacterCountBits(version)
	if numLetters >= (1 << uint(numBits)) {
		return scan.NewWriterException(
			"%v is bigger than %v", numLetters, (1 << uint(numBits)))
	}
	_ = bits.AppendBits(numLetters, numBits)
	return nil
}

func appendBytes(content string, mode *decoder.Mode, bits *scan.BitArray, encoding textencoding.Encoding) scan.WriterException {
	switch mode {
	case decoder.Mode_NUMERIC:
		appendNumericBytes(content, bits)
		return nil
	case decoder.Mode_ALPHANUMERIC:
		return appendAlphanumericBytes(content, bits)
	case decoder.Mode_BYTE:
		return append8BitBytes(content, bits, encoding)
	case decoder.Mode_KANJI:
		return appendKanjiBytes(content, bits)
	default:
		return scan.NewWriterException("Invalid mode: %v", mode)
	}
}

func appendNumericBytes(content string, bits *scan.BitArray) {
	length := len(content)
	i := 0
	for i < length {
		num1 := int(content[i]) - '0'
		if i+2 < length {
			num2 := int(content[i+1]) - '0'
			num3 := int(content[i+2]) - '0'
			_ = bits.AppendBits(num1*100+num2*10+num3, 10)
			i += 3
		} else if i+1 < length {
			num2 := int(content[i+1]) - '0'
			_ = bits.AppendBits(num1*10+num2, 7)
			i += 2
		} else {
			_ = bits.AppendBits(num1, 4)
			i++
		}
	}
}

func appendAlphanumericBytes(content string, bits *scan.BitArray) scan.WriterException {
	length := len(content)
	i := 0
	for i < length {
		code1 := getAlphanumericCode(content[i])
		if code1 == -1 {
			return scan.NewWriterException("appendAlphanumericBytes")
		}
		if i+1 < length {
			code2 := getAlphanumericCode(content[i+1])
			if code2 == -1 {
				return scan.NewWriterException("appendAlphanumericBytes")
			}
			_ = bits.AppendBits(code1*45+code2, 11)
			i += 2
		} else {
			_ = bits.AppendBits(code1, 6)
			i++
		}
	}
	return nil
}

func append8BitBytes(content string, bits *scan.BitArray, encoding textencoding.Encoding) scan.WriterException {
	bytes := []byte(content)

	var e error
	bytes, e = encoding.NewEncoder().Bytes([]byte(content))
	if e != nil {
		return scan.WrapWriterException(e)
	}

	for _, b := range bytes {
		_ = bits.AppendBits(int(b), 8)
	}
	return nil
}

func appendKanjiBytes(content string, bits *scan.BitArray) scan.WriterException {
	enc := common.StringUtils_SHIFT_JIS_CHARSET.NewEncoder()
	bytes, e := enc.Bytes([]byte(content))
	if e != nil {
		return scan.WrapWriterException(e)
	}
	if len(bytes)%2 != 0 {
		return scan.NewWriterException("Kanji byte size not even")
	}
	maxI := len(bytes) - 1
	for i := 0; i < maxI; i += 2 {
		byte1 := int(bytes[i]) & 0xFF
		byte2 := int(bytes[i+1]) & 0xFF
		code := (byte1 << 8) | byte2
		subtracted := -1
		if code >= 0x8140 && code <= 0x9ffc {
			subtracted = code - 0x8140
		} else if code >= 0xe040 && code <= 0xebbf {
			subtracted = code - 0xc140
		}
		if subtracted == -1 {
			return scan.NewWriterException("Invalid byte sequence")
		}
		encoded := ((subtracted >> 8) * 0xc0) + (subtracted & 0xff)
		_ = bits.AppendBits(encoded, 13)
	}
	return nil
}

func appendECI(eci *common.CharacterSetECI, bits *scan.BitArray) {
	_ = bits.AppendBits(decoder.Mode_ECI.GetBits(), 4)
	_ = bits.AppendBits(eci.GetValue(), 8)
}
