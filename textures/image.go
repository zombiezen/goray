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

package textures

import (
	"errors"
	"math"

	"bitbucket.org/zombiezen/goray"
	"bitbucket.org/zombiezen/goray/color"
	"bitbucket.org/zombiezen/goray/shaders/texmap"
	"bitbucket.org/zombiezen/goray/vecutil"
	yamldata "bitbucket.org/zombiezen/goray/yaml/data"
	"bitbucket.org/zombiezen/goray/yaml/parser"
	"bitbucket.org/zombiezen/goray/yamlscene"
	"bitbucket.org/zombiezen/math3/vec64"
)

type Interpolation int

const (
	NoInterpolation Interpolation = 0
	Bilinear
	Bicubic
)

type ClipMode int

const (
	ClipExtend ClipMode = iota
	Clip
	ClipCube
	ClipRepeat
)

type Texture struct {
	Image         *goray.Image
	Interpolation Interpolation
	UseAlpha      bool

	ClipMode         ClipMode
	RepeatX, RepeatY int
}

var _ texmap.DiscreteTexture = &Texture{}

func (t *Texture) ColorAt(pt vec64.Vector) (col color.AlphaColor) {
	pt = vec64.Vector{pt[0], -pt[1], pt[2]}
	pt, outside := t.mapping(pt)
	if outside {
		return color.RGBA{}
	}
	col = interpolateImage(t.Image, t.Interpolation, pt)
	if !t.UseAlpha {
		col = color.NewRGBAFromColor(col, 1.0)
	}
	return
}

func (t *Texture) ScalarAt(pt vec64.Vector) float64 {
	return color.Energy(t.ColorAt(pt))
}

func (t *Texture) Is3D() bool                { return false }
func (t *Texture) IsNormalMap() bool         { return false }
func (t *Texture) Resolution() (x, y, z int) { x, y = t.Image.Width, t.Image.Height; return }

func (t *Texture) mapping(texPt vec64.Vector) (p vec64.Vector, outside bool) {
	texPt = texPt.Scale(0.5).AddScalar(0.5)
	if t.ClipMode == ClipRepeat {
		if t.RepeatX > 1 {
			texPt[vecutil.X] = mapRepeat(texPt[vecutil.X], t.RepeatX)
		}
		if t.RepeatY > 1 {
			texPt[vecutil.Y] = mapRepeat(texPt[vecutil.Y], t.RepeatY)
		}
	}
	switch t.ClipMode {
	case ClipCube:
		if texPt[vecutil.X] < 0 || texPt[vecutil.X] > 1 || texPt[vecutil.Y] < 0 || texPt[vecutil.Y] > 1 || texPt[vecutil.Z] < -1 || texPt[vecutil.Z] > 1 {
			outside = true
		}
	case Clip:
		if texPt[vecutil.X] < 0 || texPt[vecutil.X] > 1 || texPt[vecutil.Y] < 0 || texPt[vecutil.Y] > 1 {
			outside = true
		}
	case ClipExtend:
		texPt[vecutil.X] = mapExtend(texPt[vecutil.X])
		texPt[vecutil.Y] = mapExtend(texPt[vecutil.Y])
	}
	p = texPt
	return
}

func mapRepeat(x float64, repeat int) float64 {
	x *= float64(repeat)
	if x > 1 {
		x -= math.Trunc(x)
	} else if x < 0 {
		x += 1 - math.Trunc(x)
	}
	return x
}

func mapExtend(x float64) float64 {
	if x >= 1 {
		return 1
	} else if x < 0 {
		return 0
	}
	return x
}

func cubicInterpolate(c1, c2, c3, c4 color.AlphaColor, x float64) (col color.AlphaColor) {
	x2 := x * x
	x3 := x2 * x
	col = color.ScalarMulAlpha(c1, (-1/3)*x3+(4/5)*x2-(7/15)*x)
	col = color.AddAlpha(col, color.ScalarMulAlpha(c2, x3-(9/5)*x2-(1/5)*x+(1/15)))
	col = color.AddAlpha(col, color.ScalarMulAlpha(c3, -x3+(6/5)*x2+(4/5)*x))
	col = color.AddAlpha(col, color.ScalarMulAlpha(c4, (1/3)*x3-(1/5)*x2-(2/15)*x))
	return
}

func interpolateImage(img *goray.Image, intp Interpolation, p vec64.Vector) color.AlphaColor {
	xf := float64(img.Width) * (p[vecutil.X] - math.Floor(p[vecutil.X]))
	yf := float64(img.Width) * (p[vecutil.Y] - math.Floor(p[vecutil.Y]))
	if intp != NoInterpolation {
		xf -= 0.5
		yf -= 0.5
	}
	x, y := clampToRes(int(xf), int(yf), img.Width, img.Height)
	c1 := img.Pix[y*img.Width+x]
	if intp == NoInterpolation {
		return c1
	}

	// Now for the fun stuff:
	x2, y2 := clampToRes(x+1, y+1, img.Width, img.Height)
	c2 := img.Pixel(x2, y)
	c3 := img.Pixel(x, y2)
	c4 := img.Pixel(x2, y2)
	dx, dy := xf-math.Floor(xf), yf-math.Floor(yf)
	if intp == Bilinear {
		w0, w1, w2, w3 := (1-dx)*(1-dy), (1-dx)*dy, dx*(1-dy), dx*dy
		return color.RGBA{
			w0*c1.Red() + w1*c3.Red() + w2*c2.Red() + w3*c4.Red(),
			w0*c1.Green() + w1*c3.Green() + w2*c2.Green() + w3*c4.Green(),
			w0*c1.Blue() + w1*c3.Blue() + w2*c2.Blue() + w3*c4.Blue(),
			w0*c1.Alpha() + w1*c3.Alpha() + w2*c2.Alpha() + w3*c4.Alpha(),
		}
	}
	x0, y0 := clampToRes(x-1, y-1, img.Width, img.Height)
	x3, y3 := clampToRes(x2+1, y2+1, img.Width, img.Height)
	c0 := color.AlphaColor(img.Pixel(x0, y0))
	c5 := color.AlphaColor(img.Pixel(x, y0))
	c6 := color.AlphaColor(img.Pixel(x2, y0))
	c7 := color.AlphaColor(img.Pixel(x3, y0))
	c8 := color.AlphaColor(img.Pixel(x0, y))
	c9 := color.AlphaColor(img.Pixel(x3, y))
	cA := color.AlphaColor(img.Pixel(x0, y2))
	cB := color.AlphaColor(img.Pixel(x3, y2))
	cC := color.AlphaColor(img.Pixel(x0, y3))
	cD := color.AlphaColor(img.Pixel(x, y3))
	cE := color.AlphaColor(img.Pixel(x2, y3))
	cF := color.AlphaColor(img.Pixel(x3, y3))
	c0 = cubicInterpolate(c0, c5, c6, c7, dx)
	c8 = cubicInterpolate(c8, c1, c2, c9, dx)
	cA = cubicInterpolate(cA, c3, c4, cB, dx)
	cC = cubicInterpolate(cC, cD, cE, cF, dx)
	return cubicInterpolate(c0, c8, cA, cC, dy)
}

func clampToRes(x0, y0, w, h int) (x, y int) {
	switch {
	case x0 < 0:
		x = 0
	case x0 >= w:
		x = w - 1
	default:
		x = x0
	}
	switch {
	case y0 < 0:
		y = 0
	case y0 >= h:
		y = h - 1
	default:
		y = y0
	}
	return
}

func init() {
	yamlscene.Constructor[yamlscene.StdPrefix+"textures/image"] = yamldata.ConstructorFunc(constructImage)
}

func constructImage(n parser.Node, ud interface{}) (interface{}, error) {
	mm, ok := n.(*parser.Mapping)
	if !ok {
		return nil, errors.New("Constructor requires a mapping")
	}

	var loader ImageLoader
	if ud != nil {
		userData, ok := ud.(yamlscene.Params)
		if ok && userData != nil {
			loader, ok = userData["ImageLoader"].(ImageLoader)
			if !ok && userData["ImageLoader"] != nil {
				return nil, errors.New("ImageLoader does not implement goray/std/texture/image.Loader interface")
			}
		}
	}
	if loader == nil {
		return nil, errors.New("No image loader provided")
	}

	m := yamldata.Map(mm.Map()).Copy()
	m.SetDefault("interpolation", "none")
	m.SetDefault("useAlpha", true)
	m.SetDefault("clip", "extend")
	m.SetDefault("repeatX", 1)
	m.SetDefault("repeatY", 1)

	// Image name
	name, ok := m["name"].(string)
	if !ok {
		return nil, errors.New("Image must contain name")
	}

	// Interpolation
	intpName, ok := m["interpolation"].(string)
	if !ok {
		return nil, errors.New("interpolation must be a string")
	}
	var intp Interpolation
	switch intpName {
	case "none":
		intp = NoInterpolation
	case "bilinear":
		intp = Bilinear
	case "bicubic":
		intp = Bicubic
	default:
		return nil, errors.New("Invalid interpolation method")
	}

	// Use Alpha
	useAlpha, ok := yamldata.AsBool(m["useAlpha"])
	if !ok {
		return nil, errors.New("useAlpha must be a boolean")
	}

	// Clipping Mode
	clipName, ok := m["clip"].(string)
	if !ok {
		return nil, errors.New("clip must be a string")
	}
	var clip ClipMode
	switch clipName {
	case "extend":
		clip = ClipExtend
	case "clip":
		clip = Clip
	case "cube":
		clip = ClipCube
	case "repeat":
		clip = ClipRepeat
	default:
		return nil, errors.New("Invalid clipping mode")
	}

	// Repeat X and Y
	repeatX, ok := yamldata.AsUint(m["repeatX"])
	if !ok {
		return nil, errors.New("repeatX must be an integer")
	}
	repeatY, ok := yamldata.AsUint(m["repeatY"])
	if !ok {
		return nil, errors.New("repeatY must be an integer")
	}

	// Open image file
	img, err := loader.LoadImage(name)
	if err != nil {
		return nil, err
	}

	// Construct texture
	return &Texture{
		Image:         img,
		Interpolation: intp,
		UseAlpha:      useAlpha,
		ClipMode:      clip,
		RepeatX:       int(repeatX),
		RepeatY:       int(repeatY),
	}, nil
}
