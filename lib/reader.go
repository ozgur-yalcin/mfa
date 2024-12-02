package lib

type Reader interface {
	DecodeWithoutHints(image *BinaryBitmap) (*Result, error)
	Decode(image *BinaryBitmap, hints map[DecodeHintType]interface{}) (*Result, error)
	Reset()
}
