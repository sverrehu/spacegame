package utils

import (
	"math"

	"github.com/gogpu/gg"
)

type Color struct {
	R, G, B float64
}

func NewColor(r, g, b float64) Color {
	return Color{r, g, b}
}

func NewColor255(r, g, b uint8) Color {
	return NewColor(float64(r)/255.0, float64(g)/255.0, float64(b)/255.0)
}

func (c *Color) ToRGBA() gg.RGBA {
	return gg.RGB(c.R, c.G, c.B)
}

func (c *Color) ToRGB255() (r, g, b uint8) {
	return uint8(c.R*255.0) & 0xff, uint8(c.G*255.0) & 0xff, uint8(c.B*255.0) & 0xff
}

func HSVToColor(h, s, v float64) Color {
	var fR, fG, fB float64

	i := math.Floor(h * 6.0)
	f := h*6.0 - i
	p := v * (1.0 - s)
	q := v * (1.0 - s*f)
	t := v * (1.0 - s*(1.0-f))

	switch int(i) % 6 {
	case 0:
		fR, fG, fB = v, t, p
	case 1:
		fR, fG, fB = q, v, p
	case 2:
		fR, fG, fB = p, v, t
	case 3:
		fR, fG, fB = p, q, v
	case 4:
		fR, fG, fB = t, p, v
	case 5:
		fR, fG, fB = v, p, q
	}
	return Color{fR, fG, fB}
}
