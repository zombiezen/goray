//
//	goray/core/intersect/intersect.go
//	goray
//
//	Created by Ross Light on 2010-05-29.
//

/*
	The intersect package provides an interface and several algorithms for
	handling ray-collision detection.
*/
package intersect

import (
	"sort"
	"goray/logging"
	"goray/core/bound"
	"goray/core/color"
	"goray/core/kdtree"
	"goray/core/material"
	"goray/core/primitive"
	"goray/core/render"
	"goray/core/ray"
	"goray/core/vector"
)

// Interface defines a type that can detect ray collisions.
//
// For most cases, this will involve an algorithm that partitions the scene to make these operations faster.
type Interface interface {
	// Intersect determines the primitive that a ray collides with.
	Intersect(r ray.Ray, dist float64) primitive.Collision
	// IsShadowed checks whether a ray collides with any primitives (for shadow-detection).
	IsShadowed(r ray.Ray, dist float64) bool
	// DoTransparentShadows computes the color of a transparent shadow after bouncing around.
	DoTransparentShadows(state *render.State, r ray.Ray, maxDepth int, dist float64, filt *color.Color) bool
	// GetBound returns a bounding box that contains all of the primitives that the intersecter knows about.
	GetBound() *bound.Bound
}

type simple struct {
	prims []primitive.Primitive
	bound *bound.Bound
}

// NewSimple creates a partitioner that doesn't split up the scene at all.
// This should only be used for debugging code, the complexity is O(N).
func NewSimple(prims []primitive.Primitive) Interface {
	part := &simple{prims, prims[0].GetBound()}
	for _, p := range part.prims[1:] {
		part.bound = bound.Union(part.bound, p.GetBound())
	}
	return part
}

func (s *simple) GetBound() *bound.Bound { return s.bound }

func (s *simple) Intersect(r ray.Ray, dist float64) (coll primitive.Collision) {
	for _, p := range s.prims {
		if newColl := p.Intersect(r); newColl.Hit() {
			if newColl.RayDepth < dist && newColl.RayDepth > r.TMin && (!coll.Hit() || newColl.RayDepth < coll.RayDepth) {
				coll = newColl
			}
		}
	}
	return
}

func (s *simple) IsShadowed(r ray.Ray, dist float64) bool {
	for _, p := range s.prims {
		if newColl := p.Intersect(r); newColl.Hit() {
			return true
		}
	}
	return false
}

func (s *simple) DoTransparentShadows(state *render.State, r ray.Ray, maxDepth int, dist float64, filt *color.Color) bool {
	depth := 0
	for _, p := range s.prims {
		if coll := p.Intersect(r); coll.Hit() && coll.RayDepth < dist && coll.RayDepth > r.TMin {
			mat, trans := coll.Primitive.GetMaterial().(material.TransparentMaterial)
			if !trans {
				return true
			}
			if depth < maxDepth {
				sp := coll.Primitive.GetSurface(coll)
				*filt = color.Mul(*filt, mat.GetTransparency(state, sp, r.Dir))
				depth++
			} else {
				// We've hit the depth limit.  Cut it off.
				return false
			}
		}
	}
	return false
}

type kdPartition struct {
	*kdtree.Tree
}

func primGetDim(v kdtree.Value, axis vector.Axis) (min, max float64) {
	bd := v.(primitive.Primitive).GetBound()
	return bd.GetMin()[axis], bd.GetMax()[axis]
}

func NewKD(prims []primitive.Primitive, log logging.Handler) Interface {
	vals := make([]kdtree.Value, len(prims))
	for i, p := range prims {
		vals[i] = p
	}
	opts := kdtree.MakeOptions(primGetDim, log)
	tree := kdtree.New(vals, opts)
	return &kdPartition{tree}
}

type followFrame struct {
	node  kdtree.Node
	t     float64
	point vector.Vector3D
}

type collideList []primitive.Collision

func (cl collideList) Len() int { return len(cl) }

func (cl collideList) Less(a, b int) bool {
	return cl[a].RayDepth < cl[b].RayDepth
}

func (cl collideList) Swap(a, b int) {
	cl[a], cl[b] = cl[b], cl[a]
}

func (kd *kdPartition) followRay(r ray.Ray, minDist, maxDist float64, firstOnly bool, ch chan<- primitive.Collision) {
	defer close(ch)

	var a, b, t float64
	var hit bool

	if a, b, hit = kd.GetBound().Cross(r.From, r.Dir, maxDist); !hit {
		return
	}

	invDir := r.Dir.Inverse()
	enterStack := make([]followFrame, 0)
	{
		frame := followFrame{t: a}
		if a >= 0.0 {
			frame.point = vector.Add(r.From, vector.ScalarMul(r.Dir, a))
		} else {
			frame.point = r.From
		}
		enterStack = append(enterStack, frame)
	}

	exitStack := make([]followFrame, len(enterStack)+1)
	copy(exitStack, enterStack)
	exitStack[len(exitStack)-1] = followFrame{nil, b, vector.Add(r.From, vector.ScalarMul(r.Dir, b))}

	for currNode := kd.GetRoot(); currNode != nil && !closed(ch); {
		var farChild kdtree.Node
		// Stop looping if we've passed the maximum distance
		if enterStack[len(enterStack)-1].t > maxDist {
			break
		}
		// Traverse to the leaves
		for !currNode.IsLeaf() {
			currInter := currNode.(*kdtree.Interior)
			axis := currInter.GetAxis()
			pivot := currInter.GetPivot()

			if enterStack[len(enterStack)-1].point[axis] <= pivot {
				currNode = currInter.GetLeft()
				if exitStack[len(exitStack)-1].point[axis] <= pivot {
					continue
				}
				farChild = currInter.GetRight()
			} else {
				currNode = currInter.GetRight()
				if exitStack[len(exitStack)-1].point[axis] > pivot {
					continue
				}
				farChild = currInter.GetLeft()
			}

			t = (pivot - r.From[axis]) * invDir[axis]

			// Set up the new exit point
			var pt vector.Vector3D
			prevAxis, nextAxis := (axis+1)%3, (axis+2)%3
			pt[axis] = pivot
			pt[nextAxis] = r.From[nextAxis] + t*r.Dir[nextAxis]
			pt[prevAxis] = r.From[prevAxis] + t*r.Dir[prevAxis]
			frame := followFrame{farChild, t, pt}
			exitStack = append(exitStack, frame)
		}

		// Okay, we've reached a leaf.
		// Now check for any intersections.
		prims := currNode.(*kdtree.Leaf).GetValues()
		if firstOnly {
			var firstColl primitive.Collision
			for _, v := range prims {
				p := v.(primitive.Primitive)
				coll := p.Intersect(r)
				if coll.Hit() && coll.RayDepth > minDist && coll.RayDepth < maxDist && (!firstColl.Hit() || coll.RayDepth < firstColl.RayDepth) {
					firstColl = coll
				}
			}

			if firstColl.Hit() {
				ch <- firstColl
				return
			}
		} else {
			cl := make(collideList, 0, len(prims))
			for _, v := range prims {
				p := v.(primitive.Primitive)
				if coll := p.Intersect(r); coll.Hit() && coll.RayDepth > minDist && coll.RayDepth < maxDist {
					cl = append(cl, coll)
				}
			}
			// Yield the collisions in order.
			sort.Sort(cl)
			for _, coll := range cl {
				ch <- coll
			}
		}

		// Update stack
		if cap(enterStack) < len(exitStack) {
			enterStack = make([]followFrame, len(exitStack))
		} else {
			enterStack = enterStack[:len(exitStack)]
		}
		copy(enterStack, exitStack)

		currNode = exitStack[len(exitStack)-1].node
		exitStack = exitStack[:len(exitStack)-1]
	}
}

func (kd *kdPartition) Intersect(r ray.Ray, dist float64) (coll primitive.Collision) {
	ch := make(chan primitive.Collision)
	go kd.followRay(r, r.TMin, dist, true, ch)
	return <-ch
}

func (kd *kdPartition) IsShadowed(r ray.Ray, dist float64) bool {
	ch := make(chan primitive.Collision)
	go kd.followRay(r, r.TMin, dist, true, ch)
	coll := <-ch
	return coll.Hit()
}

func (kd *kdPartition) DoTransparentShadows(state *render.State, r ray.Ray, maxDepth int, dist float64, filt *color.Color) bool {
	ch := make(chan primitive.Collision)
	defer close(ch)

	go kd.followRay(r, r.TMin, dist, false, ch)
	depth := 0
	hitList := make(map[primitive.Primitive]bool)
	for coll := range ch {
		mat, trans := coll.Primitive.GetMaterial().(material.TransparentMaterial)
		if !trans {
			return true
		}
		if hit, _ := hitList[coll.Primitive]; !hit {
			hitList[coll.Primitive] = true
			if depth >= maxDepth {
				return false
			}
			sp := coll.Primitive.GetSurface(coll)
			*filt = color.Mul(*filt, mat.GetTransparency(state, sp, r.Dir))
			depth++
		}
	}
	return false
}
