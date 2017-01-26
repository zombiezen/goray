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
	"zombiezen.com/go/goray/internal/color"
)

// Background is an interface for a rendering background.
type Background interface {
	// Color returns the background color for a given ray.
	Color(r Ray, state *RenderState, filtered bool) color.Color

	// Light returns the light source representing background lighting.
	// This may be nil if the background should only be sampled from BSDFs.
	Light() Light
}
