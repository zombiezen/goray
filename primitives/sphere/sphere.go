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

// Package sphere provides a spherical primitive.
package sphere

import (
	"math"

	"bitbucket.org/zombiezen/goray"
	"bitbucket.org/zombiezen/goray/bound"
	"bitbucket.org/zombiezen/math3/vec64"
)

type sphere struct {
	center   vec64.Vector
	radius   float64
	material goray.Material
}

var _ goray.Primitive = &sphere{}

// New creates a spherical primitive.
func New(center vec64.Vector, radius float64, material goray.Material) goray.Primitive {
	return &sphere{center, radius, material}
}

func (s *sphere) Bound() bound.Bound {
	r := vec64.Vector{s.radius * 1.0001, s.radius * 1.0001, s.radius * 1.0001}
	return bound.Bound{vec64.Sub(s.center, r), vec64.Add(s.center, r)}
}

func (s *sphere) IntersectsBound(b bound.Bound) bool { return true }
func (s *sphere) Material() goray.Material           { return s.material }

func (s *sphere) Intersect(r goray.Ray) (coll goray.Collision) {
	coll.Ray = r

	vf := vec64.Sub(r.From, s.center)
	ea := r.Dir.LengthSqr()
	eb := vec64.Dot(vf, r.Dir) * 2.0
	ec := vf.LengthSqr() - s.radius*s.radius
	osc := eb*eb - 4.0*ea*ec
	if osc < 0 {
		return
	}

	osc = math.Sqrt(osc)
	sol1 := (-eb - osc) / (ea * 2.0)
	sol2 := (-eb + osc) / (ea * 2.0)
	coll.RayDepth = sol1
	if coll.RayDepth < r.TMin {
		coll.RayDepth = sol2
		if coll.RayDepth < r.TMin {
			return
		}
	}
	coll.Primitive = s
	return
}

func (s *sphere) Surface(coll goray.Collision) (sp goray.SurfacePoint) {
	normal := vec64.Sub(coll.Point(), s.center)
	sp.HasOrco = true
	sp.OrcoPosition = normal
	normal = normal.Normalize()

	sp.Material = s.material
	sp.Primitive = s

	sp.Position = coll.Point()
	sp.Normal = normal
	sp.GeometricNormal = normal
	sp.U = math.Atan2(normal[1], normal[0])*(1.0/math.Pi) + 1
	sp.V = 1.0 - math.Acos(normal[2])*(1.0/math.Pi)
	return
}
