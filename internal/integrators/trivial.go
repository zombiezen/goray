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

package integrators

import (
	"zombiezen.com/go/goray/internal/color"
	"zombiezen.com/go/goray/internal/goray"
)

type trivial struct{}

var _ goray.SurfaceIntegrator = trivial{}

// NewTrivial creates a surface integrator that outputs white for collisions and
// gray for no collisions.
func NewTrivial() goray.SurfaceIntegrator {
	return trivial{}
}

func (ti trivial) SurfaceIntegrator()         {}
func (ti trivial) Preprocess(sc *goray.Scene) {}

func (ti trivial) Integrate(sc *goray.Scene, s *goray.RenderState, r goray.DifferentialRay) color.AlphaColor {
	if coll := sc.Intersect(r.Ray, -1); coll.Hit() {
		return color.NewRGBAFromColor(color.White, 1.0)
	}
	return color.NewRGBAFromColor(color.Gray(0.1), 0.0)
}
