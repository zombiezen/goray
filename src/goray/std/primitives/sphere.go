//
//  goray/std/primitives/sphere.go
//  goray
//
//  Created by Ross Light on 2010-06-05.
//

/* The goray/std/primitives/sphere package provides a spherical primitive */
package sphere

import "math"
import "./fmath"

import (
	"./goray/bound"
	"./goray/material"
	"./goray/primitive"
	"./goray/ray"
	"./goray/surface"
	"./goray/vector"
)

type sphere struct {
	center   vector.Vector3D
	radius   float
	material material.Material
}

/* New creates a spherical primitive. */
func New(center vector.Vector3D, radius float, material material.Material) primitive.Primitive {
	return &sphere{center, radius, material}
}

func (s *sphere) GetBound() *bound.Bound {
	r := vector.Vector3D{s.radius * 1.0001, s.radius * 1.0001, s.radius * 1.0001}
	return bound.New(vector.Sub(s.center, r), vector.Add(s.center, r))
}

func (s *sphere) IntersectsBound(b *bound.Bound) bool { return true }
func (s *sphere) GetMaterial() material.Material      { return s.material }

func (s *sphere) Intersect(r ray.Ray) (coll primitive.Collision) {
	coll.Ray = r

	vf := vector.Sub(r.From(), s.center)
	ea := r.Dir().LengthSqr()
	eb := vector.Dot(vf, r.Dir()) * 2.0
	ec := vf.LengthSqr() - s.radius*s.radius
	osc := eb*eb - 4.0*ea*ec
	if osc < 0 {
		return
	}

	osc = fmath.Sqrt(osc)
	sol1 := (-eb - osc) / (ea * 2.0)
	sol2 := (-eb + osc) / (ea * 2.0)
	coll.RayDepth = sol1
	if coll.RayDepth < r.TMin() {
		coll.RayDepth = sol2
		if coll.RayDepth < r.TMin() {
			return
		}
	}
	coll.Primitive = s
	return
}

func (s *sphere) GetSurface(coll primitive.Collision) (sp surface.Point) {
	normal := vector.Sub(coll.GetPoint(), s.center)
	sp.HasOrco = true
	sp.OrcoPosition = normal
	normal = normal.Normalize()

	sp.Material = s.material
	sp.Primitive = s

	sp.Position = coll.GetPoint()
	sp.Normal = normal
	sp.GeometricNormal = normal
	sp.U = fmath.Atan2(normal.Y, normal.X)*(1.0/math.Pi) + 1
	sp.V = 1.0 - fmath.Acos(normal.Z)*(1.0/math.Pi)
	return
}
