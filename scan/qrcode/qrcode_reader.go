package qrcode

import (
	"strconv"

	"github.com/ozgur-yalcin/mfa/scan"
	"github.com/ozgur-yalcin/mfa/scan/common"
	"github.com/ozgur-yalcin/mfa/scan/common/util"
	"github.com/ozgur-yalcin/mfa/scan/qrcode/decoder"
	"github.com/ozgur-yalcin/mfa/scan/qrcode/detector"
)

type QRCodeReader struct {
	decoder *decoder.Decoder
}

func NewQRCodeReader() scan.Reader {
	return &QRCodeReader{
		decoder.NewDecoder(),
	}
}

func (this *QRCodeReader) GetDecoder() *decoder.Decoder {
	return this.decoder
}

func (this *QRCodeReader) DecodeWithoutHints(image *scan.BinaryBitmap) (*scan.Result, error) {
	return this.Decode(image, nil)
}

func (this *QRCodeReader) Decode(image *scan.BinaryBitmap, hints map[scan.DecodeHintType]interface{}) (*scan.Result, error) {
	var decoderResult *common.DecoderResult
	var points []scan.ResultPoint

	blackMatrix, e := image.GetBlackMatrix()
	if e != nil {
		return nil, e
	}
	if _, ok := hints[scan.DecodeHintType_PURE_BARCODE]; ok {
		bits, e := this.extractPureBits(blackMatrix)
		if e != nil {
			return nil, e
		}
		decoderResult, e = this.decoder.Decode(bits, hints)
		if e != nil {
			return nil, e
		}
		points = []scan.ResultPoint{}
	} else {
		detectorResult, e := detector.NewDetector(blackMatrix).Detect(hints)
		if e != nil {
			return nil, e
		}
		decoderResult, e = this.decoder.Decode(detectorResult.GetBits(), hints)
		if e != nil {
			return nil, e
		}
		points = detectorResult.GetPoints()
	}

	if metadata, ok := decoderResult.GetOther().(*decoder.QRCodeDecoderMetaData); ok {
		metadata.ApplyMirroredCorrection(points)
	}

	result := scan.NewResult(decoderResult.GetText(), decoderResult.GetRawBytes(), points, scan.BarcodeFormat_QR_CODE)
	byteSegments := decoderResult.GetByteSegments()
	if len(byteSegments) > 0 {
		result.PutMetadata(scan.ResultMetadataType_BYTE_SEGMENTS, byteSegments)
	}
	ecLevel := decoderResult.GetECLevel()
	if ecLevel != "" {
		result.PutMetadata(scan.ResultMetadataType_ERROR_CORRECTION_LEVEL, ecLevel)
	}
	if decoderResult.HasStructuredAppend() {
		result.PutMetadata(
			scan.ResultMetadataType_STRUCTURED_APPEND_SEQUENCE,
			decoderResult.GetStructuredAppendSequenceNumber())
		result.PutMetadata(
			scan.ResultMetadataType_STRUCTURED_APPEND_PARITY,
			decoderResult.GetStructuredAppendParity())
	}
	result.PutMetadata(
		scan.ResultMetadataType_SYMBOLOGY_IDENTIFIER, "]Q"+strconv.Itoa(decoderResult.GetSymbologyModifier()))
	return result, nil
}

func (this *QRCodeReader) Reset() {
}

func (this *QRCodeReader) extractPureBits(image *scan.BitMatrix) (*scan.BitMatrix, error) {

	leftTopBlack := image.GetTopLeftOnBit()
	rightBottomBlack := image.GetBottomRightOnBit()
	if leftTopBlack == nil || rightBottomBlack == nil {
		return nil, scan.NewNotFoundException()
	}

	moduleSize, e := this.moduleSize(leftTopBlack, image)
	if e != nil {
		return nil, e
	}

	top := leftTopBlack[1]
	bottom := rightBottomBlack[1]
	left := leftTopBlack[0]
	right := rightBottomBlack[0]

	if left >= right || top >= bottom {
		return nil, scan.NewNotFoundException(
			"(left,right)=(%v,%v), (top,bottom)=(%v,%v)", left, right, top, bottom)
	}

	if bottom-top != right-left {
		right = left + (bottom - top)
		if right >= image.GetWidth() {
			return nil, scan.NewNotFoundException("right = %v, width = %v", right, image.GetWidth())
		}
	}

	matrixWidth := util.MathUtils_Round(float64(right-left+1) / moduleSize)
	matrixHeight := util.MathUtils_Round(float64(bottom-top+1) / moduleSize)
	if matrixWidth <= 0 || matrixHeight <= 0 {
		return nil, scan.NewNotFoundException("matrixWidth/Height = %v, %v", matrixWidth, matrixHeight)
	}
	if matrixHeight != matrixWidth {
		return nil, scan.NewNotFoundException("matrixWidth/Height = %v, %v", matrixWidth, matrixHeight)
	}

	nudge := int(moduleSize / 2.0)
	top += nudge
	left += nudge

	nudgedTooFarRight := left + int(float64(matrixWidth-1)*moduleSize) - right
	if nudgedTooFarRight > 0 {
		if nudgedTooFarRight > nudge {
			return nil, scan.NewNotFoundException("Neither way fits")
		}
		left -= nudgedTooFarRight
	}
	nudgedTooFarDown := top + int(float64(matrixHeight-1)*moduleSize) - bottom
	if nudgedTooFarDown > 0 {
		if nudgedTooFarDown > nudge {
			return nil, scan.NewNotFoundException("Neither way fits")
		}
		top -= nudgedTooFarDown
	}

	bits, _ := scan.NewBitMatrix(matrixWidth, matrixHeight)
	for y := 0; y < matrixHeight; y++ {
		iOffset := top + int(float64(y)*moduleSize)
		for x := 0; x < matrixWidth; x++ {
			if image.Get(left+int(float64(x)*moduleSize), iOffset) {
				bits.Set(x, y)
			}
		}
	}
	return bits, nil
}

func (this *QRCodeReader) moduleSize(leftTopBlack []int, image *scan.BitMatrix) (float64, error) {
	height := image.GetHeight()
	width := image.GetWidth()
	x := leftTopBlack[0]
	y := leftTopBlack[1]
	inBlack := true
	transitions := 0
	for x < width && y < height {
		if inBlack != image.Get(x, y) {
			transitions++
			if transitions == 5 {
				break
			}
			inBlack = !inBlack
		}
		x++
		y++
	}
	if x == width || y == height {
		return 0, scan.NewNotFoundException()
	}
	return float64(x-leftTopBlack[0]) / 7.0, nil
}
