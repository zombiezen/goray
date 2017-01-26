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
	"zombiezen.com/go/goray/internal/bound"
	"zombiezen.com/go/goray/internal/vecutil"
)

// Collision stores information about a ray intersection.
type Collision struct {
	Primitive Primitive
	Ray       Ray
	RayDepth  float64
	UserData  interface{}
}

// Hit returns whether the ray intersection succeeded.
// If this method returns false, then the rest of the structure is meaningless.
func (c Collision) Hit() bool { return c.Primitive != nil }

// Point returns the point in world coordinates where the ray intersected.
func (c Collision) Point() vec64.Vector {
	return vec64.Add(c.Ray.From, c.Ray.Dir.Scale(c.RayDepth))
}

// Surface returns the surface point where the ray intersected.
func (c Collision) Surface() (sp SurfacePoint) {
	sp = c.Primitive.Surface(c)
	sp.Primitive = c.Primitive
	return
}

// Primitive defines a basic 3D entity in a scene.
type Primitive interface {
	// Bound returns the bounding box in global (world) coordinates.
	Bound() bound.Bound

	// IntersectsBound returns whether a bounding box intersects the primitive.
	// This can be used to implement more precise partitioning.
	IntersectsBound(bound.Bound) bool

	// Intersect checks whether a ray collides with the primitive.
	// This should not skip intersections outside of [TMin, TMax].
	Intersect(r Ray) Collision

	// Surface obtains information about a point on the primitive's surface.
	//
	// You can only get the surface point by ray casting to it.  Admittedly, this is slightly inflexible,
	// but it's the only use-case for this method.  The advantage is that Intersect can pass any extra data
	// that it could need to efficiently implement GetSurface in the Collision struct.
	Surface(Collision) SurfacePoint

	// Material returns the material associated with this primitive.
	Material() Material
}

type Clipper interface {
	// Clip calculates the overlapping bounding box of a given bounding box and the primitive.
	// If the bounding box returned is nil, then no such bound exists.
	Clip(bound bound.Bound, axis vecutil.Axis, lower bool, oldData interface{}) (clipped bound.Bound, newData interface{})
}

type ClipPrimitive interface {
	Primitive
	Clipper
}
