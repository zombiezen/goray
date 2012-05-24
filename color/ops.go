/*
	Copyright (c) 2011 Ross Light.
	Copyright (c) 2005 Mathias Wein, Alejandro Conty, and Alfredo de Greef.

	This file is part of goray.

	goray is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	goray is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with goray.  If not, see <http://www.gnu.org/licenses/>.
*/

package color

import (
	"math"
)

// IsBlack determines whether a color is absolute black.
func IsBlack(c Color) bool {
	return c.Red() == 0 && c.Green() == 0 && c.Blue() == 0
}

// Energy calculates the overall brightness of a color.
func Energy(c Color) float64 {
	return (c.Red() + c.Green() + c.Blue()) / 3.0
}

// Invert computes the inverse of the color.  However, black will always be black.
func Invert(c Color) Color {
	doInvert := func(comp float64) float64 {
		if comp == 0.0 {
			return 0.0
		}
		return 1.0 / comp
	}
	return RGB{doInvert(c.Red()), doInvert(c.Green()), doInvert(c.Blue())}
}

// Abs ensures that a color is positive.
func Abs(c Color) Color {
	return RGB{math.Abs(c.Red()), math.Abs(c.Green()), math.Abs(c.Blue())}
}

// Add creates a new color that is equivalent to the sum of the colors given to it, disregarding alpha information.
func Add(c1, c2 Color) Color {
	return RGB{c1.Red() + c2.Red(), c1.Green() + c2.Green(), c1.Blue() + c2.Blue()}
}

// AddAlpha creates a new color that is equivalent to the sum of the colors given to it.
func AddAlpha(c1, c2 AlphaColor) AlphaColor {
	return RGBA{c1.Red() + c2.Red(), c1.Green() + c2.Green(), c1.Blue() + c2.Blue(), c1.Alpha() + c2.Alpha()}
}

// Sub creates a new color that is equivalent to the difference of the colors given to it, disregarding alpha information.
func Sub(c1, c2 Color) Color {
	return RGB{c1.Red() - c2.Red(), c1.Green() - c2.Green(), c1.Blue() - c2.Blue()}
}

// SubAlpha creates a new color that is equivalent to the difference of the colors given to it.
func SubAlpha(c1, c2 AlphaColor) AlphaColor {
	return RGBA{c1.Red() - c2.Red(), c1.Green() - c2.Green(), c1.Blue() - c2.Blue(), c1.Alpha() - c2.Alpha()}
}

// Mul creates a new color that is equivalent to the product of the colors given to it, disregarding alpha information.
func Mul(c1, c2 Color) Color {
	return RGB{c1.Red() * c2.Red(), c1.Green() * c2.Green(), c1.Blue() * c2.Blue()}
}

// MulAlpha creates a new color that is equivalent to the product of the colors given to it.
func MulAlpha(c1, c2 AlphaColor) AlphaColor {
	return RGBA{c1.Red() * c2.Red(), c1.Green() * c2.Green(), c1.Blue() * c2.Blue(), c1.Alpha() * c2.Alpha()}
}

// ScalarMul creates a new color that is equivalent to the color multiplied by a constant factor, disregarding alpha information.
func ScalarMul(c Color, f float64) Color {
	return RGB{c.Red() * f, c.Green() * f, c.Blue() * f}
}

// ScalarMulAlpha creates a new color that is equivalent to the color multiplied by a constant factor.
func ScalarMulAlpha(c AlphaColor, f float64) AlphaColor {
	return RGBA{c.Red() * f, c.Green() * f, c.Blue() * f, c.Alpha() * f}
}

// ScalarDiv creates a new color that is equivalent to the color divided by a constant factor, disregarding alpha information.
func ScalarDiv(c Color, f float64) Color {
	return RGB{c.Red() / f, c.Green() / f, c.Blue() / f}
}

// ScalarDivAlpha creates a new color that is equivalent to the color divided by a constant factor.
func ScalarDivAlpha(c AlphaColor, f float64) AlphaColor {
	return RGBA{c.Red() / f, c.Green() / f, c.Blue() / f, c.Alpha() / f}
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
