package lib

type BarcodeFormat int
type BarcodeFormats []BarcodeFormat

const (
	BarcodeFormat_AZTEC = BarcodeFormat(iota)

	BarcodeFormat_CODABAR

	BarcodeFormat_CODE_39

	BarcodeFormat_CODE_93

	BarcodeFormat_CODE_128

	BarcodeFormat_DATA_MATRIX

	BarcodeFormat_EAN_8

	BarcodeFormat_EAN_13

	BarcodeFormat_ITF

	BarcodeFormat_MAXICODE

	BarcodeFormat_PDF_417

	BarcodeFormat_QR_CODE

	BarcodeFormat_RSS_14

	BarcodeFormat_RSS_EXPANDED

	BarcodeFormat_UPC_A

	BarcodeFormat_UPC_E

	BarcodeFormat_UPC_EAN_EXTENSION
)

func (f BarcodeFormat) String() string {
	switch f {
	case BarcodeFormat_AZTEC:
		return "AZTEC"
	case BarcodeFormat_CODABAR:
		return "CODABAR"
	case BarcodeFormat_CODE_39:
		return "CODE_39"
	case BarcodeFormat_CODE_93:
		return "CODE_93"
	case BarcodeFormat_CODE_128:
		return "CODE_128"
	case BarcodeFormat_DATA_MATRIX:
		return "DATA_MATRIX"
	case BarcodeFormat_EAN_8:
		return "EAN_8"
	case BarcodeFormat_EAN_13:
		return "EAN_13"
	case BarcodeFormat_ITF:
		return "ITF"
	case BarcodeFormat_MAXICODE:
		return "MAXICODE"
	case BarcodeFormat_PDF_417:
		return "PDF_417"
	case BarcodeFormat_QR_CODE:
		return "QR_CODE"
	case BarcodeFormat_RSS_14:
		return "RSS_14"
	case BarcodeFormat_RSS_EXPANDED:
		return "RSS_EXPANDED"
	case BarcodeFormat_UPC_A:
		return "UPC_A"
	case BarcodeFormat_UPC_E:
		return "UPC_E"
	case BarcodeFormat_UPC_EAN_EXTENSION:
		return "UPC_EAN_EXTENSION"
	default:
		return "unknown format"
	}
}

func (barcodes BarcodeFormats) Contains(c BarcodeFormat) bool {
	for _, bc := range barcodes {
		if bc == c {
			return true
		}
	}
	return false
}
