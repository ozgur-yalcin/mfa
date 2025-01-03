package decoder

import (
	"github.com/ozgur-yalcin/mfa/lib"
	"github.com/ozgur-yalcin/mfa/lib/common"
	"github.com/ozgur-yalcin/mfa/lib/common/reedsolomon"
)

type Decoder struct {
	rsDecoder *reedsolomon.ReedSolomonDecoder
}

func NewDecoder() *Decoder {
	return &Decoder{
		rsDecoder: reedsolomon.NewReedSolomonDecoder(reedsolomon.GenericGF_QR_CODE_FIELD_256),
	}
}

func (this *Decoder) DecodeBoolMapWithoutHint(image [][]bool) (*common.DecoderResult, error) {
	return this.DecodeBoolMap(image, nil)
}

func (this *Decoder) DecodeBoolMap(image [][]bool, hints map[lib.DecodeHintType]interface{}) (*common.DecoderResult, error) {
	bits, e := lib.ParseBoolMapToBitMatrix(image)
	if e != nil {
		return nil, e
	}
	return this.Decode(bits, hints)
}

func (this *Decoder) DecodeWithoutHint(bits *lib.BitMatrix) (*common.DecoderResult, error) {
	return this.Decode(bits, nil)
}

func (this *Decoder) Decode(bits *lib.BitMatrix, hints map[lib.DecodeHintType]interface{}) (*common.DecoderResult, error) {

	parser, e := NewBitMatrixParser(bits)
	if e != nil {
		return nil, lib.WrapFormatException(e)
	}
	var fece lib.ReaderException

	result, e := this.decode(parser, hints)
	if e == nil {
		return result, nil
	}

	switch e.(type) {
	case lib.FormatException, lib.ChecksumException:
		fece = e.(lib.ReaderException)
	}

	parser.Remask()

	parser.SetMirror(true)

	_, e = parser.ReadVersion()

	if e == nil {
		_, e = parser.ReadFormatInformation()
	}

	if e == nil {
		parser.Mirror()
	}

	if e == nil {
		result, e = this.decode(parser, hints)
	}

	if e == nil {
		result.SetOther(NewQRCodeDecoderMetaData(true))
		return result, nil
	}

	switch e.(type) {
	case lib.FormatException, lib.ChecksumException:
		return nil, fece
	default:
		return nil, e
	}
}

func (this *Decoder) decode(parser *BitMatrixParser, hints map[lib.DecodeHintType]interface{}) (*common.DecoderResult, error) {
	version, e := parser.ReadVersion()
	if e != nil {
		return nil, lib.WrapFormatException(e)
	}
	formatinfo, e := parser.ReadFormatInformation()
	if e != nil {
		return nil, lib.WrapFormatException(e)
	}
	ecLevel := formatinfo.GetErrorCorrectionLevel()

	codewords, e := parser.ReadCodewords()
	if e != nil {
		return nil, lib.WrapFormatException(e)
	}
	dataBlocks, e := DataBlock_GetDataBlocks(codewords, version, ecLevel)
	if e != nil {
		return nil, lib.WrapFormatException(e)
	}

	totalBytes := 0
	for _, dataBlock := range dataBlocks {
		totalBytes += dataBlock.GetNumDataCodewords()
	}
	resultBytes := make([]byte, totalBytes)
	resultOffset := 0

	for _, dataBlock := range dataBlocks {
		codewordBytes := dataBlock.GetCodewords()
		numDataCodewords := dataBlock.GetNumDataCodewords()
		e := this.correctErrors(codewordBytes, numDataCodewords)
		if e != nil {
			return nil, e
		}
		for i := 0; i < numDataCodewords; i++ {
			resultBytes[resultOffset] = codewordBytes[i]
			resultOffset++
		}
	}

	return DecodedBitStreamParser_Decode(resultBytes, version, ecLevel, hints)
}

func (this *Decoder) correctErrors(codewordBytes []byte, numDataCodewords int) error {
	numCodewords := len(codewordBytes)
	codewordsInts := make([]int, numCodewords)
	for i := 0; i < numCodewords; i++ {
		codewordsInts[i] = int(codewordBytes[i] & 0xFF)
	}

	e := this.rsDecoder.Decode(codewordsInts, numCodewords-numDataCodewords)
	if e != nil {
		return lib.WrapChecksumException(e)
	}
	for i := 0; i < numDataCodewords; i++ {
		codewordBytes[i] = byte(codewordsInts[i])
	}
	return nil
}
