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
	// Shadowed checks whether a ray collides with any primitives (for shadow-detection).
	Shadowed(r ray.Ray, dist float64) bool
	// TransparentShadow computes the color of a transparent shadow after bouncing around.
	TransparentShadow(state *render.State, r ray.Ray, maxDepth int, dist float64, filt *color.Color) (opaque bool)
	// Bound returns a bounding box that contains all of the primitives that the intersecter knows about.
	Bound() *bound.Bound
}

type simple struct {
	prims []primitive.Primitive
	bound *bound.Bound
}

// NewSimple creates a partitioner that doesn't split up the scene at all.
// This should only be used for debugging code, the complexity is O(N).
func NewSimple(prims []primitive.Primitive) Interface {
	part := &simple{prims: prims}
	if len(prims) == 0 {
		part.bound = bound.New(vector.Vector3D{}, vector.Vector3D{})
		return part
	}
	part.bound = part.prims[0].Bound()
	for _, p := range part.prims[1:] {
		part.bound = bound.Union(part.bound, p.Bound())
	}
	return part
}

func (s *simple) Bound() *bound.Bound { return s.bound }

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
				return true
			}
		}
	}
	return false
}

type kdPartition struct {
	*kdtree.Tree
}

func primGetDim(v kdtree.Value, axis vector.Axis) (min, max float64) {
	bd := v.(primitive.Primitive).Bound()
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
	a, b, hit := kd.Bound().Cross(f.Ray.From, f.Ray.Dir, f.MaxDist)
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

func (f *kdFollower) Hit() bool {
	for f.currNode != nil {
		if !f.findLeaf() {
			break
		}
		for _, v := range f.currNode.Values() {
			p := v.(primitive.Primitive)
			coll := p.Intersect(f.Ray)
			if coll.Hit() && coll.RayDepth > f.MinDist && coll.RayDepth < f.MaxDist {
				return true
			}
		}
		f.pop()
	}
	return false
}

type kdTranspFollower struct {
	kdFollower
	hitList   map[primitive.Primitive]bool
	currPrims []primitive.Primitive
}

func (f *kdTranspFollower) Init(kd *kdtree.Tree) {
	f.kdFollower.Init(kd)
	f.hitList = make(map[primitive.Primitive]bool)
	if f.currPrims == nil {
		f.currPrims = make([]primitive.Primitive, 0, 10)
	} else {
		f.currPrims = f.currPrims[:0]
	}
}

func (f *kdTranspFollower) findMore() {
	for f.currNode != nil && len(f.currPrims) == 0 {
		if !f.findLeaf() {
			break
		}
		vals := f.currNode.Values()
		for _, v := range vals {
			p := v.(primitive.Primitive)
			if f.hitList[p] {
				continue
			}
			f.currPrims = append(f.currPrims, p)
			f.hitList[p] = true
		}
		f.pop()
	}
}

func (f *kdTranspFollower) Next() (coll primitive.Collision) {
	for f.findMore(); len(f.currPrims) > 0; f.findMore() {
		p := f.currPrims[len(f.currPrims)-1]
		f.currPrims = f.currPrims[:len(f.currPrims)-1]
		if coll = p.Intersect(f.Ray); coll.Hit() && coll.RayDepth > f.MinDist && coll.RayDepth < f.MaxDist {
			return
		}
	}
	return
}

func (kd *kdPartition) Intersect(r ray.Ray, dist float64) (coll primitive.Collision) {
	f := &kdFollower{Ray: r, MinDist: r.TMin, MaxDist: dist}
	f.Init(kd.Tree)
	return f.First()
}

func (kd *kdPartition) IsShadowed(r ray.Ray, dist float64) bool {
	f := &kdFollower{Ray: r, MinDist: r.TMin, MaxDist: dist}
	f.Init(kd.Tree)
	return f.Hit()
}

func (kd *kdPartition) DoTransparentShadows(state *render.State, r ray.Ray, maxDepth int, dist float64, filt *color.Color) (hitOpaque bool) {
	f := &kdTranspFollower{kdFollower: kdFollower{Ray: r, MinDist: r.TMin, MaxDist: dist}}
	f.Init(kd.Tree)
	depth := 0
	for coll := f.Next(); coll.Hit(); coll = f.Next() {
		if depth >= maxDepth {
			// Too much depth, just say it's opaque.
			return true
		}
		tmat, ok := coll.Primitive.GetMaterial().(material.TransparentMaterial)
		if !ok {
			// Material is opaque
			return true
		}
		*filt = color.Mul(*filt, tmat.GetTransparency(state, coll.GetSurface(), f.Ray.Dir))
		depth++
	}
	return
}
