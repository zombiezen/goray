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

package goray

import (
	color_ "bitbucket.org/zombiezen/goray/color"
	"bitbucket.org/zombiezen/goray/vector"
	"image"
	"image/color"
)

// RenderState stores information concerning the current rendering state.
type RenderState struct {
	RayLevel       int
	Depth          float64
	Contribution   float64
	SkipElement    interface{}
	CurrentPass    int
	PixelSample    int
	RayDivision    int
	RayOffset      int
	Dc1, Dc2       float64
	Traveled       float64
	PixelNumber    int
	SamplingOffset uint
	ScreenPos      vector.Vector3D
	Chromatic      bool
	IncludeLights  bool
	WaveLength     float64
	Time           float64
	MaterialData   interface{}
}

// Init initializes the state.
func (st *RenderState) Init() {
	st.CurrentPass = 0
	st.PixelSample = 0
	st.RayOffset = 0
	st.IncludeLights = false
	st.SetDefaults()
}

// SetDefaults changes some values to their initial values.
func (st *RenderState) SetDefaults() {
	st.RayLevel = 0
	st.Chromatic = true
	st.RayDivision = 1
	st.Dc1 = 0.0
	st.Dc2 = 0.0
	st.Traveled = 0.0
}

// Fragment stores a single element of an image.
type Fragment struct {
	Color color_.AlphaColor
	X, Y  int
}

// Image stores a two-dimensional array of colors.
// It implements the image.Image interface, so you can use it directly with the standard library.
type Image struct {
	Width, Height int
	Pix           []color_.RGBA
}

// NewImage creates a new, blank image with the given width and height.
func NewImage(w, h int) (img *Image) {
	return &Image{
		Width:  w,
		Height: h,
		Pix:    make([]color_.RGBA, w*h),
	}
}

// NewGoImage creates a new image based on an image from the standard Go library.
func NewGoImage(oldImage image.Image) (newImage *Image) {
	bd := oldImage.Bounds()
	newImage = &Image{Width: bd.Dx(), Height: bd.Dy(), Pix: make([]color_.RGBA, bd.Dx()*bd.Dy())}
	model := newImage.ColorModel()
	for y := 0; y < newImage.Height; y++ {
		for x := 0; x < newImage.Width; x++ {
			col := model.Convert(oldImage.At(bd.Min.X+x, bd.Min.Y+y)).(color_.RGBA)
			newImage.Pix[y*newImage.Width+x] = col
		}
	}
	return
}

func (i *Image) ColorModel() color.Model {
	return color_.Model
}

func (i *Image) At(x, y int) color.Color {
	return i.Pix[y*i.Width+x]
}

func (i *Image) Bounds() image.Rectangle {
	return image.Rect(0, 0, i.Width, i.Height)
}

// Pixel returns the color.RGBA at a position. If you are iterating over the pixels, use the Pix slice directly.
func (i *Image) Pixel(x, y int) color_.RGBA {
	return i.Pix[y*i.Width+x]
}

// Clear sets all of the pixels in the image to a given color.
func (i *Image) Clear(clearColor color_.AlphaColor) {
	for j, _ := range i.Pix {
		i.Pix[j].Copy(clearColor)
	}
}

// Acquire receives fragments from a channel until the channel is closed.
func (i *Image) Acquire(ch <-chan Fragment) {
	for frag := range ch {
		i.Pix[frag.Y*i.Width+frag.X].Copy(frag.Color)
	}
}
