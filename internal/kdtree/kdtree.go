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

// Package kdtree provides a generic kd-tree implementation.
package kdtree

import (
	"fmt"

	"bitbucket.org/zombiezen/math3/vec64"
	"zombiezen.com/go/goray/internal/bound"
	"zombiezen.com/go/goray/internal/vecutil"
)

// A type that implements kdtree.Interface can be partitioned.
type Interface interface {
	Len() int
	Dimension(i int, axis vecutil.Axis) (min, max float64)
}

type Clipper interface {
	Clip(i int, bound bound.Bound, axis vecutil.Axis, lower bool, oldData interface{}) (clipped bound.Bound, newData interface{})
}

func getBound(data Interface, i int) bound.Bound {
	minX, maxX := data.Dimension(i, vecutil.X)
	minY, maxY := data.Dimension(i, vecutil.Y)
	minZ, maxZ := data.Dimension(i, vecutil.Z)
	return bound.Bound{vec64.Vector{minX, minY, minZ}, vec64.Vector{maxX, maxY, maxZ}}
}

// Tree is a generic kd-tree.
type Tree struct {
	root  *Node
	bound bound.Bound
}

// Root returns the root of the kd-tree.
func (tree *Tree) Root() *Node {
	return tree.root
}

// Bound returns a bounding box that encloses all objects in the tree.
func (tree *Tree) Bound() bound.Bound {
	return tree.bound
}

// Options allows you to tune the parameters of kd-tree construction.
type Options struct {
	MaxDepth       int // MaxDepth limits how many levels the tree can have
	LeafSize       int // LeafSize is the desired leaf size.  Some leaves may not obey this size.
	FaultTolerance int // FaultTolerance specifies the number of bad splits before a branch is considered a fault.
	ClipThreshold  int // ClipThreshold specifies the maximum number of values in a node to do primitive clipping.
}

// Reasonable build options
var DefaultOptions = Options{
	MaxDepth:       64,
	LeafSize:       2,
	FaultTolerance: 2,
	ClipThreshold:  32,
}

// buildState holds information for building a level of a kd-tree.
type buildState struct {
	Data    Interface
	Options Options

	TreeBound  bound.Bound
	OldCost    float64
	BadRefines int
	Clips      map[int]clipInfo
	ClipAxis   vecutil.Axis
	ClipLower  bool
}

type clipInfo struct {
	Bound        bound.Bound
	InternalData interface{}
}

func (state *buildState) ClippedDimension(i int, axis vecutil.Axis) (min, max float64) {
	if state.Clips != nil {
		if info, ok := state.Clips[i]; ok {
			return info.Bound.Min[axis], info.Bound.Max[axis]
		}
	}
	return state.Data.Dimension(i, axis)
}

// Split returns a copy of state with a different split.
func (state *buildState) Split(axis vecutil.Axis, lower bool, clips map[int]clipInfo) *buildState {
	return &buildState{
		Data:       state.Data,
		Options:    state.Options,
		TreeBound:  state.TreeBound,
		OldCost:    state.OldCost,
		BadRefines: state.BadRefines,
		Clips:      clips,
		ClipAxis:   axis,
		ClipLower:  lower,
	}
}

// New creates a new kd-tree by partitioning data.
func New(data Interface, opts Options) *Tree {
	tree := new(Tree)
	n := data.Len()
	if n == 0 {
		tree.root = newLeaf(nil)
		return tree
	}
	state := &buildState{
		Data:       data,
		Options:    opts,
		OldCost:    float64(n),
		BadRefines: 0,
		ClipAxis:   -1,
	}
	tree.bound = getBound(data, 0)
	for i := 1; i < n; i++ {
		tree.bound = bound.Union(tree.bound, getBound(data, i))
	}
	state.TreeBound = tree.bound
	indices := make([]int, n)
	for i := 0; i < n; i++ {
		indices[i] = i
	}
	tree.root = build(indices, tree.bound, state)
	return tree
}

// Depth returns the number of levels in the tree (excluding leaves).
func (tree *Tree) Depth() int {
	return nodeDepth(tree.root)
}

func nodeDepth(n *Node) int {
	if !n.IsLeaf() {
		leftDepth, rightDepth := nodeDepth(n.Left()), nodeDepth(n.Right())
		if leftDepth >= rightDepth {
			return leftDepth + 1
		} else {
			return rightDepth + 1
		}
	}
	return 0
}

func (tree *Tree) String() string {
	return nodeString(tree.root, 0)
}

func nodeString(n *Node, indent int) string {
	tab := "  "
	indentString := ""
	for i := 0; i < indent; i++ {
		indentString += tab
	}
	if n.IsLeaf() {
		return fmt.Sprint(n.Indices())
	}
	return fmt.Sprintf("{%c at %.2f\n%sL: %v\n%sR: %v\n%s}",
		"XYZ"[n.Axis()], n.Pivot(),
		indentString+tab, nodeString(n.Left(), indent+1),
		indentString+tab, nodeString(n.Right(), indent+1),
		indentString)
}

func build(indices []int, bd bound.Bound, state *buildState) *Node {
	n := len(indices)

	// Clip any primitives
	if n <= state.Options.ClipThreshold {
		indices, state.Clips, _ = clip(indices, bd, state)
	}

	// If we're within acceptable bounds (or we're just sick of building the tree),
	// then make a leaf.
	if n <= state.Options.LeafSize || state.Options.MaxDepth <= 0 {
		return newLeaf(indices)
	}

	// Pick a pivot
	axis, pivot, cost := split(indices, bd, state)

	// Is this bad?
	if cost > state.OldCost {
		state.BadRefines++
	}
	if (cost > state.OldCost*1.6 && n < 16) || state.BadRefines >= state.Options.FaultTolerance {
		// We've done some *bad* splitting.  Just leaf it.
		return newLeaf(indices)
	}

	// Sort out values
	left, right := make([]int, 0, n), make([]int, 0, n)
	for _, i := range indices {
		min, max := state.Data.Dimension(i, axis)
		if min < pivot {
			left = append(left, i)
		}
		if min >= pivot || max > pivot {
			right = append(right, i)
		}
	}
	var leftClip, rightClip map[int]clipInfo
	if state.Clips != nil {
		leftClip = make(map[int]clipInfo, len(left))
		for _, i := range left {
			leftClip[i] = state.Clips[i]
		}
		rightClip = make(map[int]clipInfo, len(right))
		for _, i := range right {
			rightClip[i] = state.Clips[i]
		}
	}

	// Calculate new bounds
	leftBound, rightBound := bd, bd
	leftBound.Max[axis] = pivot
	rightBound.Min[axis] = pivot

	// Build subtrees
	state.OldCost = cost
	leftChan, rightChan := make(chan *Node, 1), make(chan *Node, 1)
	state.Options.MaxDepth--
	go func() {
		leftChan <- build(left, leftBound, state.Split(axis, false, leftClip))
	}()
	go func() {
		rightChan <- build(right, rightBound, state.Split(axis, true, rightClip))
	}()

	// Return interior node
	return newInterior(axis, pivot, <-leftChan, <-rightChan)
}

func clip(indices []int, nodeBound bound.Bound, state *buildState) ([]int, map[int]clipInfo, bound.Bound) {
	const (
		treeSizeWeight = 1e-5
		nodeSizeWeight = 0.021
	)

	n := len(indices)
	clipper, ok := state.Data.(Clipper)
	if !ok || n == 0 {
		return indices, nil, nodeBound
	}

	var bd bound.Bound
	for axis := range bd.Min {
		treeSize := state.TreeBound.Max[axis] - state.TreeBound.Min[axis]
		nodeSize := nodeBound.Max[axis] - nodeBound.Min[axis]
		delta := treeSize*treeSizeWeight + nodeSize*nodeSizeWeight

		bd.Min[axis] = nodeBound.Min[axis] - delta
		bd.Max[axis] = nodeBound.Max[axis] + delta
	}

	clipIndices := make([]int, 0, n)
	clips := make(map[int]clipInfo, n)
	clipBox := bound.Bound{}
	for _, i := range indices {
		info := state.Clips[i]
		if info.Bound.IsZero() {
			info.Bound = getBound(state.Data, i)
		}
		newBound, newData := clipper.Clip(i, bd, state.ClipAxis, state.ClipLower, info.InternalData)
		if !newBound.IsZero() {
			// Polygon clipped and still with us.
			info.Bound, info.InternalData = newBound, newData
			clipIndices = append(clipIndices, i)
			clips[i] = info

			// Update bound
			if clipBox.IsZero() {
				clipBox = newBound
			} else {
				clipBox = bound.Union(clipBox, newBound)
			}
		}
	}

	return clipIndices, clips, clipBox
}
