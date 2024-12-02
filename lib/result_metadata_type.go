package lib

type ResultMetadataType int

const (
	ResultMetadataType_OTHER = ResultMetadataType(iota)

	ResultMetadataType_ORIENTATION

	ResultMetadataType_BYTE_SEGMENTS

	ResultMetadataType_ERROR_CORRECTION_LEVEL

	ResultMetadataType_ISSUE_NUMBER

	ResultMetadataType_SUGGESTED_PRICE

	ResultMetadataType_POSSIBLE_COUNTRY

	ResultMetadataType_UPC_EAN_EXTENSION

	ResultMetadataType_PDF417_EXTRA_METADATA

	ResultMetadataType_STRUCTURED_APPEND_SEQUENCE

	ResultMetadataType_STRUCTURED_APPEND_PARITY

	ResultMetadataType_SYMBOLOGY_IDENTIFIER
)

func (t ResultMetadataType) String() string {
	switch t {
	case ResultMetadataType_OTHER:
		return "OTHER"
	case ResultMetadataType_ORIENTATION:
		return "ORIENTATION"
	case ResultMetadataType_BYTE_SEGMENTS:
		return "BYTE_SEGMENTS"
	case ResultMetadataType_ERROR_CORRECTION_LEVEL:
		return "ERROR_CORRECTION_LEVEL"
	case ResultMetadataType_ISSUE_NUMBER:
		return "ISSUE_NUMBER"
	case ResultMetadataType_SUGGESTED_PRICE:
		return "SUGGESTED_PRICE"
	case ResultMetadataType_POSSIBLE_COUNTRY:
		return "POSSIBLE_COUNTRY"
	case ResultMetadataType_UPC_EAN_EXTENSION:
		return "UPC_EAN_EXTENSION"
	case ResultMetadataType_PDF417_EXTRA_METADATA:
		return "PDF417_EXTRA_METADATA"
	case ResultMetadataType_STRUCTURED_APPEND_SEQUENCE:
		return "STRUCTURED_APPEND_SEQUENCE"
	case ResultMetadataType_STRUCTURED_APPEND_PARITY:
		return "STRUCTURED_APPEND_PARITY"
	case ResultMetadataType_SYMBOLOGY_IDENTIFIER:
		return "SYMBOLOGY_IDENTIFIER"
	default:
		return "unknown metadata type"
	}
}
