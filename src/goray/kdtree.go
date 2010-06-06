//
//  goray/kdtree.go
//  goray
//
//  Created by Ross Light on 2010-06-02.
//

package kdtree

import (
	"sort"
	"./goray/bound"
	"./goray/vector"
)

type Tree struct {
	root  Node
	bound *bound.Bound
}

type CompareFunc func(v Value, axis int, pivot float) bool
type AxisFunc func(v Value, axis int) float

type buildParams struct {
	GetAxis           AxisFunc
	LeftCmp, RightCmp CompareFunc
	MaxDepth          int
	LeafSize          int
}

func getPoint(v Value, getAxis AxisFunc) vector.Vector3D {
	return vector.New(getAxis(v, 0), getAxis(v, 1), getAxis(v, 2))
}

func New(vals []Value, getAxis AxisFunc, left, right CompareFunc) (tree *Tree) {
	params := buildParams{getAxis, left, right, 64, 2}
	if len(vals) > 0 {
		pt := getPoint(vals[0], getAxis)
		tree.bound = bound.New(pt, pt)
		for _, v := range vals[1:] {
			pt = getPoint(v, getAxis)
			tree.bound.Include(pt)
		}
	} else {
		tree.bound = bound.New(vector.New(0, 0, 0), vector.New(0, 0, 0))
	}
	tree.root = build(vals, tree.bound, params)
	return tree
}

func simpleSplit(vals []Value, bd *bound.Bound, params buildParams) (axis int, pivot float) {
	axis = bd.GetLargestAxis()
	data := make([]float, len(vals))
	for i, v := range vals {
		data[i] = params.GetAxis(v, axis)
	}
	sort.SortFloats(data)
	pivot = data[len(data)/2]
	return
}

func build(vals []Value, bd *bound.Bound, params buildParams) Node {
	// If we're within acceptable bounds (or we're just sick of building the tree),
	// then make a leaf.
	if len(vals) <= params.LeafSize || params.MaxDepth <= 0 {
		return newLeaf(vals)
	}
	// Pick a pivot
	axis, pivot := simpleSplit(vals, bd, params)
	// Sort out values
	left, right := make([]Value, 0, len(vals)), make([]Value, 0, len(vals))
	for _, v := range vals {
		if params.LeftCmp(v, axis, pivot) {
			left = left[0 : len(left)+1]
			left[len(left)-1] = v
		}
		if params.RightCmp(v, axis, pivot) {
			right = right[0 : len(right)+1]
			right[len(right)-1] = v
		}
	}
	// Calculate new bounds
	leftBound, rightBound := bound.New(bd.Get()), bound.New(bd.Get())
	switch axis {
	case 0:
		leftBound.SetMaxX(pivot)
		rightBound.SetMinX(pivot)
	case 1:
		leftBound.SetMaxY(pivot)
		rightBound.SetMinY(pivot)
	case 2:
		leftBound.SetMaxZ(pivot)
		rightBound.SetMinZ(pivot)
	}
	// Build subtrees
	leftChan, rightChan := make(chan Node), make(chan Node)
	params.MaxDepth--
	go func() {
		leftChan <- build(left, leftBound, params)
	}()
	go func() {
		rightChan <- build(right, rightBound, params)
	}()
	// Return interior node
	return newInterior(axis, pivot, <-leftChan, <-rightChan)
}

func (tree *Tree) GetRoot() Node          { return tree.root }
func (tree *Tree) GetBound() *bound.Bound { return bound.New(tree.bound.Get()) }

type Value interface{}
type Node interface {
	IsLeaf() bool
}

type Leaf struct {
	values []Value
}

func newLeaf(vals []Value) *Leaf      { return &Leaf{vals} }
func (leaf *Leaf) IsLeaf() bool       { return true }
func (leaf *Leaf) GetValues() []Value { return leaf.values }

type Interior struct {
	axis        int8
	pivot       float
	left, right Node
}

func newInterior(axis int, pivot float, left, right Node) *Interior {
	return &Interior{int8(axis), pivot, left, right}
}

func (i *Interior) IsLeaf() bool    { return false }
func (i *Interior) GetAxis() int    { return int(i.axis) }
func (i *Interior) GetPivot() float { return i.pivot }
func (i *Interior) GetLeft() Node   { return i.left }
func (i *Interior) GetRight() Node  { return i.right }
