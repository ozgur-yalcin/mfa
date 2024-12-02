package decoder

import (
	"github.com/ozgur-yalcin/mfa/scan"
	"github.com/ozgur-yalcin/mfa/scan/common"
	"github.com/ozgur-yalcin/mfa/scan/common/reedsolomon"
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

func (this *Decoder) DecodeBoolMap(image [][]bool, hints map[scan.DecodeHintType]interface{}) (*common.DecoderResult, error) {
	bits, e := scan.ParseBoolMapToBitMatrix(image)
	if e != nil {
		return nil, e
	}
	return this.Decode(bits, hints)
}

func (this *Decoder) DecodeWithoutHint(bits *scan.BitMatrix) (*common.DecoderResult, error) {
	return this.Decode(bits, nil)
}

func (this *Decoder) Decode(bits *scan.BitMatrix, hints map[scan.DecodeHintType]interface{}) (*common.DecoderResult, error) {

	parser, e := NewBitMatrixParser(bits)
	if e != nil {
		return nil, scan.WrapFormatException(e)
	}
	var fece scan.ReaderException

	result, e := this.decode(parser, hints)
	if e == nil {
		return result, nil
	}

	switch e.(type) {
	case scan.FormatException, scan.ChecksumException:
		fece = e.(scan.ReaderException)
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
	case scan.FormatException, scan.ChecksumException:
		return nil, fece
	default:
		return nil, e
	}
}

func (this *Decoder) decode(parser *BitMatrixParser, hints map[scan.DecodeHintType]interface{}) (*common.DecoderResult, error) {
	version, e := parser.ReadVersion()
	if e != nil {
		return nil, scan.WrapFormatException(e)
	}
	formatinfo, e := parser.ReadFormatInformation()
	if e != nil {
		return nil, scan.WrapFormatException(e)
	}
	ecLevel := formatinfo.GetErrorCorrectionLevel()

	codewords, e := parser.ReadCodewords()
	if e != nil {
		return nil, scan.WrapFormatException(e)
	}
	dataBlocks, e := DataBlock_GetDataBlocks(codewords, version, ecLevel)
	if e != nil {
		return nil, scan.WrapFormatException(e)
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
		return scan.WrapChecksumException(e)
	}
	for i := 0; i < numDataCodewords; i++ {
		codewordBytes[i] = byte(codewordsInts[i])
	}
	return nil
}
