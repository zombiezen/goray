//
//  goray/color.go
//  goray
//
//  Created by Ross Light on 2010-05-22.
//

package goray

import "fmath"

type Color struct {
    R, G, B, A float
}

func NewColorRGB(r, g, b float) Color { return Color{r, g, b, 1.0} }

func (c Color) IsBlack() bool {
    return c.R == 0 && c.G == 0 && c.B == 0
}

func (c Color) GetEnergy() float {
    return (c.R + c.G + c.B) * 0.33333333333333
}

func (c Color) InvertRGB() Color {
    newColor := Color{A:c.A}
    if c.R != 0 { newColor.R = 1.0 / c.R }
    if c.G != 0 { newColor.G = 1.0 / c.G }
    if c.B != 0 { newColor.B = 1.0 / c.B }
    return newColor
}

func (c Color) AbsRGB() Color {
    return Color{fmath.Abs(c.R), fmath.Abs(c.G), fmath.Abs(c.B), c.A}
}

func ColorAdd(c1, c2 Color) Color {
    return Color{c1.R + c2.R, c1.G + c2.G, c1.B + c2.B, c1.A + c2.A}
}

func ColorSub(c1, c2 Color) Color {
    return Color{c1.R - c2.R, c1.G - c2.G, c1.B - c2.B, c1.A - c2.A}
}

func ColorMul(c1, c2 Color) Color {
    return Color{c1.R * c2.R, c1.G * c2.G, c1.B * c2.B, c1.A * c2.A}
}

func ColorScalarMul(c Color, f float) Color {
    return Color{c.R * f, c.G * f, c.B * f, c.A * f}
}

func ColorScalarDiv(c Color, f float) Color {
    return Color{c.R / f, c.G / f, c.B / f, c.A / f}
}

func (c Color) AlphaPremultiply() Color {
    return Color{c.R * c.A, c.G * c.A, c.B * c.A, c.A}
}

func ColorMix(a, b Color, point float) Color {
    if point < 0 {
        return b
    } else if point > 1 {
        return a
    }
    return ColorAdd(ColorScalarMul(a, point), ColorScalarMul(b, 1 - point))
}
