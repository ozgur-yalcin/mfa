package scan

import errors "golang.org/x/xerrors"

type RGBLuminanceSource struct {
	LuminanceSourceBase
	luminances []byte
	dataWidth  int
	dataHeight int
	left       int
	top        int
}

func NewRGBLuminanceSource(width, height int, pixels []int) LuminanceSource {
	dataWidth := width
	dataHeight := height
	left := 0
	top := 0

	//
	size := width * height
	luminances := make([]byte, size)
	for offset := 0; offset < size; offset++ {
		pixel := pixels[offset]
		r := (pixel >> 16) & 0xff
		g2 := (pixel >> 7) & 0x1fe
		b := pixel & 0xff
		luminances[offset] = byte((r + g2 + b) / 4)
	}

	return &RGBLuminanceSource{
		LuminanceSourceBase{width, height},
		luminances,
		dataWidth,
		dataHeight,
		left,
		top,
	}
}

func (this *RGBLuminanceSource) GetRow(y int, row []byte) ([]byte, error) {
	if y < 0 || y >= this.GetHeight() {
		return row, errors.Errorf("IllegalArgumentException: Requested row is outside the image: %d", y)
	}
	width := this.GetWidth()
	if row == nil || len(row) < width {
		row = make([]byte, width)
	}
	offset := (y+this.top)*this.dataWidth + this.left
	copy(row, this.luminances[offset:offset+width])
	return row, nil
}

func (this *RGBLuminanceSource) GetMatrix() []byte {
	width := this.GetWidth()
	height := this.GetHeight()

	if width == this.dataWidth && height == this.dataHeight {
		return this.luminances
	}

	area := width * height
	matrix := make([]byte, area)
	inputOffset := this.top*this.dataWidth + this.left

	if width == this.dataWidth {
		copy(matrix, this.luminances[inputOffset:inputOffset+area])
		return matrix
	}

	for y := 0; y < height; y++ {
		outputOffset := y * width
		copy(matrix[outputOffset:outputOffset+width], this.luminances[inputOffset:inputOffset+width])
		inputOffset += this.dataWidth
	}
	return matrix
}

func (this *RGBLuminanceSource) IsCropSupported() bool {
	return true
}

func (this *RGBLuminanceSource) Crop(left, top, width, height int) (LuminanceSource, error) {
	if left+width > this.dataWidth || top+height > this.dataHeight {
		return nil, errors.New("IllegalArgumentException: Crop rectangle does not fit within image data")
	}
	return &RGBLuminanceSource{
		LuminanceSourceBase: LuminanceSourceBase{width, height},
		luminances:          this.luminances,
		dataWidth:           this.dataWidth,
		dataHeight:          this.dataHeight,
		left:                this.left + left,
		top:                 this.top + top,
	}, nil
}

func (this *RGBLuminanceSource) Invert() LuminanceSource {
	return LuminanceSourceInvert(this)
}

func (this *RGBLuminanceSource) String() string {
	return LuminanceSourceString(this)
}
