package scan

import errors "golang.org/x/xerrors"

type LuminanceSource interface {
	GetRow(y int, row []byte) ([]byte, error)
	GetMatrix() []byte
	GetWidth() int
	GetHeight() int
	IsCropSupported() bool
	Crop(left, top, width, height int) (LuminanceSource, error)
	IsRotateSupported() bool
	Invert() LuminanceSource
	RotateCounterClockwise() (LuminanceSource, error)
	RotateCounterClockwise45() (LuminanceSource, error)
	String() string
}

type LuminanceSourceBase struct {
	Width  int
	Height int
}

func (this *LuminanceSourceBase) GetWidth() int {
	return this.Width
}

func (this *LuminanceSourceBase) GetHeight() int {
	return this.Height
}

func (this *LuminanceSourceBase) IsCropSupported() bool {
	return false
}

func (this *LuminanceSourceBase) Crop(left, top, width, height int) (LuminanceSource, error) {
	return nil, errors.New("UnsupportedOperationException: This luminance source does not support cropping")
}

func (this *LuminanceSourceBase) IsRotateSupported() bool {
	return false
}

func (this *LuminanceSourceBase) RotateCounterClockwise() (LuminanceSource, error) {
	return nil, errors.New("UnsupportedOperationException: This luminance source does not support rotation by 90 degrees")
}

func (this *LuminanceSourceBase) RotateCounterClockwise45() (LuminanceSource, error) {
	return nil, errors.New("UnsupportedOperationException: This luminance source does not support rotation by 45 degrees")
}

func LuminanceSourceInvert(this LuminanceSource) LuminanceSource {
	return NewInvertedLuminanceSource(this)
}

func LuminanceSourceString(this LuminanceSource) string {
	width := this.GetWidth()
	height := this.GetHeight()
	row := make([]byte, width)
	result := make([]byte, 0, height*(width+1))

	for y := 0; y < height; y++ {
		row, _ = this.GetRow(y, row)
		for x := 0; x < width; x++ {
			luminance := row[x] & 0xFF
			var c byte
			if luminance < 0x40 {
				c = '#'
			} else if luminance < 0x80 {
				c = '+'
			} else if luminance < 0xC0 {
				c = '.'
			} else {
				c = ' '
			}
			result = append(result, c)
		}
		result = append(result, '\n')
	}
	return string(result)
}
