//
//	goray/core/color/color.go
//	goray
//
//	Created by Ross Light on 2010-05-22.
//

/*
	The color package provides abstracted color.

	This interface specifically differes from the image.Color interfaces because
	the render uses floating point math.  Also, colors in this package are not
	clamped to [0, 1], they are clamped to [0, Inf).
*/
package color

import (
	"fmt"
	"image"
	"math"
)

func quantizeComponent(f float) uint32 {
	temp := uint32(f * math.MaxUint16)
	if temp > math.MaxUint16 {
		return math.MaxUint16
	} else if temp < 0 {
		return 0
	}
	return temp
}

// Alpha defines anything that has an alpha channel.
type Alpha interface {
	GetA() float
}

// Color defines anything that has red, green, and blue channels.
type Color interface {
	GetR() float
	GetG() float
	GetB() float
}

// AlphaColor defines anything that has red, green, blue, and alpha channels.
type AlphaColor interface {
	Color
	Alpha
}

// Gray defines a grayscale color.  It fulfills the Color interface.
type Gray float

func (g Gray) GetR() float { return float(g) }
func (g Gray) GetG() float { return float(g) }
func (g Gray) GetB() float { return float(g) }

func (col Gray) RGBA() (r, g, b, a uint32) {
	r = quantizeComponent(float(col))
	g = r
	b = r
	a = math.MaxUint32
	return
}

func (g Gray) String() string {
	return fmt.Sprintf("Gray(%.3f)", float(g))
}

// RGB defines a color that has red, green, and blue channels.  It fulfills the Color interface.
type RGB struct {
	R, G, B float
}

func NewRGB(r, g, b float) RGB    { return RGB{r, g, b} }
func DiscardAlpha(c Color) RGB    { return NewRGB(c.GetR(), c.GetG(), c.GetB()) }
func (c *RGB) Init(r, g, b float) { c.R = r; c.G = g; c.B = b }
func (c *RGB) Copy(src Color) {
	c.R = src.GetR()
	c.G = src.GetG()
	c.B = src.GetB()
}

func (c RGB) GetR() float { return c.R }
func (c RGB) GetG() float { return c.G }
func (c RGB) GetB() float { return c.B }

func (c RGB) RGBA() (r, g, b, a uint32) {
	r = quantizeComponent(c.R)
	g = quantizeComponent(c.G)
	b = quantizeComponent(c.B)
	a = math.MaxUint32
	return
}

func (c RGB) String() string {
	return fmt.Sprintf("RGB(%.3f, %.3f, %.3f)", c.R, c.G, c.B)
}

// RGBA colors

type RGBA struct {
	RGB
	A float
}

func NewRGBA(r, g, b, a float) RGBA   { return RGBA{NewRGB(r, g, b), a} }
func (c *RGBA) Init(r, g, b, a float) { c.RGB.Init(r, g, b); c.A = a }
func (c *RGBA) Copy(src AlphaColor) {
	c.RGB.Copy(src)
	c.A = src.GetA()
}

// NewRGBAFromColor creates an RGBA value from a color and an alpha value.
func NewRGBAFromColor(c Color, a float) RGBA {
	newColor := RGBA{A: a}
	newColor.RGB.Copy(c)
	return newColor
}

func (c RGBA) GetA() float { return c.A }

func (c RGBA) RGBA() (r, g, b, a uint32) {
	r, g, b, a = c.AlphaPremultiply().RGB.RGBA()
	a = quantizeComponent(c.A)
	return
}

func (c RGBA) String() string {
	return fmt.Sprintf("RGBA(%.3f, %.3f, %.3f, %.3f)", c.R, c.G, c.B, c.A)
}

func (c RGBA) AlphaPremultiply() RGBA {
	return NewRGBA(c.R*c.A, c.G*c.A, c.B*c.A, c.A)
}

// Predefined colors
var (
	Black Color = Gray(0)
	White Color = Gray(1)

	Red   Color = RGB{1, 0, 0}
	Green Color = RGB{0, 1, 0}
	Blue  Color = RGB{0, 0, 1}

	Cyan    Color = RGB{0, 1, 1}
	Yellow  Color = RGB{1, 1, 0}
	Magenta Color = RGB{1, 0, 1}
)

// Operations

func toGorayColor(col image.Color) image.Color {
	switch col.(type) {
	case RGB, RGBA, Gray:
		return col
	}
	r, g, b, a := col.RGBA()
	return NewRGBA(
		float(r)/math.MaxUint32,
		float(g)/math.MaxUint32,
		float(b)/math.MaxUint32,
		float(a)/math.MaxUint32,
	)
}

// Model is the color model for the renderer.
var Model image.ColorModel = image.ColorModelFunc(toGorayColor)
