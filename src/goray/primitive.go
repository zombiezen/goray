//
//  goray/primitive.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

/* The goray/primitive package provides the basic components of a scene. */
package primitive

import (
	"math"
	"./fmath"
	"./goray/bound"
	"./goray/material"
	"./goray/ray"
	"./goray/surface"
	"./goray/vector"
)

/* Primitive defines a basic 3D entity in a scene. */
type Primitive interface {
	/* GetBound returns the bounding box in global (world) coordinates. */
	GetBound() *bound.Bound
	/*
	   IntersectsBound returns whether a bounding box intersects the primitive.
	   This can be used to implement more precise partitioning.
	*/
	IntersectsBound(*bound.Bound) bool
	/* HasClippingSupport indicates if the object has a clipping implementation. */
	HasClippingSupport() bool
	/*
	   ClipToBound calculates the overlapping bounding box of a given bounding box and the primitive.
	   It returns true only if a valid clip exists.
	*/
	ClipToBound(bound [2][3]float, axis int) (clipped *bound.Bound, ok bool)
	/*
	   Intersect checks whether a ray collides with the primitive.
	   This should not skip intersections outside of [TMin, TMax].
	*/
	Intersect(ray ray.Ray) (raydepth float, hit bool)
	/* GetSurface obtains information about a point on the primitive's surface. */
	GetSurface(pt vector.Vector3D) surface.Point
	/* GetMaterial returns the material associated with this primitive. */
	GetMaterial() material.Material
}

type sphere struct {
	center   vector.Vector3D
	radius   float
	material material.Material
}

/* NewSphere creates a spherical primitive. */
func NewSphere(center vector.Vector3D, radius float, material material.Material) Primitive {
	return &sphere{center, radius, material}
}

func (s *sphere) GetBound() *bound.Bound {
	r := vector.Vector3D{s.radius * 1.0001, s.radius * 1.0001, s.radius * 1.0001}
	return bound.New(vector.Sub(s.center, r), vector.Add(s.center, r))
}

func (s *sphere) IntersectsBound(b *bound.Bound) bool { return true }
func (s *sphere) HasClippingSupport() bool            { return false }
func (s *sphere) GetMaterial() material.Material      { return s.material }

func (s *sphere) ClipToBound(b [2][3]float, axis int) (*bound.Bound, bool) {
	return nil, false
}

func (s *sphere) Intersect(ray ray.Ray) (raydepth float, hit bool) {
	vf := vector.Sub(ray.From, s.center)
	ea := ray.Dir.LengthSqr()
	eb := vector.Dot(vf, ray.Dir) * 2.0
	ec := vf.LengthSqr() - s.radius*s.radius
	osc := eb*eb - 4.0*ea*ec
	if osc < 0 {
		hit = false
		return
	}

	osc = fmath.Sqrt(osc)
	sol1 := (-eb - osc) / (ea * 2.0)
	sol2 := (-eb + osc) / (ea * 2.0)
	raydepth = sol1
	if raydepth < ray.TMin {
		raydepth = sol2
		if raydepth < ray.TMin {
			hit = false
			raydepth = 0.0
			return
		}
	}
	hit = true
	return
}

func (s *sphere) GetSurface(pt vector.Vector3D) (sp surface.Point) {
	normal := vector.Sub(pt, s.center)
	sp.OrcoPosition = normal
	normal = normal.Normalize()
	sp.Material = s.material
	sp.Normal = normal
	sp.GeometricNormal = normal
	sp.HasOrco = true
	sp.Position = pt
	sp.U = fmath.Atan2(normal.Y, normal.X)*(1.0/math.Pi) + 1
	sp.V = 1.0 - fmath.Acos(normal.Z)*(1.0/math.Pi)
	sp.Light = nil
	return
}
