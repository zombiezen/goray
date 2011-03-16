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
	node  *kdtree.Node
	t     float64
	point vector.Vector3D
}

type kdFollower struct {
	Ray              ray.Ray
	MaxDist, MinDist float64

	currNode              *kdtree.Node
	enterStack, exitStack []followFrame
}

func (f *kdFollower) Init(kd *kdtree.Tree) {
	a, b, hit := kd.GetBound().Cross(f.Ray.From, f.Ray.Dir, f.MaxDist)
	if !hit {
		f.currNode = nil
		return
	}

	if f.enterStack == nil {
		f.enterStack = make([]followFrame, 0, 10)
	} else {
		f.enterStack = f.enterStack[:0]
	}
	f.enterStack = append(f.enterStack, followFrame{t: a, point: f.Ray.From})
	if a > 0 {
		// XXX: May not align exactly with box
		f.enterStack[0].point = vector.Add(f.enterStack[0].point, vector.ScalarMul(f.Ray.Dir, a))
	}

	if f.exitStack == nil {
		f.exitStack = make([]followFrame, 0, 10)
	} else {
		f.exitStack = f.exitStack[:0]
	}
	f.exitStack = append(
		f.exitStack,
		f.enterStack[0],
		followFrame{t: b, point: vector.Add(f.Ray.From, vector.ScalarMul(f.Ray.Dir, b))},
	)

	f.currNode = kd.GetRoot()
}

func (f *kdFollower) findLeaf() bool {
	if f.enterStack[len(f.enterStack)-1].t > f.MaxDist {
		return false
	}
	for !f.currNode.Leaf() {
		var farChild *kdtree.Node
		axis, pivot := f.currNode.Axis(), f.currNode.Pivot()
		// TODO: Check logic
		if f.enterStack[len(f.enterStack)-1].point[axis] < pivot {
			if f.exitStack[len(f.exitStack)-1].point[axis] < pivot {
				f.currNode = f.currNode.Left()
				continue
			}
			farChild = f.currNode.Right()
			f.currNode = f.currNode.Left()
		} else {
			if f.exitStack[len(f.exitStack)-1].point[axis] >= pivot {
				f.currNode = f.currNode.Right()
				continue
			}
			farChild = f.currNode.Left()
			f.currNode = f.currNode.Right()
		}

		t := (pivot - f.Ray.From[axis]) * f.Ray.Dir.Inverse()[axis]

		// Set up new exit point
		var pt vector.Vector3D
		pAxis, nAxis := axis.Prev(), axis.Next()
		pt[axis] = pivot
		pt[nAxis] = f.Ray.From[nAxis] + t*f.Ray.Dir[nAxis]
		pt[pAxis] = f.Ray.From[pAxis] + t*f.Ray.Dir[pAxis]
		f.exitStack = append(f.exitStack, followFrame{farChild, t, pt})
	}
	return true
}

func (f *kdFollower) pop() {
	if cap(f.enterStack) < len(f.exitStack) {
		f.enterStack = make([]followFrame, len(f.exitStack))
	} else {
		f.enterStack = f.enterStack[:len(f.exitStack)]
	}
	copy(f.enterStack, f.exitStack)
	f.currNode = f.exitStack[len(f.exitStack)-1].node
	f.exitStack = f.exitStack[:len(f.exitStack)-1]
}

func (f *kdFollower) First() (firstColl primitive.Collision) {
	for f.currNode != nil {
		if !f.findLeaf() {
			break
		}
		for _, v := range f.currNode.Values() {
			p := v.(primitive.Primitive)
			coll := p.Intersect(f.Ray)
			if coll.Hit() && coll.RayDepth > f.MinDist && coll.RayDepth < f.MaxDist && (!firstColl.Hit() || coll.RayDepth < firstColl.RayDepth) {
				firstColl = coll
			}
		}
		if firstColl.Hit() {
			return
		}
		f.pop()
	}
	return
}

func (kd *kdPartition) followRayFull(r ray.Ray, minDist, maxDist float64, ch chan<- primitive.Collision) {
	a, b, hit := kd.GetBound().Cross(r.From, r.Dir, maxDist)
	if !hit {
		return
	}

	invDir := r.Dir.Inverse()
	enterStack := []followFrame{
		{t: a, point: r.From},
	}
	if a >= 0.0 {
		enterStack[0].point = vector.Add(enterStack[0].point, vector.ScalarMul(r.Dir, a))
	}

	exitStack := []followFrame{
		enterStack[0],
		{t: b, point: vector.Add(r.From, vector.ScalarMul(r.Dir, b))},
	}

	for currNode := kd.GetRoot(); currNode != nil; {
		// Stop looping if we've passed the maximum distance
		if enterStack[len(enterStack)-1].t > maxDist {
			break
		}
		// Traverse to the leaves
		for !currNode.Leaf() {
			axis, pivot := currNode.Axis(), currNode.Pivot()

			var farChild *kdtree.Node
			if enterStack[len(enterStack)-1].point[axis] < pivot {
				farChild = currNode.Right()
				currNode = currNode.Left()
				if exitStack[len(exitStack)-1].point[axis] < pivot {
					continue
				}
			} else {
				farChild = currNode.Left()
				currNode = currNode.Right()
				if exitStack[len(exitStack)-1].point[axis] >= pivot {
					continue
				}
			}

			t := (pivot - r.From[axis]) * invDir[axis]

			// Set up the new exit point
			var pt vector.Vector3D
			prevAxis, nextAxis := (axis+1)%3, (axis+2)%3
			pt[axis] = pivot
			pt[nextAxis] = r.From[nextAxis] + t*r.Dir[nextAxis]
			pt[prevAxis] = r.From[prevAxis] + t*r.Dir[prevAxis]
			exitStack = append(exitStack, followFrame{farChild, t, pt})
		}

		// Okay, we've reached a leaf.
		// Now check for any intersections.
		prims := currNode.Values()
		clist := make([]primitive.Collision, 0, len(prims))
		for _, v := range prims {
			p := v.(primitive.Primitive)
			if coll := p.Intersect(r); coll.Hit() && coll.RayDepth > minDist && coll.RayDepth < maxDist {
				clist = append(clist, coll)
				// Move new collision to proper location (insertion sort while inserting! :D)
				var i int
				for i = len(clist) - 1; i > 0 && coll.RayDepth < clist[i-1].RayDepth; i-- {
					clist[i] = clist[i-1]
				}
				clist[i] = coll
			}
		}
		// Yield the collisions in order.
		for _, coll := range clist {
			ch <- coll
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
	f := &kdFollower{Ray: r, MinDist: r.TMin, MaxDist: dist}
	f.Init(kd.Tree)
	return f.First()
}

func (kd *kdPartition) IsShadowed(r ray.Ray, dist float64) bool {
	f := &kdFollower{Ray: r, MinDist: r.TMin, MaxDist: dist}
	f.Init(kd.Tree)
	// TODO: Only check for a single collision
	return f.First().Hit()
}

func (kd *kdPartition) DoTransparentShadows(state *render.State, r ray.Ray, maxDepth int, dist float64, filt *color.Color) bool {
	ch := make(chan primitive.Collision)
	defer close(ch)

	go kd.followRayFull(r, r.TMin, dist, ch)
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
