//
//  goray/color.go
//  goray
//
//  Created by Ross Light on 2010-05-22.
//

/*
   The goray/color package provides abstracted color.
   This interface specifically differes from the image.Color interfaces because the render
   uses floating point math.  Also, colors in this package are not clamped to [0, 1], they
   are clamped to [0, Inf).
*/
package color

import (
	"fmt"
	"image"
	"math"
	"goray/fmath"
)

/* Alpha defines anything that has an alpha channel. */
type Alpha interface {
	GetA() float
}

/* Color defines anything that has red, green, and blue channels. */
type Color interface {
	GetR() float
	GetG() float
	GetB() float
}

/* AlphaColor defines anything that has red, green, blue, and alpha channels. */
type AlphaColor interface {
	Color
	Alpha
}

/* RGB defines a color that has red, green, and blue channels.  It fulfills the Color interface. */
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

func quantizeComponent(f float) uint32 {
	temp := uint64(f * math.MaxUint32)
	if temp > math.MaxUint32 {
		return math.MaxUint32
	} else if temp < 0 {
		return 0
	}
	return uint32(temp)
}

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

// Operations

func toGorayColor(col image.Color) image.Color {
	if _, ok := col.(RGB); ok {
		return col
	}
	if _, ok := col.(RGBA); ok {
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

/* The color model for the renderer. */
var Model image.ColorModel = image.ColorModelFunc(toGorayColor)

/* IsBlack determines whether a color is absolute black. */
func IsBlack(c Color) bool {
	return c.GetR() == 0 && c.GetG() == 0 && c.GetB() == 0
}

/* GetEnergy calculates the overall brightness of a color. */
func GetEnergy(c Color) float {
	return (c.GetR() + c.GetG() + c.GetB()) * 0.33333333333333
}

/* Invert computes the inverse of the color.  However, black will always be black. */
func Invert(c Color) Color {
	doInvert := func(comp float) float {
		if comp == 0.0 {
			return 0.0
		}
		return 1.0 / comp
	}
	return NewRGB(doInvert(c.GetR()), doInvert(c.GetG()), doInvert(c.GetB()))
}

/* Abs ensures that a color is positive. */
func Abs(c Color) Color {
	return NewRGB(fmath.Abs(c.GetR()), fmath.Abs(c.GetG()), fmath.Abs(c.GetB()))
}

/* Add creates a new color that is equivalent to the sum of the colors given to it, disregarding alpha information. */
func Add(c1, c2 Color) Color {
	return NewRGB(c1.GetR()+c2.GetR(), c1.GetG()+c2.GetG(), c1.GetB()+c2.GetB())
}

/* AddAlpha creates a new color that is equivalent to the sum of the colors given to it. */
func AddAlpha(c1, c2 AlphaColor) AlphaColor {
	return NewRGBA(c1.GetR()+c2.GetR(), c1.GetG()+c2.GetG(), c1.GetB()+c2.GetB(), c1.GetA()+c2.GetA())
}

/* Sub creates a new color that is equivalent to the difference of the colors given to it, disregarding alpha information. */
func Sub(c1, c2 Color) Color {
	return NewRGB(c1.GetR()-c2.GetR(), c1.GetG()-c2.GetG(), c1.GetB()-c2.GetB())
}

/* SubAlpha creates a new color that is equivalent to the difference of the colors given to it. */
func SubAlpha(c1, c2 AlphaColor) AlphaColor {
	return NewRGBA(c1.GetR()-c2.GetR(), c1.GetG()-c2.GetG(), c1.GetB()-c2.GetB(), c1.GetA()-c2.GetA())
}

/* Mul creates a new color that is equivalent to the product of the colors given to it, disregarding alpha information. */
func Mul(c1, c2 Color) Color {
	return NewRGB(c1.GetR()*c2.GetR(), c1.GetG()*c2.GetG(), c1.GetB()*c2.GetB())
}

/* MulAlpha creates a new color that is equivalent to the product of the colors given to it. */
func MulAlpha(c1, c2 AlphaColor) AlphaColor {
	return NewRGBA(c1.GetR()*c2.GetR(), c1.GetG()*c2.GetG(), c1.GetB()*c2.GetB(), c1.GetA()*c2.GetA())
}

/* ScalarMul creates a new color that is equivalent to the color multiplied by a constant factor, disregarding alpha information. */
func ScalarMul(c Color, f float) Color {
	return NewRGB(c.GetR()*f, c.GetG()*f, c.GetB()*f)
}

/* ScalarMulAlpha creates a new color that is equivalent to the color multiplied by a constant factor. */
func ScalarMulAlpha(c AlphaColor, f float) AlphaColor {
	return NewRGBA(c.GetR()*f, c.GetG()*f, c.GetB()*f, c.GetA()*f)
}

/* ScalarDiv creates a new color that is equivalent to the color divided by a constant factor, disregarding alpha information. */
func ScalarDiv(c Color, f float) Color {
	return NewRGB(c.GetR()/f, c.GetG()/f, c.GetB()/f)
}

/* ScalarDivAlpha creates a new color that is equivalent to the color divided by a constant factor. */
func ScalarDivAlpha(c AlphaColor, f float) AlphaColor {
	return NewRGBA(c.GetR()/f, c.GetG()/f, c.GetB()/f, c.GetA()/f)
}

/* Mix creates a new color that is the additive mix of the two colors, disregarding alpha information. */
func Mix(a, b Color, point float) Color {
	if point < 0 {
		return b
	} else if point > 1 {
		return a
	}
	return Add(ScalarMul(a, point), ScalarMul(b, 1-point))
}

/* MixAlpha creates a new color that is the additive mix of the two colors, using alpha to influence the mixing. */
func MixAlpha(a, b AlphaColor, point float) AlphaColor {
	if point < 0 {
		return b
	} else if point > 1 {
		return a
	}
	return AddAlpha(ScalarMulAlpha(a, point), ScalarMulAlpha(b, 1-point))
}
