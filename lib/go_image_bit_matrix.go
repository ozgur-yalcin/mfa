package lib

import (
	"image"
	"image/color"
)

func (img *BitMatrix) ColorModel() color.Model {
	return color.GrayModel
}

func (img *BitMatrix) Bounds() image.Rectangle {
	return image.Rect(0, 0, img.GetWidth(), img.GetHeight())
}

func (img *BitMatrix) At(x, y int) color.Color {
	c := color.Gray{0}
	if !img.Get(x, y) {
		c.Y = 255
	}
	return c
}
