//
//  goray/core/primitive.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

/* The primitive package provides the basic components of a scene. */
package primitive

import (
	"goray/core/bound"
	"goray/core/material"
	"goray/core/ray"
	"goray/core/surface"
	"goray/core/vector"
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
	/*
	   Intersect checks whether a ray collides with the primitive.
	   This should not skip intersections outside of [TMin, TMax].
	*/
	Intersect(r ray.Ray) Collision
	/*
	   GetSurface obtains information about a point on the primitive's surface.

	   You can only get the surface point by ray casting to it.  Admittedly, this is slightly inflexible,
	   but it's the only use-case for this method.  The advantage is that Intersect can pass any extra data
	   that it could need to efficiently implement GetSurface in the Collision struct.
	*/
	GetSurface(Collision) surface.Point
	/* GetMaterial returns the material associated with this primitive. */
	GetMaterial() material.Material
}

type Clippable interface {
	/*
	   ClipToBound calculates the overlapping bounding box of a given bounding box and the primitive.
	   It returns true only if a valid clip exists.
	*/
	ClipToBound(bound [2][3]float, axis int) (clipped *bound.Bound, ok bool)
}

type ClippablePrimitive interface {
	Primitive
	Clippable
}
