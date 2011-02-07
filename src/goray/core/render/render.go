//
//	goray/core/render/render.go
//	goray
//
//	Created by Ross Light on 2010-05-27.
//

// The render package provides the infrastructure for rendering an image.  It cannot render an image itself.
package render

import (
	"image"
	"goray/core/color"
	"goray/core/vector"
)

// State stores information concerning the current rendering state.
type State struct {
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
}

// Init initializes the state.
func (st *State) Init() {
	st.CurrentPass = 0
	st.PixelSample = 0
	st.RayOffset = 0
	st.IncludeLights = false
	st.SetDefaults()
}

// SetDefaults changes some values to their initial values.
func (st *State) SetDefaults() {
	st.RayLevel = 0
	st.Chromatic = true
	st.RayDivision = 1
	st.Dc1 = 0.0
	st.Dc2 = 0.0
	st.Traveled = 0.0
}

// A Fragment stores a part of an image.
type Fragment struct {
	Color color.AlphaColor
	X, Y  int
}

// Image stores a two-dimensional array of colors.
// It implements the image.Image interface, so you can use it directly with the standard library.
type Image struct {
	Width, Height int
	Pix           []color.RGBA
}

// NewImage creates a new, black image with the given width and height.
func NewImage(w, h int) (img *Image) {
	return &Image{
		Width:  w,
		Height: h,
		Pix:    make([]color.RGBA, w*h),
	}
}

func (i *Image) ColorModel() image.ColorModel { return color.Model }
func (i *Image) At(x, y int) image.Color      { return i.Pix[y*i.Width+x] }
func (i *Image) Bounds() image.Rectangle      { return image.Rect(0, 0, i.Width, i.Height) }

// Clear sets all of the pixels in the image to a given color.
func (i *Image) Clear(clearColor color.AlphaColor) {
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
