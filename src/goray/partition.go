//
//  goray/partition.go
//  goray
//
//  Created by Ross Light on 2010-05-29.
//

package partition

import (
	"./goray/bound"
	"./goray/color"
	"./goray/primitive"
	"./goray/render"
	"./goray/ray"
	"./goray/vector"
)

type Partitioner interface {
	Intersect(r ray.Ray, dist float) (hit bool, prim primitive.Primitive, z float)
	IntersectS(r ray.Ray, dist float) (hit bool, prim primitive.Primitive)
	IntersectTS(state *render.State, r ray.Ray, maxDepth int, dist float, filt *color.Color) (hit bool, prim primitive.Primitive)
	GetBound() *bound.Bound
}

type simple struct {
	prims []primitive.Primitive
	bound *bound.Bound
}

func NewSimple(prims []primitive.Primitive) Partitioner {
	part := &simple{prims, prims[0].GetBound()}
	for _, p := range part.prims[1:] {
		part.bound = bound.Union(part.bound, p.GetBound())
	}
	return part
}

func (s *simple) GetBound() *bound.Bound { return s.bound }

func (s *simple) Intersect(r ray.Ray, dist float) (hit bool, prim primitive.Primitive, z float) {
	for _, p := range s.prims {
		if hit, z = p.Intersect(r); hit {
			if z < dist && z > r.TMin {
				prim = p
				return
			}
			hit = false
		}
	}
	return
}

func (s *simple) IntersectS(r ray.Ray, dist float) (hit bool, prim primitive.Primitive) {
	var z float
	for _, p := range s.prims {
		if hit, z = p.Intersect(r); hit {
			if z < dist {
				prim = p
				return
			}
			hit = false
		}
	}
	return
}

func (s *simple) IntersectTS(state *render.State, r ray.Ray, maxDepth int, dist float, filt *color.Color) (hit bool, prim primitive.Primitive) {
	depth := 0
	for _, p := range s.prims {
		if intersects, z := p.Intersect(r); intersects && z < dist && z > r.TMin {
			hit, prim = true, p
			mat := prim.GetMaterial()
			if !mat.IsTransparent() {
				return
			}
			if depth < maxDepth {
				h := vector.Add(r.From, vector.ScalarMul(r.Dir, z))
				sp := prim.GetSurface(h)
				*filt = color.Mul(*filt, mat.GetTransparency(state, sp, r.Dir))
				depth++
			} else {
				// We've hit the depth limit.  Cut it off.
				return
			}
		}
	}
	return
}
