package common

import (
	"fmt"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"

	"github.com/ozgur-yalcin/mfa/scan"
)

const (
	StringUtils_ASSUME_SHIFT_JIS = false
	StringUtils_SHIFT_JIS        = "SJIS"
	StringUtils_GB2312           = "GB2312"
)

var (
	StringUtils_PLATFORM_DEFAULT_ENCODING = unicode.UTF8
	StringUtils_SHIFT_JIS_CHARSET         = japanese.ShiftJIS
	StringUtils_GB2312_CHARSET            = simplifiedchinese.GB18030
	StringUtils_EUC_JP                    = japanese.EUCJP
)

func StringUtils_guessEncoding(bytes []byte, hints map[scan.DecodeHintType]interface{}) (string, error) {
	c, err := StringUtils_guessCharset(bytes, hints)
	if err != nil {
		return "", err
	}
	if c == StringUtils_SHIFT_JIS_CHARSET {
		return "SJIS", nil
	} else if c == unicode.UTF8 {
		return "UTF8", nil
	} else if c == charmap.ISO8859_1 {
		return "ISO8859_1", nil
	}
	return ianaindex.IANA.Name(c)
}

func StringUtils_guessCharset(bytes []byte, hints map[scan.DecodeHintType]interface{}) (encoding.Encoding, error) {
	if hint, ok := hints[scan.DecodeHintType_CHARACTER_SET]; ok {
		if charset, ok := hint.(encoding.Encoding); ok {
			return charset, nil
		}
		name := fmt.Sprintf("%v", hint)
		if eci, ok := GetCharacterSetECIByName(name); ok {
			return eci.GetCharset(), nil
		}

		return ianaindex.IANA.Encoding(name)
	}

	if len(bytes) > 2 {
		if bytes[0] == 0xfe && bytes[1] == 0xff {
			return unicode.UTF16(unicode.BigEndian, unicode.UseBOM), nil
		}
		if bytes[0] == 0xff && bytes[1] == 0xfe {
			return unicode.UTF16(unicode.LittleEndian, unicode.UseBOM), nil
		}
	}

	length := len(bytes)
	canBeISO88591 := true
	canBeShiftJIS := true
	canBeUTF8 := true
	utf8BytesLeft := 0
	utf2BytesChars := 0
	utf3BytesChars := 0
	utf4BytesChars := 0
	sjisBytesLeft := 0
	sjisKatakanaChars := 0
	sjisCurKatakanaWordLength := 0
	sjisCurDoubleBytesWordLength := 0
	sjisMaxKatakanaWordLength := 0
	sjisMaxDoubleBytesWordLength := 0
	isoHighOther := 0

	utf8bom := len(bytes) > 3 &&
		bytes[0] == 0xEF &&
		bytes[1] == 0xBB &&
		bytes[2] == 0xBF

	for i := 0; i < length && (canBeISO88591 || canBeShiftJIS || canBeUTF8); i++ {

		value := bytes[i] & 0xFF

		if canBeUTF8 {
			if utf8BytesLeft > 0 {
				if (value & 0x80) == 0 {
					canBeUTF8 = false
				} else {
					utf8BytesLeft--
				}
			} else if (value & 0x80) != 0 {
				if (value & 0x40) == 0 {
					canBeUTF8 = false
				} else {
					utf8BytesLeft++
					if (value & 0x20) == 0 {
						utf2BytesChars++
					} else {
						utf8BytesLeft++
						if (value & 0x10) == 0 {
							utf3BytesChars++
						} else {
							utf8BytesLeft++
							if (value & 0x08) == 0 {
								utf4BytesChars++
							} else {
								canBeUTF8 = false
							}
						}
					}
				}
			}
		}

		if canBeISO88591 {
			if value > 0x7F && value < 0xA0 {
				canBeISO88591 = false
			} else if value > 0x9F && (value < 0xC0 || value == 0xD7 || value == 0xF7) {
				isoHighOther++
			}
		}

		if canBeShiftJIS {
			if sjisBytesLeft > 0 {
				if value < 0x40 || value == 0x7F || value > 0xFC {
					canBeShiftJIS = false
				} else {
					sjisBytesLeft--
				}
			} else if value == 0x80 || value == 0xA0 || value > 0xEF {
				canBeShiftJIS = false
			} else if value > 0xA0 && value < 0xE0 {
				sjisKatakanaChars++
				sjisCurDoubleBytesWordLength = 0
				sjisCurKatakanaWordLength++
				if sjisCurKatakanaWordLength > sjisMaxKatakanaWordLength {
					sjisMaxKatakanaWordLength = sjisCurKatakanaWordLength
				}
			} else if value > 0x7F {
				sjisBytesLeft++
				//sjisDoubleBytesChars++;
				sjisCurKatakanaWordLength = 0
				sjisCurDoubleBytesWordLength++
				if sjisCurDoubleBytesWordLength > sjisMaxDoubleBytesWordLength {
					sjisMaxDoubleBytesWordLength = sjisCurDoubleBytesWordLength
				}
			} else {
				//sjisLowChars++;
				sjisCurKatakanaWordLength = 0
				sjisCurDoubleBytesWordLength = 0
			}
		}
	}

	if canBeUTF8 && utf8BytesLeft > 0 {
		canBeUTF8 = false
	}
	if canBeShiftJIS && sjisBytesLeft > 0 {
		canBeShiftJIS = false
	}

	if canBeUTF8 && (utf8bom || utf2BytesChars+utf3BytesChars+utf4BytesChars > 0) {
		return unicode.UTF8, nil
	}
	if canBeShiftJIS && (sjisMaxKatakanaWordLength >= 3 || sjisMaxDoubleBytesWordLength >= 3) {
		return StringUtils_SHIFT_JIS_CHARSET, nil
	}
	if canBeISO88591 && canBeShiftJIS {
		if (sjisMaxKatakanaWordLength == 2 && sjisKatakanaChars == 2) || isoHighOther*10 >= length {
			return StringUtils_SHIFT_JIS_CHARSET, nil
		}
		return charmap.ISO8859_1, nil
	}

	if canBeISO88591 {
		return charmap.ISO8859_1, nil
	}
	if canBeShiftJIS {
		return StringUtils_SHIFT_JIS_CHARSET, nil
	}
	if canBeUTF8 {
		return unicode.UTF8, nil
	}
	return StringUtils_PLATFORM_DEFAULT_ENCODING, nil
}