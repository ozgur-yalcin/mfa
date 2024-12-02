package qrcode

import (
	"strconv"

	"github.com/ozgur-yalcin/mfa/lib"
	"github.com/ozgur-yalcin/mfa/lib/qrcode/decoder"
	"github.com/ozgur-yalcin/mfa/lib/qrcode/encoder"
)

const (
	qrcodeWriter_QUIET_ZONE_SIZE = 4
)

type QRCodeWriter struct{}

func NewQRCodeWriter() *QRCodeWriter {
	return &QRCodeWriter{}
}

func (this *QRCodeWriter) EncodeWithoutHint(
	contents string, format lib.BarcodeFormat, width, height int) (*lib.BitMatrix, error) {
	return this.Encode(contents, format, width, height, nil)
}

func (this *QRCodeWriter) Encode(
	contents string, format lib.BarcodeFormat, width, height int,
	hints map[lib.EncodeHintType]interface{}) (*lib.BitMatrix, error) {

	if len(contents) == 0 {
		return nil, lib.NewWriterException("IllegalArgumentException: Found empty contents")
	}

	if format != lib.BarcodeFormat_QR_CODE {
		return nil, lib.NewWriterException(
			"IllegalArgumentException: Can only encode QR_CODE, but got %v", format)
	}

	if width < 0 || height < 0 {
		return nil, lib.NewWriterException(
			"IllegalArgumentException: Requested dimensions are too small: %vx%v", width, height)
	}

	errorCorrectionLevel := decoder.ErrorCorrectionLevel_L
	quietZone := qrcodeWriter_QUIET_ZONE_SIZE
	if hints != nil {
		if ec, ok := hints[lib.EncodeHintType_ERROR_CORRECTION]; ok {
			if ecl, ok := ec.(decoder.ErrorCorrectionLevel); ok {
				errorCorrectionLevel = ecl
			} else if str, ok := ec.(string); ok {
				ecl, e := decoder.ErrorCorrectionLevel_ValueOf(str)
				if e != nil {
					return nil, lib.NewWriterException("EncodeHintType_ERROR_CORRECTION: %w", e)
				}
				errorCorrectionLevel = ecl
			} else {
				return nil, lib.NewWriterException(
					"IllegalArgumentException: EncodeHintType_ERROR_CORRECTION %v", ec)
			}
		}
		if m, ok := hints[lib.EncodeHintType_MARGIN]; ok {
			if qz, ok := m.(int); ok {
				quietZone = qz
			} else if str, ok := m.(string); ok {
				qz, e := strconv.Atoi(str)
				if e != nil {
					return nil, lib.NewWriterException("EncodeHintType_MARGIN = \"%v\": %w", m, e)
				}
				quietZone = qz
			} else {
				return nil, lib.NewWriterException(
					"IllegalArgumentException: EncodeHintType_MARGIN %v", m)
			}
		}
	}

	code, e := encoder.Encoder_encode(contents, errorCorrectionLevel, hints)
	if e != nil {
		return nil, e
	}
	return renderResult(code, width, height, quietZone)
}

func renderResult(code *encoder.QRCode, width, height, quietZone int) (*lib.BitMatrix, error) {
	input := code.GetMatrix()
	if input == nil {
		return nil, lib.NewWriterException("IllegalStateException")
	}
	inputWidth := input.GetWidth()
	inputHeight := input.GetHeight()
	qrWidth := inputWidth + (quietZone * 2)
	qrHeight := inputHeight + (quietZone * 2)
	outputWidth := qrWidth
	if outputWidth < width {
		outputWidth = width
	}
	outputHeight := qrHeight
	if outputHeight < height {
		outputHeight = height
	}

	multiple := outputWidth / qrWidth
	if h := outputHeight / qrHeight; multiple > h {
		multiple = h
	}
	leftPadding := (outputWidth - (inputWidth * multiple)) / 2
	topPadding := (outputHeight - (inputHeight * multiple)) / 2

	output, e := lib.NewBitMatrix(outputWidth, outputHeight)
	if e != nil {
		return nil, lib.WrapWriterException(e)
	}

	for inputY, outputY := 0, topPadding; inputY < inputHeight; inputY, outputY = inputY+1, outputY+multiple {
		for inputX, outputX := 0, leftPadding; inputX < inputWidth; inputX, outputX = inputX+1, outputX+multiple {
			if input.Get(inputX, inputY) == 1 {
				output.SetRegion(outputX, outputY, multiple, multiple)
			}
		}
	}

	return output, nil
}
