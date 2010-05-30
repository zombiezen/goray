//
//  goray/color.go
//  goray
//
//  Created by Ross Light on 2010-05-22.
//

package color

import (
	"fmt"
	"image"
	"math"
	"./fmath"
)

type Alpha interface {
	GetA() float
}

type Color interface {
	GetR() float
	GetG() float
	GetB() float
}

type AlphaColor interface {
	Color
	Alpha
}

// RGB Color

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

var Model image.ColorModel = image.ColorModelFunc(toGorayColor)

func IsBlack(c Color) bool {
	return c.GetR() == 0 && c.GetG() == 0 && c.GetB() == 0
}

func GetEnergy(c Color) float {
	return (c.GetR() + c.GetG() + c.GetB()) * 0.33333333333333
}

func Invert(c Color) Color {
	doInvert := func(comp float) float {
		if comp == 0.0 {
			return 0.0
		}
		return 1.0 / comp
	}
	return NewRGB(doInvert(c.GetR()), doInvert(c.GetG()), doInvert(c.GetB()))
}

func Abs(c Color) Color {
	return NewRGB(fmath.Abs(c.GetR()), fmath.Abs(c.GetG()), fmath.Abs(c.GetB()))
}

func Add(c1, c2 Color) Color {
	return NewRGB(c1.GetR()+c2.GetR(), c1.GetG()+c2.GetG(), c1.GetB()+c2.GetB())
}

func AddAlpha(c1, c2 AlphaColor) AlphaColor {
	return NewRGBA(c1.GetR()+c2.GetR(), c1.GetG()+c2.GetG(), c1.GetB()+c2.GetB(), c1.GetA()+c2.GetA())
}

func Sub(c1, c2 Color) Color {
	return NewRGB(c1.GetR()-c2.GetR(), c1.GetG()-c2.GetG(), c1.GetB()-c2.GetB())
}

func SubAlpha(c1, c2 AlphaColor) AlphaColor {
	return NewRGBA(c1.GetR()-c2.GetR(), c1.GetG()-c2.GetG(), c1.GetB()-c2.GetB(), c1.GetA()-c2.GetA())
}

func Mul(c1, c2 Color) Color {
	return NewRGB(c1.GetR()*c2.GetR(), c1.GetG()*c2.GetG(), c1.GetB()*c2.GetB())
}

func MulAlpha(c1, c2 AlphaColor) AlphaColor {
	return NewRGBA(c1.GetR()*c2.GetR(), c1.GetG()*c2.GetG(), c1.GetB()*c2.GetB(), c1.GetA()*c2.GetA())
}

func ScalarMul(c Color, f float) Color {
	return NewRGB(c.GetR()*f, c.GetG()*f, c.GetB()*f)
}

func ScalarMulAlpha(c AlphaColor, f float) AlphaColor {
	return NewRGBA(c.GetR()*f, c.GetG()*f, c.GetB()*f, c.GetA()*f)
}

func ScalarDiv(c Color, f float) Color {
	return NewRGB(c.GetR()/f, c.GetG()/f, c.GetB()/f)
}

func ScalarDivAlpha(c AlphaColor, f float) AlphaColor {
	return NewRGBA(c.GetR()/f, c.GetG()/f, c.GetB()/f, c.GetA()/f)
}

func Mix(a, b Color, point float) Color {
	if point < 0 {
		return b
	} else if point > 1 {
		return a
	}
	return Add(ScalarMul(a, point), ScalarMul(b, 1-point))
}

func MixAlpha(a, b AlphaColor, point float) AlphaColor {
	if point < 0 {
		return b
	} else if point > 1 {
		return a
	}
	return AddAlpha(ScalarMulAlpha(a, point), ScalarMulAlpha(b, 1-point))
}
