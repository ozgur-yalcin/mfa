package decoder

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"

	"github.com/ozgur-yalcin/mfa/scan"
	"github.com/ozgur-yalcin/mfa/scan/common"
)

const GB2312_SUBSET = 1

var ALPHANUMERIC_CHARS = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:")

func DecodedBitStreamParser_Decode(
	bytes []byte, version *Version, ecLevel ErrorCorrectionLevel,
	hints map[scan.DecodeHintType]interface{}) (*common.DecoderResult, error) {

	bits := common.NewBitSource(bytes)
	result := make([]byte, 0, 50)
	byteSegments := make([][]byte, 0, 1)
	symbolSequence := -1
	parityData := -1
	symbologyModifier := 0

	var currentCharacterSetECI *common.CharacterSetECI
	fc1InEffect := false
	hasFNC1first := false
	hasFNC1second := false
	var mode *Mode
	var e error

	for {
		if bits.Available() < 4 {
			mode = Mode_TERMINATOR
		} else {
			bit4, _ := bits.ReadBits(4)
			mode, e = ModeForBits(bit4)
			if e != nil {
				return nil, scan.WrapFormatException(e)
			}
		}
		switch mode {
		case Mode_TERMINATOR:
		case Mode_FNC1_FIRST_POSITION:
			hasFNC1first = true
			fc1InEffect = true
		case Mode_FNC1_SECOND_POSITION:
			hasFNC1second = true
			fc1InEffect = true
		case Mode_STRUCTURED_APPEND:
			symbolSequence, e = bits.ReadBits(8)
			if e != nil {
				return nil, scan.WrapFormatException(e)
			}
			parityData, e = bits.ReadBits(8)
			if e != nil {
				return nil, scan.WrapFormatException(e)
			}
		case Mode_ECI:
			value, e := DecodedBitStreamParser_parseECIValue(bits)
			if e != nil {
				return nil, e
			}
			currentCharacterSetECI, e = common.GetCharacterSetECIByValue(value)
			if e != nil || currentCharacterSetECI == nil {
				return nil, scan.WrapFormatException(e)
			}
		case Mode_HANZI:
			subset, e := bits.ReadBits(4)
			if e != nil {
				return nil, scan.WrapFormatException(e)
			}
			countHanzi, e := bits.ReadBits(mode.GetCharacterCountBits(version))
			if e != nil {
				return nil, scan.WrapFormatException(e)
			}
			if subset == GB2312_SUBSET {
				result, e = DecodedBitStreamParser_decodeHanziSegment(bits, result, countHanzi)
				if e != nil {
					return nil, e
				}
			}
		default:
			count, e := bits.ReadBits(mode.GetCharacterCountBits(version))
			if e != nil {
				return nil, scan.WrapFormatException(e)
			}
			switch mode {
			case Mode_NUMERIC:
				result, e = DecodedBitStreamParser_decodeNumericSegment(bits, result, count)
				if e != nil {
					return nil, e
				}
			case Mode_ALPHANUMERIC:
				result, e = DecodedBitStreamParser_decodeAlphanumericSegment(bits, result, count, fc1InEffect)
				if e != nil {
					return nil, e
				}
			case Mode_BYTE:
				result, byteSegments, e = DecodedBitStreamParser_decodeByteSegment(bits, result, count, currentCharacterSetECI, byteSegments, hints)
				if e != nil {
					return nil, e
				}
			case Mode_KANJI:
				result, e = DecodedBitStreamParser_decodeKanjiSegment(bits, result, count)
				if e != nil {
					return nil, e
				}
			default:
				return nil, scan.NewFormatException("Unknown mode")
			}
			break
		}

		if mode == Mode_TERMINATOR {
			break
		}
	}

	if currentCharacterSetECI != nil {
		if hasFNC1first {
			symbologyModifier = 4
		} else if hasFNC1second {
			symbologyModifier = 6
		} else {
			symbologyModifier = 2
		}
	} else {
		if hasFNC1first {
			symbologyModifier = 3
		} else if hasFNC1second {
			symbologyModifier = 5
		} else {
			symbologyModifier = 1
		}
	}

	if len(byteSegments) == 0 {
		byteSegments = nil
	}
	return common.NewDecoderResultWithParams(bytes,
		string(result),
		byteSegments,
		ecLevel.String(),
		symbolSequence,
		parityData,
		symbologyModifier), nil
}

func DecodedBitStreamParser_decodeHanziSegment(bits *common.BitSource, result []byte, count int) ([]byte, error) {
	if count*13 > bits.Available() {
		return result, scan.NewFormatException("bits.Available() = %v", bits.Available())
	}

	buffer := make([]byte, 2*count)
	offset := 0
	for count > 0 {
		twoBytes, _ := bits.ReadBits(13)
		assembledTwoBytes := ((twoBytes / 0x060) << 8) | (twoBytes % 0x060)
		if assembledTwoBytes < 0x00a00 {
			assembledTwoBytes += 0x0A1A1
		} else {
			assembledTwoBytes += 0x0A6A1
		}
		buffer[offset] = (byte)((assembledTwoBytes >> 8) & 0xFF)
		buffer[offset+1] = (byte)(assembledTwoBytes & 0xFF)
		offset += 2
		count--
	}

	dec := common.StringUtils_GB2312_CHARSET.NewDecoder()
	result, _, e := transform.Append(dec, result, buffer[:offset])
	if e != nil {
		return result, scan.WrapFormatException(e)
	}
	return result, nil
}

func DecodedBitStreamParser_decodeKanjiSegment(bits *common.BitSource, result []byte, count int) ([]byte, error) {
	if count*13 > bits.Available() {
		return result, scan.NewFormatException("bits.Available() = %v", bits.Available())
	}

	buffer := make([]byte, 2*count)
	offset := 0
	for count > 0 {
		twoBytes, _ := bits.ReadBits(13)
		assembledTwoBytes := ((twoBytes / 0x0C0) << 8) | (twoBytes % 0x0C0)
		if assembledTwoBytes < 0x01F00 {
			assembledTwoBytes += 0x08140
		} else {
			assembledTwoBytes += 0x0C140
		}
		buffer[offset] = byte(assembledTwoBytes >> 8)
		buffer[offset+1] = byte(assembledTwoBytes)
		offset += 2
		count--
	}

	dec := common.StringUtils_SHIFT_JIS_CHARSET.NewDecoder()
	result, _, e := transform.Append(dec, result, buffer[:offset])
	if e != nil {
		return result, scan.WrapFormatException(e)
	}
	return result, nil
}

func DecodedBitStreamParser_decodeByteSegment(bits *common.BitSource,
	result []byte, count int, currentCharacterSetECI *common.CharacterSetECI,
	byteSegments [][]byte, hints map[scan.DecodeHintType]interface{}) ([]byte, [][]byte, error) {

	if 8*count > bits.Available() {
		return result, byteSegments, scan.NewFormatException("bits.Available = %v", bits.Available())
	}

	readBytes := make([]byte, count)
	for i := 0; i < count; i++ {
		b, _ := bits.ReadBits(8)
		readBytes[i] = byte(b)
	}

	var encoding encoding.Encoding
	if currentCharacterSetECI == nil {
		var err error
		encoding, err = common.StringUtils_guessCharset(readBytes, hints)
		if err != nil {
			return nil, nil, scan.WrapFormatException(err)
		}
	} else {
		encoding = currentCharacterSetECI.GetCharset()
	}

	dec := encoding.NewDecoder()
	result, _, e := transform.Append(dec, result, readBytes)
	if e != nil {
		return result, byteSegments, scan.WrapFormatException(e)
	}

	byteSegments = append(byteSegments, readBytes)
	return result, byteSegments, nil
}

func toAlphaNumericChar(value int) (byte, error) {
	if value >= len(ALPHANUMERIC_CHARS) {
		return 0, scan.NewFormatException("%v >= len(ALPHANUMERIC_CHARS)", value)
	}
	return ALPHANUMERIC_CHARS[value], nil
}

func DecodedBitStreamParser_decodeAlphanumericSegment(bits *common.BitSource, result []byte, count int, fc1InEffect bool) ([]byte, error) {
	start := len(result)
	for count > 1 {
		nextTwoCharsBits, e := bits.ReadBits(11)
		if e != nil {
			return result, scan.WrapFormatException(e)
		}
		char, e := toAlphaNumericChar(nextTwoCharsBits / 45)
		if e != nil {
			return result, scan.WrapFormatException(e)
		}
		result = append(result, char)
		char, _ = toAlphaNumericChar(nextTwoCharsBits % 45)
		result = append(result, char)
		count -= 2
	}
	if count == 1 {
		nextCharBits, e := bits.ReadBits(6)
		if e != nil {
			return result, scan.WrapFormatException(e)
		}
		char, e := toAlphaNumericChar(nextCharBits)
		if e != nil {
			return result, scan.WrapFormatException(e)
		}
		result = append(result, char)
	}
	if fc1InEffect {
		for i := start; i < len(result); i++ {
			if result[i] == '%' {
				if i < len(result)-1 && result[i+1] == '%' {
					result = append(result[:i], result[i+1:]...)
				} else {
					result[i] = byte(0x1D)
				}
			}
		}
	}
	return result, nil
}

func DecodedBitStreamParser_decodeNumericSegment(bits *common.BitSource, result []byte, count int) ([]byte, error) {
	for count >= 3 {
		threeDigitsBits, e := bits.ReadBits(10)
		if e != nil {
			return result, scan.WrapFormatException(e)
		}
		if threeDigitsBits >= 1000 {
			return result, scan.NewFormatException("threeDigitalBits = %v", threeDigitsBits)
		}
		result = append(result, byte('0'+(threeDigitsBits/100)))
		result = append(result, byte('0'+((threeDigitsBits/10)%10)))
		result = append(result, byte('0'+(threeDigitsBits%10)))
		count -= 3
	}
	if count == 2 {
		twoDigitsBits, e := bits.ReadBits(7)
		if e != nil {
			return result, scan.WrapFormatException(e)
		}
		if twoDigitsBits >= 100 {
			return result, scan.NewFormatException("twoDigitsBits = %v", twoDigitsBits)
		}
		result = append(result, byte('0'+(twoDigitsBits/10)))
		result = append(result, byte('0'+(twoDigitsBits%10)))
	} else if count == 1 {
		digitBits, e := bits.ReadBits(4)
		if e != nil {
			return result, scan.WrapFormatException(e)
		}
		if digitBits >= 10 {
			return result, scan.NewFormatException("digitBits = %v", digitBits)
		}
		result = append(result, byte('0'+digitBits))
	}
	return result, nil
}

func DecodedBitStreamParser_parseECIValue(bits *common.BitSource) (int, error) {
	firstByte, e := bits.ReadBits(8)
	if e != nil {
		return -1, scan.WrapFormatException(e)
	}
	if (firstByte & 0x80) == 0 {
		return firstByte & 0x7F, nil
	}
	if (firstByte & 0xC0) == 0x80 {
		secondByte, e := bits.ReadBits(8)
		if e != nil {
			return -1, scan.WrapFormatException(e)
		}
		return ((firstByte & 0x3F) << 8) | secondByte, nil
	}
	if (firstByte & 0xE0) == 0xC0 {
		secondThirdBytes, e := bits.ReadBits(16)
		if e != nil {
			return -1, scan.WrapFormatException(e)
		}
		return ((firstByte & 0x1F) << 16) | secondThirdBytes, nil
	}
	return -1, scan.NewFormatException()
}
