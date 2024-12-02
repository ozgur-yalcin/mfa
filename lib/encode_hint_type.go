package lib

type EncodeHintType int

const (
	EncodeHintType_ERROR_CORRECTION = iota

	EncodeHintType_CHARACTER_SET

	EncodeHintType_DATA_MATRIX_SHAPE

	EncodeHintType_MIN_SIZE

	EncodeHintType_MAX_SIZE

	EncodeHintType_MARGIN

	EncodeHintType_PDF417_COMPACT

	EncodeHintType_PDF417_COMPACTION

	EncodeHintType_PDF417_DIMENSIONS

	EncodeHintType_AZTEC_LAYERS

	EncodeHintType_QR_VERSION

	EncodeHintType_QR_MASK_PATTERN

	EncodeHintType_GS1_FORMAT

	EncodeHintType_FORCE_CODE_SET
)

func (this EncodeHintType) String() string {
	switch this {
	case EncodeHintType_ERROR_CORRECTION:
		return "ERROR_CORRECTION"
	case EncodeHintType_CHARACTER_SET:
		return "CHARACTER_SET"
	case EncodeHintType_DATA_MATRIX_SHAPE:
		return "DATA_MATRIX_SHAPE"
	case EncodeHintType_MIN_SIZE:
		return "MIN_SIZE"
	case EncodeHintType_MAX_SIZE:
		return "MAX_SIZE"
	case EncodeHintType_MARGIN:
		return "MARGIN"
	case EncodeHintType_PDF417_COMPACT:
		return "PDF417_COMPACT"
	case EncodeHintType_PDF417_COMPACTION:
		return "PDF417_COMPACTION"
	case EncodeHintType_PDF417_DIMENSIONS:
		return "PDF417_DIMENSIONS"
	case EncodeHintType_AZTEC_LAYERS:
		return "AZTEC_LAYERS"
	case EncodeHintType_QR_VERSION:
		return "QR_VERSION"
	case EncodeHintType_QR_MASK_PATTERN:
		return "QR_MASK_PATTERN"
	case EncodeHintType_GS1_FORMAT:
		return "GS1_FORMAT"
	case EncodeHintType_FORCE_CODE_SET:
		return "FORCE_CODE_SET"
	}
	return ""
}
