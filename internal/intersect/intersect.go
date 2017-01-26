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

// Package intersect contains algorithms for handling ray-collision detection.
package intersect

import (
	"bitbucket.org/zombiezen/math3/vec64"
	"zombiezen.com/go/goray/internal/bound"
	"zombiezen.com/go/goray/internal/color"
	"zombiezen.com/go/goray/internal/goray"
	"zombiezen.com/go/goray/internal/kdtree"
	"zombiezen.com/go/goray/internal/log"
	"zombiezen.com/go/goray/internal/vecutil"
)

type simple struct {
	prims []goray.Primitive
	goray bound.Bound
}

// NewSimple creates a partitioner that doesn't split up the scene at all.
// This should only be used for debugging code, the complexity is O(N).
func NewSimple(prims []goray.Primitive) goray.Intersecter {
	part := &simple{prims: prims}
	if len(prims) == 0 {
		part.goray = bound.Bound{}
		return part
	}
	part.goray = part.prims[0].Bound()
	for _, p := range part.prims[1:] {
		part.goray = bound.Union(part.goray, p.Bound())
	}
	return part
}

func (s *simple) Bound() bound.Bound { return s.goray }

func (s *simple) Intersect(r goray.Ray, dist float64) (coll goray.Collision) {
	for _, p := range s.prims {
		if newColl := p.Intersect(r); newColl.Hit() {
			if newColl.RayDepth < dist && newColl.RayDepth > r.TMin && (!coll.Hit() || newColl.RayDepth < coll.RayDepth) {
				coll = newColl
			}
		}
	}
	return
}

func (s *simple) Shadowed(r goray.Ray, dist float64) bool {
	for _, p := range s.prims {
		if newColl := p.Intersect(r); newColl.Hit() {
			return true
		}
	}
	return false
}

func (s *simple) TransparentShadow(state *goray.RenderState, r goray.Ray, maxDepth int, dist float64) (filt color.Color, hit bool) {
	depth := 0
	filt = color.White
	for _, p := range s.prims {
		if coll := p.Intersect(r); coll.Hit() && coll.RayDepth < dist && coll.RayDepth > r.TMin {
			mat, trans := coll.Primitive.Material().(goray.TransparentMaterial)
			if !trans {
				return color.Black, true
			}
			if depth < maxDepth {
				filt = color.Mul(filt, mat.Transparency(state, coll.Surface(), r.Dir))
				depth++
			} else {
				// We've hit the depth limit.  Cut it off.
				return color.Black, true
			}
		}
	}
	return
}

type kdPartition struct {
	*kdtree.Tree
	prims []goray.Primitive
	tris  []*goray.Triangle
}

func (kd *kdPartition) Len() int {
	return len(kd.prims)
}

func (kd *kdPartition) Dimension(i int, axis vecutil.Axis) (min, max float64) {
	bd := kd.prims[i].Bound()
	return bd.Min[axis], bd.Max[axis]
}

func (kd *kdPartition) Clip(i int, bound bound.Bound, axis vecutil.Axis, lower bool, data interface{}) (bound.Bound, interface{}) {
	if clipper, ok := kd.prims[i].(goray.Clipper); ok {
		return clipper.Clip(bound, axis, lower, data)
	}
	return bound, data
}

func NewKD(prims []goray.Primitive, log log.Logger) goray.Intersecter {
	kd := &kdPartition{
		prims: prims,
		tris:  make([]*goray.Triangle, len(prims)),
	}
	for i := range prims {
		if tri, ok := prims[i].(*goray.Triangle); ok {
			kd.tris[i] = tri
		} else {
			kd.tris = nil
			break
		}
	}
	kd.Tree = kdtree.New(kd, kdtree.DefaultOptions)
	return kd
}

type followFrame struct {
	node  *kdtree.Node
	t     float64
	point vec64.Vector
}

type kdFollower struct {
	Partition        *kdPartition
	Ray              goray.Ray
	MaxDist, MinDist float64

	currNode              *kdtree.Node
	enterStack, exitStack []followFrame
}

func (f *kdFollower) Init(kd *kdPartition) {
	f.Partition = kd
	a, b, hit := kd.Tree.Bound().Cross(f.Ray.From, f.Ray.Dir, f.MaxDist)
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
		f.enterStack[0].point = vec64.Add(f.enterStack[0].point, f.Ray.Dir.Scale(a))
	}

	if f.exitStack == nil {
		f.exitStack = make([]followFrame, 0, 10)
	} else {
		f.exitStack = f.exitStack[:0]
	}
	f.exitStack = append(
		f.exitStack,
		f.enterStack[0],
		followFrame{t: b, point: vec64.Add(f.Ray.From, f.Ray.Dir.Scale(b))},
	)

	f.currNode = kd.Tree.Root()
}

func (f *kdFollower) findLeaf() bool {
	if f.enterStack[len(f.enterStack)-1].t > f.MaxDist {
		return false
	}
	for !f.currNode.IsLeaf() {
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
		var pt vec64.Vector
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

func (f *kdFollower) First() (firstColl goray.Collision) {
	for f.currNode != nil {
		if !f.findLeaf() {
			break
		}
		if f.Partition.tris != nil {
			for _, i := range f.currNode.Indices() {
				coll := f.Partition.tris[i].Intersect(f.Ray)
				if coll.Hit() && coll.RayDepth > f.MinDist && coll.RayDepth < f.MaxDist && (!firstColl.Hit() || coll.RayDepth < firstColl.RayDepth) {
					firstColl = coll
				}
			}
		} else {
			for _, i := range f.currNode.Indices() {
				coll := f.Partition.prims[i].Intersect(f.Ray)
				if coll.Hit() && coll.RayDepth > f.MinDist && coll.RayDepth < f.MaxDist && (!firstColl.Hit() || coll.RayDepth < firstColl.RayDepth) {
					firstColl = coll
				}
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
		if f.Partition.tris != nil {
			for _, i := range f.currNode.Indices() {
				coll := f.Partition.tris[i].Intersect(f.Ray)
				if coll.Hit() && coll.RayDepth > f.MinDist && coll.RayDepth < f.MaxDist {
					return true
				}
			}
		} else {
			for _, i := range f.currNode.Indices() {
				coll := f.Partition.prims[i].Intersect(f.Ray)
				if coll.Hit() && coll.RayDepth > f.MinDist && coll.RayDepth < f.MaxDist {
					return true
				}
			}
		}
		f.pop()
	}
	return false
}

type kdTranspFollower struct {
	kdFollower
	hitList   map[int]bool
	currPrims []goray.Primitive
}

func (f *kdTranspFollower) Init(kd *kdPartition) {
	f.kdFollower.Init(kd)
	f.hitList = make(map[int]bool)
	if f.currPrims == nil {
		f.currPrims = make([]goray.Primitive, 0, 10)
	} else {
		f.currPrims = f.currPrims[:0]
	}
}

func (f *kdTranspFollower) findMore() {
	for f.currNode != nil && len(f.currPrims) == 0 {
		if !f.findLeaf() {
			break
		}
		indices := f.currNode.Indices()
		for _, i := range indices {
			p := f.Partition.prims[i]
			if f.hitList[i] {
				continue
			}
			f.currPrims = append(f.currPrims, p)
			f.hitList[i] = true
		}
		f.pop()
	}
}

func (f *kdTranspFollower) Next() (coll goray.Collision) {
	for f.findMore(); len(f.currPrims) > 0; f.findMore() {
		p := f.currPrims[len(f.currPrims)-1]
		f.currPrims = f.currPrims[:len(f.currPrims)-1]
		if coll = p.Intersect(f.Ray); coll.Hit() && coll.RayDepth > f.MinDist && coll.RayDepth < f.MaxDist {
			return
		}
	}
	return
}

func (kd *kdPartition) Intersect(r goray.Ray, dist float64) (coll goray.Collision) {
	f := kdFollower{Ray: r, MinDist: r.TMin, MaxDist: dist}
	f.Init(kd)
	return f.First()
}

func (kd *kdPartition) Shadowed(r goray.Ray, dist float64) bool {
	f := kdFollower{Ray: r, MinDist: r.TMin, MaxDist: dist}
	f.Init(kd)
	return f.Hit()
}

func (kd *kdPartition) TransparentShadow(state *goray.RenderState, r goray.Ray, maxDepth int, dist float64) (filt color.Color, hit bool) {
	f := kdTranspFollower{kdFollower: kdFollower{Ray: r, MinDist: r.TMin, MaxDist: dist}}
	f.Init(kd)
	depth := 0
	filt = color.White
	for coll := f.Next(); coll.Hit(); coll = f.Next() {
		if depth >= maxDepth {
			// Too much depth, just say it's opaque.
			return color.Black, true
		}
		tmat, ok := coll.Primitive.Material().(goray.TransparentMaterial)
		if !ok {
			// Material does not have transparency.
			return color.Black, true
		}
		filt = color.Mul(filt, tmat.Transparency(state, coll.Surface(), f.Ray.Dir))
		if color.IsBlack(filt) {
			// Material is opaque.
			return color.Black, true
		}
		depth++
	}
	return
}
