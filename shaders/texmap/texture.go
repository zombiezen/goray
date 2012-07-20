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

package texmap

import (
	"bitbucket.org/zombiezen/goray/color"
	"bitbucket.org/zombiezen/math3/vec64"
)

// A Texture is a 2D/3D function for surface values.  This is usually based on a raster image, but could be procedurally generated.
type Texture interface {
	ColorAt(pt vec64.Vector) color.AlphaColor
	ScalarAt(pt vec64.Vector) float64
	Is3D() bool
	IsNormalMap() bool
}

// A DiscreteTexture is a texture which has quantized values (a raster image).
type DiscreteTexture interface {
	Texture
	Resolution() (x, y, z int)
}
