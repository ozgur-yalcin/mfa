package qrcode

import (
	"sort"

	"github.com/ozgur-yalcin/mfa/lib"
	"github.com/ozgur-yalcin/mfa/lib/multi"
	"github.com/ozgur-yalcin/mfa/lib/multi/qrcode/detector"
	"github.com/ozgur-yalcin/mfa/lib/qrcode"
	"github.com/ozgur-yalcin/mfa/lib/qrcode/decoder"
)

var (
	noPoints = []lib.ResultPoint{}
)

type QRCodeMultiReader struct {
	*qrcode.QRCodeReader
}

func NewQRCodeMultiReader() multi.MultipleBarcodeReader {
	return &QRCodeMultiReader{
		qrcode.NewQRCodeReader().(*qrcode.QRCodeReader),
	}
}

func (this *QRCodeMultiReader) DecodeMultipleWithoutHint(image *lib.BinaryBitmap) ([]*lib.Result, error) {
	return this.DecodeMultiple(image, nil)
}

func (this *QRCodeMultiReader) DecodeMultiple(image *lib.BinaryBitmap, hints map[lib.DecodeHintType]interface{}) ([]*lib.Result, error) {
	results := make([]*lib.Result, 0)
	matrix, e := image.GetBlackMatrix()
	if e != nil {
		return results, e
	}
	detectorResults, e := detector.NewMultiDetector(matrix).DetectMulti(hints)
	if e != nil {
		return results, e
	}
	for _, detectorResult := range detectorResults {
		decoderResult, e := this.GetDecoder().Decode(detectorResult.GetBits(), hints)
		if e != nil {
			if _, ok := e.(lib.ReaderException); ok {
				continue
			} else {
				return results, e
			}
		}
		points := detectorResult.GetPoints()
		if metadata, ok := decoderResult.GetOther().(*decoder.QRCodeDecoderMetaData); ok {
			metadata.ApplyMirroredCorrection(points)
		}
		result := lib.NewResult(decoderResult.GetText(), decoderResult.GetRawBytes(), points,
			lib.BarcodeFormat_QR_CODE)
		byteSegments := decoderResult.GetByteSegments()
		if byteSegments != nil {
			result.PutMetadata(lib.ResultMetadataType_BYTE_SEGMENTS, byteSegments)
		}
		ecLevel := decoderResult.GetECLevel()
		if ecLevel != "" {
			result.PutMetadata(lib.ResultMetadataType_ERROR_CORRECTION_LEVEL, ecLevel)
		}
		if decoderResult.HasStructuredAppend() {
			result.PutMetadata(lib.ResultMetadataType_STRUCTURED_APPEND_SEQUENCE,
				decoderResult.GetStructuredAppendSequenceNumber())
			result.PutMetadata(lib.ResultMetadataType_STRUCTURED_APPEND_PARITY,
				decoderResult.GetStructuredAppendParity())
		}
		results = append(results, result)
	}
	if len(results) != 0 {
		results = processStructuredAppend(results)
	}
	return results, nil
}

func processStructuredAppend(results []*lib.Result) []*lib.Result {
	hasSA := false
	for _, result := range results {
		metadata := result.GetResultMetadata()
		if _, ok := metadata[lib.ResultMetadataType_STRUCTURED_APPEND_SEQUENCE]; ok {
			hasSA = true
			break
		}
	}
	if !hasSA {
		return results
	}
	newResults := make([]*lib.Result, 0)
	saResults := make([]*lib.Result, 0)
	for _, result := range results {
		metadata := result.GetResultMetadata()
		if _, ok := metadata[lib.ResultMetadataType_STRUCTURED_APPEND_SEQUENCE]; ok {
			saResults = append(saResults, result)
		} else {
			newResults = append(newResults, result)
		}
	}
	sort.Slice(saResults, newSAComparator(saResults))
	concatedText := make([]byte, 0)
	rawBytesLen := 0
	byteSegmentLength := 0
	for _, saResult := range saResults {
		concatedText = append(concatedText, []byte(saResult.GetText())...)
		rawBytesLen += len(saResult.GetRawBytes())
		metadata := saResult.GetResultMetadata()
		if byteSegments, ok := metadata[lib.ResultMetadataType_BYTE_SEGMENTS].([][]byte); ok {
			for _, segment := range byteSegments {
				byteSegmentLength += len(segment)
			}
		}
	}
	newRawBytes := make([]byte, rawBytesLen)
	newByteSegment := make([]byte, byteSegmentLength)
	newRawBytesIndex := 0
	byteSegmentIndex := 0
	for _, saResult := range saResults {
		copy(newRawBytes[newRawBytesIndex:], saResult.GetRawBytes())
		newRawBytesIndex += len(saResult.GetRawBytes())

		metadata := saResult.GetResultMetadata()
		if byteSegments, ok := metadata[lib.ResultMetadataType_BYTE_SEGMENTS].([][]byte); ok {
			for _, segment := range byteSegments {
				copy(newByteSegment[byteSegmentIndex:], segment)
				byteSegmentIndex += len(segment)
			}
		}
	}
	newResult := lib.NewResult(string(concatedText), newRawBytes, noPoints, lib.BarcodeFormat_QR_CODE)
	if byteSegmentLength > 0 {
		byteSegmentList := [][]byte{newByteSegment}
		newResult.PutMetadata(lib.ResultMetadataType_BYTE_SEGMENTS, byteSegmentList)
	}
	newResults = append(newResults, newResult)
	return newResults
}

func newSAComparator(results []*lib.Result) func(int, int) bool {
	return func(a, b int) bool {
		aNumber, _ := results[a].GetResultMetadata()[lib.ResultMetadataType_STRUCTURED_APPEND_SEQUENCE].(int)
		bNumber, _ := results[b].GetResultMetadata()[lib.ResultMetadataType_STRUCTURED_APPEND_SEQUENCE].(int)
		return aNumber < bNumber
	}
}
