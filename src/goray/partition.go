//
//  goray/partition.go
//  goray
//
//  Created by Ross Light on 2010-05-29.
//

/* The goray/partition package provides an interface for algorithms to efficiently handle ray-collision detection. */
package partition

import (
	"./goray/bound"
	"./goray/color"
	"./goray/primitive"
	"./goray/render"
	"./goray/ray"
	"./goray/vector"
)

/*
   Partitioner defines a type that can detect ray collisions.
   For most cases, this will involve an algorithm that partitions the scene to make these operations faster.
*/
type Partitioner interface {
	/* Intersect determines the primitive that a ray collides with. */
	Intersect(r ray.Ray, dist float) primitive.Collision
	/* IntersectS determines the primitive that a ray collides with for shadow-detection. */
	IntersectS(r ray.Ray, dist float) primitive.Collision
	/* IntersectTS computes the color of a transparent shadow after bouncing around. */
	IntersectTS(state *render.State, r ray.Ray, maxDepth int, dist float, filt *color.Color) primitive.Collision
	/* GetBound returns a bounding box that contains all of the primitives in the scene. */
	GetBound() *bound.Bound
}

type simple struct {
	prims []primitive.Primitive
	bound *bound.Bound
}

/*
   NewSimple creates a partitioner that doesn't split up the scene at all.
   This should only be used for debugging code, the complexity is O(N).
*/
func NewSimple(prims []primitive.Primitive) Partitioner {
	part := &simple{prims, prims[0].GetBound()}
	for _, p := range part.prims[1:] {
		part.bound = bound.Union(part.bound, p.GetBound())
	}
	return part
}

func (s *simple) GetBound() *bound.Bound { return s.bound }

func (s *simple) Intersect(r ray.Ray, dist float) (coll primitive.Collision) {
	for _, p := range s.prims {
		if coll = p.Intersect(r); coll.Hit() {
			if coll.RayDepth < dist && coll.RayDepth > r.TMin() {
				return
			}
		}
	}
	return primitive.Collision{}
}

func (s *simple) IntersectS(r ray.Ray, dist float) (coll primitive.Collision) {
	for _, p := range s.prims {
		if coll = p.Intersect(r); coll.Hit() {
			if coll.RayDepth < dist {
				return
			}
		}
	}
	return primitive.Collision{}
}

func (s *simple) IntersectTS(state *render.State, r ray.Ray, maxDepth int, dist float, filt *color.Color) (coll primitive.Collision) {
	depth := 0
	for _, p := range s.prims {
		if info := p.Intersect(r); info.Hit() && info.RayDepth < dist && info.RayDepth > r.TMin() {
			coll = info
			mat := coll.Primitive.GetMaterial()
			if !mat.IsTransparent() {
				return
			}
			if depth < maxDepth {
				h := vector.Add(r.From(), vector.ScalarMul(r.Dir(), coll.RayDepth))
				sp := coll.Primitive.GetSurface(h, coll.UserData)
				*filt = color.Mul(*filt, mat.GetTransparency(state, sp, r.Dir()))
				depth++
			} else {
				// We've hit the depth limit.  Cut it off.
				return
			}
		}
	}
	return
}
