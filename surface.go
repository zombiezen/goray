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
	"bitbucket.org/zombiezen/math3/vec64"
)

// SurfacePoint represents a single point on an object's surface.
type SurfacePoint struct {
	Material  Material
	Light     Light
	Object    Object3D
	Primitive Primitive

	OriginX, OriginY int // Only used with "win" texture coordinate mode

	Normal, GeometricNormal, OrcoNormal vec64.Vector
	Position, OrcoPosition              vec64.Vector

	HasUV, HasOrco, Available bool
	PrimitiveNumber           int

	U, V               float64      // The texture coordinates
	NormalU, NormalV   vec64.Vector // Vectors building orthogonal shading space with normal
	WorldU, WorldV     vec64.Vector // U and V axes in world space
	ShadingU, ShadingV vec64.Vector // U and V axes in shading space
	SurfaceU, SurfaceV float64      // Raw surface parametric coordinates; required to evaluate Vmaps
}

// Differentials computes and stores data for surface intersections for differential rays.
// For more information, see http://www.opticalres.com/white%20papers/DifferentialRayTracing.pdf
type Differentials struct {
	X, Y  vec64.Vector
	Point SurfacePoint
}

// NewDifferentials creates a new Differentials struct.
func NewDifferentials(p SurfacePoint, r *DifferentialRay) Differentials {
	d := -vec64.Dot(p.Normal, p.Position)
	tx := -(vec64.Dot(p.Normal, r.FromX) + d) / vec64.Dot(p.Normal, r.DirX)
	px := vec64.Mul(r.FromX.AddScalar(tx), r.DirX)
	ty := -(vec64.Dot(p.Normal, r.FromY) + d) / vec64.Dot(p.Normal, r.DirY)
	py := vec64.Mul(r.FromY.AddScalar(ty), r.DirY)
	return Differentials{
		X:     vec64.Sub(px, p.Position),
		Y:     vec64.Sub(py, p.Position),
		Point: p,
	}
}

// ReflectRay computes differentials for a scattered ray.
// For an explanation, see: http://en.wikipedia.org/wiki/Specular_reflection
func (d Differentials) ReflectRay(in, out *DifferentialRay) {
	// Compute ray differential for specular reflection
	out.FromX = vec64.Add(d.Point.Position, d.X)
	out.FromY = vec64.Add(d.Point.Position, d.Y)

	// Compute differential reflected directions
	incidenceX, incidenceY := vec64.Sub(in.Dir, in.DirX), vec64.Sub(in.Dir, in.DirY)
	normDx, normDy := vec64.Dot(incidenceX, d.Point.Normal), vec64.Dot(incidenceY, d.Point.Normal)
	out.DirX = vec64.Sum(out.Dir, incidenceX.Negate(), d.Point.Normal.Scale(2*normDx))
	out.DirY = vec64.Sum(out.Dir, incidenceY.Negate(), d.Point.Normal.Scale(2*normDy))
}

// RefractRay computes differentials for a scattered ray.
// For an explanation, see: http://en.wikipedia.org/wiki/Snell's_law#Vector_form
func (d Differentials) RefractRay(in, out *DifferentialRay, ior float64) {
	out.FromX = vec64.Add(d.Point.Position, d.X)
	out.FromY = vec64.Add(d.Point.Position, d.Y)

	incidenceX, incidenceY := vec64.Sub(in.Dir, in.DirX), vec64.Sub(in.Dir, in.DirY)
	normDx, normDy := vec64.Dot(incidenceX, d.Point.Normal), vec64.Dot(incidenceY, d.Point.Normal)

	muDeriv := ior - ior*ior*vec64.Dot(in.Dir, d.Point.Normal)/vec64.Dot(out.Dir, d.Point.Normal)
	muDx := muDeriv * normDx
	muDy := muDeriv * normDy

	out.DirX = vec64.Sum(out.Dir, incidenceX.Scale(ior), d.Point.Normal.Scale(-muDx))
	out.DirY = vec64.Sum(out.Dir, incidenceY.Scale(ior), d.Point.Normal.Scale(-muDy))
}

func (d Differentials) ProjectedPixelArea() float64 {
	return vec64.Cross(d.X, d.Y).Length()
}
