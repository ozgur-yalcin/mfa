package lib

type Binarizer interface {
	GetLuminanceSource() LuminanceSource
	GetBlackRow(y int, row *BitArray) (*BitArray, error)
	GetBlackMatrix() (*BitMatrix, error)
	CreateBinarizer(source LuminanceSource) Binarizer
	GetWidth() int
	GetHeight() int
}
