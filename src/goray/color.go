//
//  goray/color.go
//  goray
//
//  Created by Ross Light on 2010-05-22.
//

package color

import "./fmath"

type Color interface {
	GetR() float
	GetG() float
	GetB() float
	GetA() float
}

// RGB Color

type RGB struct {
	r, g, b float
}

func NewRGB(r, g, b float) RGB { return RGB{r, g, b} }
func DiscardAlpha(c Color) RGB { return NewRGB(c.GetR(), c.GetG(), c.GetB()) }
func (c RGB) GetR() float      { return c.r }
func (c RGB) GetG() float      { return c.g }
func (c RGB) GetB() float      { return c.b }
func (c RGB) GetA() float      { return 1.0 }

func IsBlack(c Color) bool {
	return c.GetR() == 0 && c.GetG() == 0 && c.GetB() == 0
}

func GetEnergy(c Color) float {
	return (c.GetR() + c.GetG() + c.GetB()) * 0.33333333333333
}

func InvertRGB(c Color) Color {
	newColor := RGBA{a: c.GetA()}
	if c.GetR() != 0 {
		newColor.RGB.r = 1.0 / c.GetR()
	}
	if c.GetG() != 0 {
		newColor.RGB.g = 1.0 / c.GetG()
	}
	if c.GetB() != 0 {
		newColor.RGB.b = 1.0 / c.GetB()
	}
	return newColor
}

func AbsRGB(c Color) Color {
	return NewRGBA(fmath.Abs(c.GetR()), fmath.Abs(c.GetG()), fmath.Abs(c.GetB()), c.GetA())
}

func Add(c1, c2 Color) Color {
	return NewRGBA(c1.GetR()+c2.GetR(), c1.GetG()+c2.GetG(), c1.GetB()+c2.GetB(), c1.GetA()+c2.GetA())
}

func Sub(c1, c2 Color) Color {
	return NewRGBA(c1.GetR()-c2.GetR(), c1.GetG()-c2.GetG(), c1.GetB()-c2.GetB(), c1.GetA()-c2.GetA())
}

func Mul(c1, c2 Color) Color {
	return NewRGBA(c1.GetR()*c2.GetR(), c1.GetG()*c2.GetG(), c1.GetB()*c2.GetB(), c1.GetA()*c2.GetA())
}

func ScalarMul(c Color, f float) Color {
	return NewRGBA(c.GetR()*f, c.GetG()*f, c.GetB()*f, c.GetA()*f)
}

func ScalarDiv(c Color, f float) Color {
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

// RGBA colors

type RGBA struct {
	RGB
	a float
}

func NewRGBA(r, g, b, a float) RGBA { return RGBA{NewRGB(r, g, b), a} }
func (c RGBA) GetA() float          { return c.a }

func (c RGBA) AlphaPremultiply() RGBA {
	return NewRGBA(c.GetR()*c.a, c.GetG()*c.a, c.GetB()*c.a, c.a)
}
