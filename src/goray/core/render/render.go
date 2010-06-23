//
//  goray/core/render/render.go
//  goray
//
//  Created by Ross Light on 2010-05-27.
//

/* The render package provides the infrastructure for rendering an image.  It cannot render an image itself. */
package render

import (
	"image"
	"rand"
	"time"
	"goray/core/color"
	"goray/core/vector"
)

/* State stores information concerning the current rendering state */
type State struct {
	RayLevel       int
	Depth          float
	Contribution   float
	SkipElement    interface{}
	CurrentPass    int
	PixelSample    int
	RayDivision    int
	RayOffset      int
	Dc1, Dc2       float
	Traveled       float
	PixelNumber    int
	SamplingOffset uint
	ScreenPos      vector.Vector3D
	Chromatic      bool
	IncludeLights  bool
	WaveLength     bool
	Time           float
	Rand           *rand.Rand
}

/* Init initializes the state. */
func (st *State) Init(rgen *rand.Rand) {
	st.CurrentPass = 0
	st.PixelSample = 0
	st.RayOffset = 0
	st.IncludeLights = false
	st.SetDefaults()

	if rgen != nil {
		st.Rand = rgen
	} else {
		st.Rand = rand.New(rand.NewSource(time.Seconds()))
	}
}

/* SetDefaults changes some values to their initial values. */
func (st *State) SetDefaults() {
	st.RayLevel = 0
	st.Chromatic = true
	st.RayDivision = 1
	st.Dc1 = 0.0
	st.Dc2 = 0.0
	st.Traveled = 0.0
}

/* A Fragment stores a part of an image. */
type Fragment struct {
	Color color.AlphaColor
	X, Y  int
}

/*
   Image stores a two-dimensional array of colors.
   It implements the image.Image interface, so you can use it directly with the standard library.
*/
type Image struct {
	width, height int
	data          [][]color.RGBA
}

/* NewImage creates a new, black image with the given width and height. */
func NewImage(width, height int) (img *Image) {
	img = new(Image)
	img.width, img.height = width, height
	// Allocate image memory
	dataBlock := make([]color.RGBA, width*height)
	img.data = make([][]color.RGBA, height)
	for i := 0; i < height; i++ {
		img.data[i] = dataBlock[i*width : (i+1)*width]
	}
	return
}

func (i *Image) ColorModel() image.ColorModel { return color.Model }
func (i *Image) Width() int                   { return i.width }
func (i *Image) Height() int                  { return i.height }
func (i *Image) At(x, y int) image.Color {
	return i.data[y][x]
}

/* Clear sets all of the pixels in the image to a given color. */
func (i *Image) Clear(clearColor color.AlphaColor) {
	for y := 0; y < i.height; y++ {
		for x := 0; x < i.width; x++ {
			i.data[y][x].Copy(clearColor)
		}
	}
}

/* Acquire receives fragments from a channel until the channel is closed. */
func (i *Image) Acquire(ch <-chan Fragment) {
	for frag := range ch {
		i.data[frag.Y][frag.X].Copy(frag.Color)
	}
}

type Renderer interface {
    Render() <-chan Fragment
}
