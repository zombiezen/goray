//
//	goray/core/color/ops.go
//	goray
//
//	Created by Ross Light on 2010-06-23.
//

package color

import (
	"math"
)

// IsBlack determines whether a color is absolute black.
func IsBlack(c Color) bool {
	return c.GetR() == 0 && c.GetG() == 0 && c.GetB() == 0
}

// GetEnergy calculates the overall brightness of a color.
func GetEnergy(c Color) float64 {
	return (c.GetR() + c.GetG() + c.GetB()) * 0.33333333333333
}

// Invert computes the inverse of the color.  However, black will always be black.
func Invert(c Color) Color {
	doInvert := func(comp float64) float64 {
		if comp == 0.0 {
			return 0.0
		}
		return 1.0 / comp
	}
	return NewRGB(doInvert(c.GetR()), doInvert(c.GetG()), doInvert(c.GetB()))
}

// Abs ensures that a color is positive.
func Abs(c Color) Color {
	return NewRGB(math.Fabs(c.GetR()), math.Fabs(c.GetG()), math.Fabs(c.GetB()))
}

// Add creates a new color that is equivalent to the sum of the colors given to it, disregarding alpha information.
func Add(c1, c2 Color) Color {
	return NewRGB(c1.GetR()+c2.GetR(), c1.GetG()+c2.GetG(), c1.GetB()+c2.GetB())
}

// AddAlpha creates a new color that is equivalent to the sum of the colors given to it.
func AddAlpha(c1, c2 AlphaColor) AlphaColor {
	return NewRGBA(c1.GetR()+c2.GetR(), c1.GetG()+c2.GetG(), c1.GetB()+c2.GetB(), c1.GetA()+c2.GetA())
}

// Sub creates a new color that is equivalent to the difference of the colors given to it, disregarding alpha information.
func Sub(c1, c2 Color) Color {
	return NewRGB(c1.GetR()-c2.GetR(), c1.GetG()-c2.GetG(), c1.GetB()-c2.GetB())
}

// SubAlpha creates a new color that is equivalent to the difference of the colors given to it.
func SubAlpha(c1, c2 AlphaColor) AlphaColor {
	return NewRGBA(c1.GetR()-c2.GetR(), c1.GetG()-c2.GetG(), c1.GetB()-c2.GetB(), c1.GetA()-c2.GetA())
}

// Mul creates a new color that is equivalent to the product of the colors given to it, disregarding alpha information.
func Mul(c1, c2 Color) Color {
	return NewRGB(c1.GetR()*c2.GetR(), c1.GetG()*c2.GetG(), c1.GetB()*c2.GetB())
}

// MulAlpha creates a new color that is equivalent to the product of the colors given to it.
func MulAlpha(c1, c2 AlphaColor) AlphaColor {
	return NewRGBA(c1.GetR()*c2.GetR(), c1.GetG()*c2.GetG(), c1.GetB()*c2.GetB(), c1.GetA()*c2.GetA())
}

// ScalarMul creates a new color that is equivalent to the color multiplied by a constant factor, disregarding alpha information.
func ScalarMul(c Color, f float64) Color {
	return NewRGB(c.GetR()*f, c.GetG()*f, c.GetB()*f)
}

// ScalarMulAlpha creates a new color that is equivalent to the color multiplied by a constant factor.
func ScalarMulAlpha(c AlphaColor, f float64) AlphaColor {
	return NewRGBA(c.GetR()*f, c.GetG()*f, c.GetB()*f, c.GetA()*f)
}

// ScalarDiv creates a new color that is equivalent to the color divided by a constant factor, disregarding alpha information.
func ScalarDiv(c Color, f float64) Color {
	return NewRGB(c.GetR()/f, c.GetG()/f, c.GetB()/f)
}

// ScalarDivAlpha creates a new color that is equivalent to the color divided by a constant factor.
func ScalarDivAlpha(c AlphaColor, f float64) AlphaColor {
	return NewRGBA(c.GetR()/f, c.GetG()/f, c.GetB()/f, c.GetA()/f)
}

// Mix creates a new color that is the additive mix of the two colors, disregarding alpha information.
func Mix(a, b Color, point float64) Color {
	if point <= 0 {
		return b
	} else if point >= 1 {
		return a
	}
	return Add(ScalarMul(a, point), ScalarMul(b, 1-point))
}

// MixAlpha creates a new color that is the additive mix of the two colors, using alpha to influence the mixing.
func MixAlpha(a, b AlphaColor, point float64) AlphaColor {
	if point <= 0 {
		return b
	} else if point >= 1 {
		return a
	}
	return AddAlpha(ScalarMulAlpha(a, point), ScalarMulAlpha(b, 1-point))
}
