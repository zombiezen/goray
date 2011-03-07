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

func quantizeComponent(f float64) uint32 {
	switch {
	case f > 1:
		return math.MaxUint16
	case f < 0:
		return 0
	}
	return uint32(f * math.MaxUint16)
}

// Alpha defines anything that has an alpha channel.
type Alpha interface {
	GetA() float64
}

// Color defines anything that has red, green, and blue channels.
type Color interface {
	GetR() float64
	GetG() float64
	GetB() float64
}

// AlphaColor defines anything that has red, green, blue, and alpha channels.
type AlphaColor interface {
	Color
	Alpha
}

// Gray defines a grayscale color.  It fulfills the Color interface.
type Gray float64

func (g Gray) GetR() float64 { return float64(g) }
func (g Gray) GetG() float64 { return float64(g) }
func (g Gray) GetB() float64 { return float64(g) }

func (col Gray) RGBA() (r, g, b, a uint32) {
	r = quantizeComponent(float64(col))
	g = r
	b = r
	a = math.MaxUint32
	return
}

func (g Gray) String() string {
	return fmt.Sprintf("Gray(%.3f)", float64(g))
}

// RGB defines a color that has red, green, and blue channels.  It fulfills the Color interface.
type RGB struct {
	R, G, B float64
}

func NewRGB(r, g, b float64) RGB    { return RGB{r, g, b} }
func DiscardAlpha(c Color) RGB      { return NewRGB(c.GetR(), c.GetG(), c.GetB()) }
func (c *RGB) Init(r, g, b float64) { c.R = r; c.G = g; c.B = b }
func (c *RGB) Copy(src Color) {
	c.R = src.GetR()
	c.G = src.GetG()
	c.B = src.GetB()
}

func (c RGB) GetR() float64 { return c.R }
func (c RGB) GetG() float64 { return c.G }
func (c RGB) GetB() float64 { return c.B }

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
	A float64
}

func NewRGBA(r, g, b, a float64) RGBA   { return RGBA{NewRGB(r, g, b), a} }
func (c *RGBA) Init(r, g, b, a float64) { c.RGB.Init(r, g, b); c.A = a }
func (c *RGBA) Copy(src AlphaColor) {
	c.RGB.Copy(src)
	c.A = src.GetA()
}

// NewRGBAFromColor creates an RGBA value from a color and an alpha value.
func NewRGBAFromColor(c Color, a float64) RGBA {
	newColor := RGBA{A: a}
	newColor.RGB.Copy(c)
	return newColor
}

func (c RGBA) GetA() float64 { return c.A }

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
		float64(r)/math.MaxUint32,
		float64(g)/math.MaxUint32,
		float64(b)/math.MaxUint32,
		float64(a)/math.MaxUint32,
	)
}

// Model is the color model for the renderer.
var Model image.ColorModel = image.ColorModelFunc(toGorayColor)
