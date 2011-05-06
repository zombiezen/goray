//
//	goray/core/color/color.go
//	goray
//
//	Created by Ross Light on 2010-05-22.
//

/*
	The color package provides abstracted color.

	This interface specifically differs from the image.Color interfaces because
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
	Alpha() float64
}

// Color defines anything that has red, green, and blue channels.
type Color interface {
	Red() float64
	Green() float64
	Blue() float64
}

// AlphaColor defines anything that has red, green, blue, and alpha channels.
type AlphaColor interface {
	Color
	Alpha
}

// Gray defines a grayscale color.  It fulfills the Color interface.
type Gray float64

func (g Gray) Red() float64   { return float64(g) }
func (g Gray) Green() float64 { return float64(g) }
func (g Gray) Blue() float64  { return float64(g) }

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

func (g Gray) GoString() string {
	return fmt.Sprintf("color.Gray(%#v)", float64(g))
}

// RGB defines a color that has red, green, and blue channels.  It fulfills the Color interface.
type RGB struct {
	R, G, B float64
}

// DiscardAlpha returns a new RGB with the red, green, and blue components of another color.
func DiscardAlpha(c Color) RGB { return RGB{c.Red(), c.Green(), c.Blue()} }

func (c *RGB) Copy(src Color) {
	c.R = src.Red()
	c.G = src.Green()
	c.B = src.Blue()
}

func (c RGB) Red() float64   { return c.R }
func (c RGB) Green() float64 { return c.G }
func (c RGB) Blue() float64  { return c.B }

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

func (c RGB) GoString() string {
	return fmt.Sprintf("color.RGB{%#v, %#v, %#v}", c.R, c.G, c.B)
}

// RGBA colors

// RGBA defines a color with red, green, blue, and alpha channels.  It fulfills the AlphaColor interface.
type RGBA struct {
	R, G, B, A float64
}

func (c *RGBA) Copy(src AlphaColor) {
	c.R = src.Red()
	c.G = src.Green()
	c.B = src.Blue()
	c.A = src.Alpha()
}

func (c RGBA) Red() float64   { return c.R }
func (c RGBA) Green() float64 { return c.G }
func (c RGBA) Blue() float64  { return c.B }
func (c RGBA) Alpha() float64 { return c.A }

// NewRGBAFromColor creates an RGBA value from a color and an alpha value.
func NewRGBAFromColor(c Color, a float64) RGBA {
	return RGBA{c.Red(), c.Green(), c.Blue(), a}
}

func (c1 RGBA) RGBA() (r, g, b, a uint32) {
	c2 := c1.AlphaPremultiply()
	r = quantizeComponent(c2.R)
	g = quantizeComponent(c2.G)
	b = quantizeComponent(c2.B)
	a = quantizeComponent(c2.A)
	return
}

// AlphaPremultiply multiplies the color channels by the alpha channel and then returns the resulting color.
func (c RGBA) AlphaPremultiply() RGBA {
	return RGBA{c.R * c.A, c.G * c.A, c.B * c.A, c.A}
}

func (c RGBA) String() string {
	return fmt.Sprintf("RGBA(%.3f, %.3f, %.3f, %.3f)", c.R, c.G, c.B, c.A)
}

func (c RGBA) GoString() string {
	return fmt.Sprintf("color.RGBA{%#v, %#v, %#v, %#v}", c.R, c.G, c.B, c.A)
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

// Model is the color model for the renderer.
var Model image.ColorModel = image.ColorModelFunc(toGorayColor)

func toGorayColor(col image.Color) image.Color {
	switch col.(type) {
	case RGB, RGBA, Gray:
		return col
	}
	r, g, b, a := col.RGBA() // pre-multiplied
	return RGBA{
		float64(r*0xffff) / (float64(a) * math.MaxUint16),
		float64(g*0xffff) / (float64(a) * math.MaxUint16),
		float64(b*0xffff) / (float64(a) * math.MaxUint16),
		float64(a) / math.MaxUint16,
	}
}
