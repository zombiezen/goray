//
//	goray/std/shaders/texmap/texture.go
//	goray
//
//	Created by Ross Light on 2011-04-02.
//

package texmap

import (
	"goray/core/color"
	"goray/core/vector"
)

// A Texture is a 2D/3D function for surface values.  This is usually based on a raster image, but could be procedurally generated.
type Texture interface {
	ColorAt(pt vector.Vector3D) color.AlphaColor
	ScalarAt(pt vector.Vector3D) float64
	Is3D() bool
	IsNormalMap() bool
}

// A DiscreteTexture is a texture which has quantized values (a raster image).
type DiscreteTexture interface {
	Texture
	Resolution() (x, y, z int)
}
