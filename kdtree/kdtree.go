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
	"bitbucket.org/zombiezen/goray/bound"
	"bitbucket.org/zombiezen/goray/logging"
	"bitbucket.org/zombiezen/goray/vector"
	"fmt"
)

// Tree is a generic kd-tree.
type Tree struct {
	root  *Node
	bound bound.Bound
}

// A DimensionFunc calculates the range of a value in a particular axis.
type DimensionFunc func(v Value, axis vector.Axis) (min, max float64)

func getBound(v Value, getDim DimensionFunc) bound.Bound {
	minX, maxX := getDim(v, 0)
	minY, maxY := getDim(v, 1)
	minZ, maxZ := getDim(v, 2)
	return bound.Bound{vector.Vector3D{minX, minY, minZ}, vector.Vector3D{maxX, maxY, maxZ}}
}

const (
	DefaultMaxDepth       = 64
	DefaultLeafSize       = 2
	DefaultFaultTolerance = 2
	DefaultClipThreshold  = 32
)

// Options allows you to tune the parameters of kd-tree construction.
type Options struct {
	GetDimension   DimensionFunc
	SplitFunc      SplitFunc
	Log            logging.Handler
	MaxDepth       int  // MaxDepth limits how many levels the tree can have
	LeafSize       int  // LeafSize is the desired leaf size.  Some leaves may not obey this size.
	FaultTolerance uint // FaultTolerance specifies the number of bad splits before a branch is considered a fault.
	ClipThreshold  uint // ClipThreshold specifies the maximum number of values in a node to do primitive clipping.
}

// MakeOptions creates a new set of build options with some reasonable defaults.
func MakeOptions(f DimensionFunc, log logging.Handler) Options {
	return Options{
		GetDimension:   f,
		SplitFunc:      DefaultSplit,
		Log:            log,
		MaxDepth:       DefaultMaxDepth,
		LeafSize:       DefaultLeafSize,
		FaultTolerance: DefaultFaultTolerance,
		ClipThreshold:  DefaultClipThreshold,
	}
}

type ClipInfo struct {
	Bound        bound.Bound
	InternalData interface{}
}

// BuildState holds information for building a level of a kd-tree.
type BuildState struct {
	Options
	TreeBound  bound.Bound
	OldCost    float64
	BadRefines uint
	Clips      []ClipInfo
	ClipAxis   vector.Axis
	ClipLower  bool
	Pool       *nodePool
}

func (state BuildState) getBound(v Value) bound.Bound {
	return getBound(v, state.GetDimension)
}

func (state BuildState) getClippedDimension(i int, v Value, axis vector.Axis) (min, max float64) {
	if info := state.getClipInfo(i); !info.Bound.IsZero() {
		return info.Bound.Min[axis], info.Bound.Max[axis]
	}
	return state.GetDimension(v, axis)
}

func (state BuildState) getClipInfo(idx int) ClipInfo {
	if state.Clips == nil {
		return ClipInfo{}
	}
	return state.Clips[idx]
}

// New creates a new kd-tree from an unordered collection of values.
func New(vals []Value, opts Options) (tree *Tree) {
	tree = new(Tree)
	state := BuildState{
		Options:    opts,
		OldCost:    float64(len(vals)),
		BadRefines: 0,
		ClipAxis:   -1,
		Pool:       newNodePool(len(vals)/2, len(vals)),
	}

	if len(vals) > 0 {
		tree.bound = state.getBound(vals[0])
		for _, v := range vals[1:] {
			tree.bound = bound.Union(tree.bound, state.getBound(v))
		}
	}
	state.TreeBound = tree.bound

	tree.root = build(vals, tree.bound, state)
	logging.Debug(opts.Log, "kd-tree is %d levels deep", tree.Depth())
	return tree
}

// Depth returns the number of levels in the tree (excluding leaves).
func (tree *Tree) Depth() int {
	return nodeDepth(tree.root)
}

func nodeDepth(n *Node) int {
	if !n.Leaf() {
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
	if n.Leaf() {
		return fmt.Sprint(n.Values())
	}
	return fmt.Sprintf("{%c at %.2f\n%sL: %v\n%sR: %v\n%s}",
		"XYZ"[n.Axis()], n.Pivot(),
		indentString+tab, nodeString(n.Left(), indent+1),
		indentString+tab, nodeString(n.Right(), indent+1),
		indentString)
}

func build(vals []Value, bd bound.Bound, state BuildState) *Node {
	// Clip any primitives
	if uint(len(vals)) <= state.ClipThreshold {
		vals, state.Clips, _ = clip(vals, bd, state)
	}

	// If we're within acceptable bounds (or we're just sick of building the tree),
	// then make a leaf.
	if len(vals) <= state.LeafSize || state.MaxDepth <= 0 {
		return state.Pool.NewLeaf(vals)
	}

	// Pick a pivot
	axis, pivot, cost := state.SplitFunc(vals, bd, state)

	// Is this bad?
	if cost > state.OldCost {
		state.BadRefines++
	}
	if (cost > state.OldCost*1.6 && len(vals) < 16) || state.BadRefines >= state.FaultTolerance {
		// We've done some *bad* splitting.  Just leaf it.
		logging.Debug(state.Log, "Faulted %d values", len(vals))
		return state.Pool.NewLeaf(vals)
	}

	// Sort out values
	left, right := make([]Value, 0, len(vals)), make([]Value, 0, len(vals))
	leftClip, rightClip := make([]ClipInfo, 0, len(vals)), make([]ClipInfo, 0, len(vals))
	for i, v := range vals {
		vMin, vMax := state.GetDimension(v, axis)
		if vMin < pivot {
			left = append(left, v)
			leftClip = append(leftClip, state.getClipInfo(i))
		}
		if vMin >= pivot || vMax > pivot {
			right = append(right, v)
			rightClip = append(rightClip, state.getClipInfo(i))
		}
	}

	// Calculate new bounds
	leftBound, rightBound := bd, bd
	leftBound.Max[axis] = pivot
	rightBound.Min[axis] = pivot

	// Build subtrees
	state.OldCost = cost
	leftChan, rightChan := make(chan *Node, 1), make(chan *Node, 1)
	state.MaxDepth--
	go func() {
		leftState := state
		leftState.ClipAxis, leftState.ClipLower = axis, false
		leftState.Clips = leftClip
		leftChan <- build(left, leftBound, leftState)
	}()
	go func() {
		rightState := state
		rightState.ClipAxis, rightState.ClipLower = axis, true
		rightState.Clips = rightClip
		rightChan <- build(right, rightBound, rightState)
	}()

	// Return interior node
	return state.Pool.NewInterior(axis, pivot, <-leftChan, <-rightChan)
}

func clip(vals []Value, nodeBound bound.Bound, state BuildState) (clipVals []Value, clipData []ClipInfo, clipBox bound.Bound) {
	const treeSizeWeight = 1e-5
	const nodeSizeWeight = 0.021

	if len(vals) == 0 {
		return vals, []ClipInfo{}, nodeBound
	}

	var bd bound.Bound
	for axis := vector.X; axis <= vector.Z; axis++ {
		treeSize := state.TreeBound.Max[axis] - state.TreeBound.Min[axis]
		nodeSize := nodeBound.Max[axis] - nodeBound.Min[axis]
		delta := treeSize*treeSizeWeight + nodeSize*nodeSizeWeight

		bd.Min[axis] = nodeBound.Min[axis] - delta
		bd.Max[axis] = nodeBound.Max[axis] + delta
	}

	clipVals = make([]Value, 0, len(vals))
	clipData = make([]ClipInfo, 0, len(vals))
	nClip := 0
	for i, v := range vals {
		if clipv, ok := v.(Clipper); ok {
			info := state.getClipInfo(i)
			if info.Bound.IsZero() {
				info.Bound = state.getBound(v)
			}
			newBound, newData := clipv.Clip(bd, state.ClipAxis, state.ClipLower, info.InternalData)
			if !newBound.IsZero() {
				// Polygon clipped and still with us.
				info.Bound, info.InternalData = newBound, newData
				clipVals = append(clipVals, v)
				clipData = append(clipData, info)

				// Update bound
				if clipBox.IsZero() {
					clipBox = newBound
				} else {
					clipBox = bound.Union(clipBox, newBound)
				}
			} else {
				// TODO: Handle "error in clipping" case differently?
				nClip++
			}
		} else {
			if clipBox.IsZero() {
				clipBox = state.getBound(v)
			} else {
				clipBox = bound.Union(clipBox, state.getBound(v))
			}
			clipVals = append(clipVals, v)
			clipData = append(clipData, ClipInfo{})
		}
	}

	if nClip > 0 {
		logging.VerboseDebug(state.Log, "Clipped %d values", nClip)
	}

	return
}

// Root returns the root of the kd-tree.
func (tree *Tree) Root() *Node { return tree.root }

// Bound returns a bounding box that encloses all objects in the tree.
func (tree *Tree) Bound() bound.Bound { return tree.bound }
