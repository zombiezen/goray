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

/* Collision stores information about a ray intersection. */
type Collision struct {
	Primitive Primitive
	Ray       ray.Ray
	RayDepth  float
	UserData  interface{}
}

/*
   Hit returns whether the ray intersection succeeded.
   If this method returns false, then the rest of the structure is meaningless.
*/
func (c Collision) Hit() bool { return c.Primitive != nil }

/* GetPoint returns the point in world coordinates where the ray intersected. */
func (c Collision) GetPoint() vector.Vector3D {
    return vector.Add(c.Ray.From(), vector.ScalarMul(c.Ray.Dir(), c.RayDepth))
}

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
	Intersect(r ray.Ray) Collision
	/* GetSurface obtains information about a point on the primitive's surface. */
	GetSurface(pt vector.Vector3D, userData interface{}) surface.Point
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

func (s *sphere) Intersect(r ray.Ray) (coll Collision) {
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

func (s *sphere) GetSurface(pt vector.Vector3D, userdata interface{}) (sp surface.Point) {
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
