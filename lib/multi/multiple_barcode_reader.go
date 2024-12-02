package multi

import (
	"github.com/ozgur-yalcin/mfa/lib"
)

type MultipleBarcodeReader interface {
	DecodeMultipleWithoutHint(image *lib.BinaryBitmap) ([]*lib.Result, error)

	DecodeMultiple(image *lib.BinaryBitmap, hints map[lib.DecodeHintType]interface{}) ([]*lib.Result, error)
}
