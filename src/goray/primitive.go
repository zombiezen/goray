//
//  goray/primitive.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

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

type Primitive interface {
	// Return the bound in global (world) coordinates
	GetBound() *bound.Bound
	IntersectsBound(*bound.Bound) bool
	HasClippingSupport() bool
	ClipToBound(bound [2][3]float, axis int) (clipped *bound.Bound, ok bool)
	Intersect(ray ray.Ray) (hit bool, raydepth float)
	GetSurface(pt vector.Vector3D) surface.Point
	GetMaterial() material.Material
}

type sphere struct {
	center   vector.Vector3D
	radius   float
	material material.Material
}

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

func (s *sphere) Intersect(ray ray.Ray) (hit bool, raydepth float) {
	vf := vector.Sub(ray.From, s.center)
	ea := vector.Dot(ray.Dir, ray.Dir)
	eb := vector.Dot(vf, ray.Dir) * 2.0
	ec := vector.Dot(vf, vf) - s.radius*s.radius
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
	//sp.Light = nil
	return
}
