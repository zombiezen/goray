//
//  goray/core/partition.go
//  goray
//
//  Created by Ross Light on 2010-05-29.
//

/* The partition package provides an interface for algorithms to efficiently handle ray-collision detection. */
package partition

import (
	"goray/logging"
	"goray/stack"
)

import (
	"goray/core/bound"
	"goray/core/color"
	"goray/core/kdtree"
	"goray/core/primitive"
	"goray/core/render"
	"goray/core/ray"
	"goray/core/vector"
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
		if newColl := p.Intersect(r); newColl.Hit() {
			if newColl.RayDepth < dist && newColl.RayDepth > r.TMin() && (!coll.Hit() || newColl.RayDepth < coll.RayDepth) {
				coll = newColl
			}
		}
	}
	return
}

func (s *simple) IntersectS(r ray.Ray, dist float) (coll primitive.Collision) {
	for _, p := range s.prims {
		if newColl := p.Intersect(r); newColl.Hit() {
			if newColl.RayDepth < dist && newColl.RayDepth > r.TMin() {
				coll = newColl
				return
			}
		}
	}
	return
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
				sp := coll.Primitive.GetSurface(coll)
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

type kdPartition struct {
	*kdtree.Tree
}

func primGetDim(v kdtree.Value, axis int) (min, max float) {
	bd := v.(primitive.Primitive).GetBound()
	switch axis {
	case 0:
		min, max = bd.GetMinX(), bd.GetMaxX()
	case 1:
		min, max = bd.GetMinY(), bd.GetMaxY()
	case 2:
		min, max = bd.GetMinZ(), bd.GetMaxZ()
	}
	return
}

func NewKD(prims []primitive.Primitive, log *logging.Logger) Partitioner {
	vals := make([]kdtree.Value, len(prims))
	for i, p := range prims {
		vals[i] = p
	}
	tree := kdtree.New(vals, primGetDim, log)
	return &kdPartition{tree}
}

type followFrame struct {
	node  kdtree.Node
	t     float
	point vector.Vector3D
}

func (kd *kdPartition) followRay(r ray.Ray, minDist, maxDist float, ch chan<- primitive.Collision) {
	defer close(ch)

	var a, b, t float
	var hit bool

	if a, b, hit = kd.GetBound().Cross(r.From(), r.Dir(), maxDist); !hit {
		return
	}

	invDir := r.Dir().Inverse()
	enterStack := stack.New()
	{
		frame := followFrame{t: a}
		if a >= 0.0 {
			frame.point = vector.Add(r.From(), vector.ScalarMul(r.Dir(), a))
		} else {
			frame.point = r.From()
		}
		enterStack.Push(frame)
	}

	exitStack := enterStack.Copy()
	exitStack.Push(followFrame{nil, b, vector.Add(r.From(), vector.ScalarMul(r.Dir(), b))})

	enter := func() followFrame {
		frame := enterStack.Top()
		return frame.(followFrame)
	}
	exit := func() followFrame {
		frame := exitStack.Top()
		return frame.(followFrame)
	}

	for currNode := kd.GetRoot(); currNode != nil; {
		var farChild kdtree.Node
		// Stop looping if we've passed the maximum distance
		if enter().t > maxDist {
			break
		}
		// Traverse to the leaves
		for !currNode.IsLeaf() {
			currInter := currNode.(*kdtree.Interior)
			axis := currInter.GetAxis()
			pivot := currInter.GetPivot()

			if enter().point.GetComponent(axis) <= pivot {
				currNode = currInter.GetLeft()
				if exit().point.GetComponent(axis) <= pivot {
					continue
				}
				farChild = currInter.GetRight()
			} else {
				currNode = currInter.GetRight()
				if exit().point.GetComponent(axis) > pivot {
					continue
				}
				farChild = currInter.GetLeft()
			}

			t = (pivot - r.From().GetComponent(axis)) * invDir.GetComponent(axis)

			// Set up the new exit point
			var pt [3]float
			prevAxis, nextAxis := (axis+1)%3, (axis+2)%3
			pt[axis] = pivot
			pt[nextAxis] = r.From().GetComponent(nextAxis) + t*r.Dir().GetComponent(nextAxis)
			pt[prevAxis] = r.From().GetComponent(prevAxis) + t*r.Dir().GetComponent(prevAxis)
			frame := followFrame{farChild, t, vector.New(pt[0], pt[1], pt[2])}
			exitStack.Push(frame)
		}

		// Okay, we've reached a leaf.
		// Now check for any intersections.
		for _, v := range currNode.(*kdtree.Leaf).GetValues() {
			p := v.(primitive.Primitive)
			if coll := p.Intersect(r); coll.Hit() && coll.RayDepth > minDist && coll.RayDepth < maxDist {
				ch <- coll
			}
		}

		// Update stack
		enterStack = exitStack.Copy()
		topExit := exitStack.Pop()
		currNode = topExit.(followFrame).node
	}
}

func (kd *kdPartition) Intersect(r ray.Ray, dist float) (coll primitive.Collision) {
	ch := make(chan primitive.Collision)
	go kd.followRay(r, r.TMin(), dist, ch)
	for newColl := range ch {
		if !coll.Hit() || newColl.RayDepth < coll.RayDepth {
			coll = newColl
		}
	}
	return
}

func (kd *kdPartition) IntersectS(r ray.Ray, dist float) (coll primitive.Collision) {
	ch := make(chan primitive.Collision)
	go kd.followRay(r, r.TMin(), dist, ch)
	for newColl := range ch {
		if !coll.Hit() || newColl.RayDepth < coll.RayDepth {
			coll = newColl
		}
	}
	return
}

func (kd *kdPartition) IntersectTS(state *render.State, r ray.Ray, maxDepth int, dist float, filt *color.Color) (coll primitive.Collision) {
	ch := make(chan primitive.Collision)

	go kd.followRay(r, r.TMin(), dist, ch)
	depth := 0
	hitList := make(map[primitive.Primitive]bool)
	for coll = range ch {
		mat := coll.Primitive.GetMaterial()
		if !mat.IsTransparent() {
			return
		}
		if hit, _ := hitList[coll.Primitive]; !hit {
			hitList[coll.Primitive] = true
			if depth >= maxDepth {
				return
			}
			sp := coll.Primitive.GetSurface(coll)
			*filt = color.Mul(*filt, mat.GetTransparency(state, sp, r.Dir()))
			depth++
		}
	}
	return
}
