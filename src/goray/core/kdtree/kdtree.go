//
//  goray/core/kdtree/kdtree.go
//  goray
//
//  Created by Ross Light on 2010-06-02.
//

/* The kdtree package provides a generic kd-tree implementation. */
package kdtree

import (
	"fmt"
	"goray/logging"
	"goray/core/bound"
	"goray/core/vector"
)

/* Tree is a generic kd-tree */
type Tree struct {
	root  Node
	bound *bound.Bound
}

/* A DimensionFunc calculates the range of a value in a particular axis. */
type DimensionFunc func(v Value, axis int) (min, max float)

type buildParams struct {
	GetDimension DimensionFunc
	MaxDepth     int
	LeafSize     int
	Log          logging.Handler
	OldCost      float
	BadRefines   uint
}

func getBound(v Value, getDim DimensionFunc) *bound.Bound {
	minX, maxX := getDim(v, 0)
	minY, maxY := getDim(v, 1)
	minZ, maxZ := getDim(v, 2)
	return bound.New(vector.New(minX, minY, minZ), vector.New(maxX, maxY, maxZ))
}

func (bp buildParams) getBound(v Value) *bound.Bound {
	return getBound(v, bp.GetDimension)
}

/* New creates a new kd-tree from an unordered collection of values. */
func New(vals []Value, getDim DimensionFunc, log logging.Handler) (tree *Tree) {
	tree = new(Tree)
	params := buildParams{getDim, 16, 2, log, float(len(vals)), 0} // TODO: Make this deeper later
	if len(vals) > 0 {
		tree.bound = bound.New(getBound(vals[0], getDim).Get())
		for _, v := range vals[1:] {
			tree.bound = bound.Union(tree.bound, params.getBound(v))
		}
	} else {
		tree.bound = bound.New(vector.New(0, 0, 0), vector.New(0, 0, 0))
	}
	tree.root = build(vals, tree.bound, params)
	logging.Debug(log, "kd-tree is %d levels deep", tree.Depth())
	return tree
}

/* Depth returns the number of levels in the tree (excluding leaves). */
func (tree *Tree) Depth() int {
	var nodeDepth func(Node) int
	nodeDepth = func(n Node) int {
		switch node := n.(type) {
		case *Leaf:
			return 0
		case *Interior:
			leftDepth, rightDepth := nodeDepth(node.left), nodeDepth(node.right)
			if leftDepth >= rightDepth {
				return leftDepth + 1
			} else {
				return rightDepth + 1
			}
		}
		return 0
	}
	return nodeDepth(tree.root)
}

func (tree *Tree) String() string {
	var nodeString func(Node, int) string
	nodeString = func(n Node, indent int) string {
		tab := "  "
		indentString := ""
		for i := 0; i < indent; i++ {
			indentString += tab
		}
		switch node := n.(type) {
		case *Leaf:
			return fmt.Sprint(node.values)
		case *Interior:
			return fmt.Sprintf("{%c at %.2f\n%sL: %v\n%sR: %v\n%s}",
				"XYZ"[node.axis], node.pivot,
				indentString+tab, nodeString(node.left, indent+1),
				indentString+tab, nodeString(node.right, indent+1),
				indentString)
		}
		return ""
	}
	return nodeString(tree.root, 0)
}

func build(vals []Value, bd *bound.Bound, params buildParams) Node {
	// If we're within acceptable bounds (or we're just sick of building the tree),
	// then make a leaf.
	if len(vals) <= params.LeafSize || params.MaxDepth <= 0 {
		return newLeaf(vals)
	}
	// Pick a pivot
	var f splitFunc
	switch {
	case len(vals) > 128:
		f = pigeonSplit
	default:
		f = minimalSplit
	}
	axis, pivot, cost := f(vals, bd, params)
	// Is this bad?
	if cost > params.OldCost {
		params.BadRefines++
	}
	if (cost > params.OldCost * 1.6 && len(vals) < 16) || params.BadRefines >= 2 {
		// We've done some *bad* splitting.  Just leaf it.
		logging.Warning(params.Log, "Bad split")
		return newLeaf(vals)
	}
	// Sort out values
	left, right := make([]Value, 0, len(vals)), make([]Value, 0, len(vals))
	for _, v := range vals {
		vMin, vMax := params.GetDimension(v, axis)
		if vMin < pivot {
			left = left[0 : len(left)+1]
			left[len(left)-1] = v
		}
		if vMin >= pivot || vMax > pivot {
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
	params.OldCost = cost
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

/* GetRoot returns the root of the kd-tree. */
func (tree *Tree) GetRoot() Node { return tree.root }

/* GetBound returns a bounding box that encloses all objects in the tree. */
func (tree *Tree) GetBound() *bound.Bound { return bound.New(tree.bound.Get()) }
