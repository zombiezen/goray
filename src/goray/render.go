//
//  goray/render.go
//  goray
//
//  Created by Ross Light on 2010-05-27.
//

package render

import (
	"image"
	"rand"
	"time"
	"./goray/color"
	"./goray/vector"
)

type RenderState struct {
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
	Rand           *rand.Rand
}

func (st *RenderState) Init(rgen *rand.Rand) {
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

func (st *RenderState) SetDefaults() {
	st.RayLevel = 0
	st.Chromatic = true
	st.RayDivision = 1
	st.Dc1 = 0.0
	st.Dc2 = 0.0
	st.Traveled = 0.0
}

type Fragment struct {
    Color color.AlphaColor
    X, Y int
}

type Image struct {
	width, height int
	data          [][]color.RGBA
}

func NewImage(width, height int) (img *Image) {
	img = new(Image)
    img.width, img.height = width, height
    // Allocate image memory
    dataBlock := make([]color.RGBA, width * height)
    img.data = make([][]color.RGBA, height)
    for i := 0; i < height; i++ {
        img.data[i] = dataBlock[i * width:(i + 1) * width]
    }
    return
}

func (i *Image) ColorModel() image.ColorModel { return color.Model }
func (i *Image) Width() int                   { return i.width }
func (i *Image) Height() int                  { return i.height }
func (i *Image) At(x, y int) image.Color {
	return i.data[y][x]
}

func (i *Image) Clear(clearColor color.AlphaColor) {
    for y := 0; y < i.height; y++ {
        for x := 0; x < i.width; x++ {
            i.data[y][x].Copy(clearColor)
        }
    }
}

func (i *Image) Acquire(ch <-chan Fragment) {
    for !closed(ch) {
        frag := <-ch
        i.data[frag.Y][frag.X].Copy(frag.Color)
    }
}
