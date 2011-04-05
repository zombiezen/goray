//
//	goray/std/textures/image.go
//	goray
//
//	Created by Ross Light on 2011-04-02.
//

package image

import (
	"math"
	"os"

	"image"
	"image/jpeg"
	"image/png"

	"goray/core/color"
	"goray/core/render"
	"goray/core/vector"

	"goray/std/shaders/texmap"

	yamldata "goyaml.googlecode.com/hg/data"
)

// Ensure that these packages get imported
var _ = jpeg.Decode
var _ = png.Decode

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
	Image         *render.Image
	Interpolation Interpolation
	UseAlpha      bool

	ClipMode         ClipMode
	RepeatX, RepeatY int
}

var _ texmap.DiscreteTexture = &Texture{}

func (t *Texture) ColorAt(pt vector.Vector3D) (col color.AlphaColor) {
	pt = vector.Vector3D{pt[vector.X], -pt[vector.Y], pt[vector.Z]}
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

func (t *Texture) ScalarAt(pt vector.Vector3D) float64 {
	return color.Energy(t.ColorAt(pt))
}

func (t *Texture) Is3D() bool                { return false }
func (t *Texture) IsNormalMap() bool         { return false }
func (t *Texture) Resolution() (x, y, z int) { x, y = t.Image.Width, t.Image.Height; return }

func (t *Texture) mapping(texPt vector.Vector3D) (p vector.Vector3D, outside bool) {
	texPt = vector.ScalarAdd(vector.ScalarMul(texPt, 0.5), 0.5)
	if t.ClipMode == ClipRepeat {
		if t.RepeatX > 1 {
			texPt[vector.X] = mapRepeat(texPt[vector.X], t.RepeatX)
		}
		if t.RepeatY > 1 {
			texPt[vector.Y] = mapRepeat(texPt[vector.Y], t.RepeatY)
		}
	}
	switch t.ClipMode {
	case ClipCube:
		if texPt[vector.X] < 0 || texPt[vector.X] > 1 || texPt[vector.Y] < 0 || texPt[vector.Y] > 1 || texPt[vector.Z] < -1 || texPt[vector.Z] > 1 {
			outside = true
		}
	case Clip:
		if texPt[vector.X] < 0 || texPt[vector.X] > 1 || texPt[vector.Y] < 0 || texPt[vector.Y] > 1 {
			outside = true
		}
	case ClipExtend:
		texPt[vector.X] = mapExtend(texPt[vector.X])
		texPt[vector.Y] = mapExtend(texPt[vector.Y])
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

func interpolateImage(img *render.Image, intp Interpolation, p vector.Vector3D) color.AlphaColor {
	xf := float64(img.Width) * (p[vector.X] - math.Floor(p[vector.X]))
	yf := float64(img.Width) * (p[vector.Y] - math.Floor(p[vector.Y]))
	if intp != NoInterpolation {
		xf -= 0.5
		yf -= 0.5
	}
	x, y := int(xf), int(yf)
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= img.Width {
		x = img.Width - 1
	}
	if y >= img.Height {
		y = img.Height - 1
	}
	c1 := img.Pix[y*img.Width+x]
	// TODO: Add interpolation
	return c1
}

func Construct(m yamldata.Map) (data interface{}, err os.Error) {
	path, ok := m["path"].(string)
	if !ok {
		err = os.NewError("Image must contain path")
		return
	}
	// TODO: Read more arguments

	// Open image file
	// XXX: Possible security issue
	imageFile, err := os.Open(path, os.O_RDONLY, 0)
	if err != nil {
		return
	}
	defer imageFile.Close()
	img, _, err := image.Decode(imageFile)
	if err != nil {
		return
	}
	// Construct texture
	tex := &Texture{
		Image: render.NewGoImage(img),
	}
	return tex, nil
}
