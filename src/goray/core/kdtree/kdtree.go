//
//	goray/core/kdtree/kdtree.go
//	goray
//
//	Created by Ross Light on 2010-06-02.
//

// The kdtree package provides a generic kd-tree implementation.
package kdtree

import (
	"fmt"
	"goray/logging"
	"goray/core/bound"
	"goray/core/primitive"
	"goray/core/vector"
)

// Tree is a generic kd-tree.
type Tree struct {
	root  Node
	bound *bound.Bound
}

// A DimensionFunc calculates the range of a value in a particular axis.
type DimensionFunc func(v Value, axis int) (min, max float)

func getBound(v Value, getDim DimensionFunc) *bound.Bound {
	minX, maxX := getDim(v, 0)
	minY, maxY := getDim(v, 1)
	minZ, maxZ := getDim(v, 2)
	return bound.New(vector.New(minX, minY, minZ), vector.New(maxX, maxY, maxZ))
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
	Bound        *bound.Bound
	InternalData interface{}
}

// BuildState holds information for building a level of a kd-tree.
type BuildState struct {
	Options
	TreeBound  *bound.Bound
	OldCost    float
	BadRefines uint
	Clips      []ClipInfo
	ClipAxis   int
}

func (state BuildState) getBound(v Value) *bound.Bound {
	return getBound(v, state.GetDimension)
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
		OldCost:    float(len(vals)),
		BadRefines: 0,
		ClipAxis:   -1,
	}

	if len(vals) > 0 {
		tree.bound = bound.New(state.getBound(vals[0]).Get())
		for _, v := range vals[1:] {
			tree.bound = bound.Union(tree.bound, state.getBound(v))
		}
	} else {
		tree.bound = bound.New(vector.New(0, 0, 0), vector.New(0, 0, 0))
	}
	state.TreeBound = tree.bound

	tree.root = build(vals, tree.bound, state)
	logging.Debug(opts.Log, "kd-tree is %d levels deep", tree.Depth())
	return tree
}

// Depth returns the number of levels in the tree (excluding leaves).
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

func build(vals []Value, bd *bound.Bound, state BuildState) Node {
	// Clip any primitives
	if uint(len(vals)) <= state.ClipThreshold {
		vals, state.Clips, bd = clip(vals, bd, state)
	}
	// If we're within acceptable bounds (or we're just sick of building the tree),
	// then make a leaf.
	if len(vals) <= state.LeafSize || state.MaxDepth <= 0 {
		return newLeaf(vals)
	}
	// Pick a pivot
	// TODO: Make split functions aware of clipping
	axis, pivot, cost := state.SplitFunc(vals, bd, state)
	// Is this bad?
	if cost > state.OldCost {
		state.BadRefines++
	}
	if (cost > state.OldCost*1.6 && len(vals) < 16) || state.BadRefines >= state.FaultTolerance {
		// We've done some *bad* splitting.  Just leaf it.
		logging.Debug(state.Log, "Faulted %d values", len(vals))
		return newLeaf(vals)
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
	state.OldCost = cost
	leftChan, rightChan := make(chan Node, 1), make(chan Node, 1)
	state.MaxDepth--
	go func() {
		leftState := state
		leftState.ClipAxis = axis
		leftState.Clips = leftClip
		leftChan <- build(left, leftBound, leftState)
	}()
	go func() {
		rightState := state
		rightState.ClipAxis = axis | 1<<2
		rightState.Clips = rightClip
		rightChan <- build(right, rightBound, rightState)
	}()
	// Return interior node
	return newInterior(axis, pivot, <-leftChan, <-rightChan)
}

func clip(vals []Value, nodeBound *bound.Bound, state BuildState) (clipVals []Value, clipData []ClipInfo, clipBox *bound.Bound) {
	const treeSizeWeight = 1e-5
	const nodeSizeWeight = 0.021

	if len(vals) == 0 {
		return vals, []ClipInfo{}, nodeBound
	}

	var bExt [2][3]float
	for axis := 0; axis < 3; axis++ {
		treeSize := state.TreeBound.GetMax().GetComponent(axis) - state.TreeBound.GetMin().GetComponent(axis)
		nodeSize := nodeBound.GetMax().GetComponent(axis) - nodeBound.GetMin().GetComponent(axis)
		delta := treeSize*treeSizeWeight + nodeSize*nodeSizeWeight

		bExt[0][axis] = nodeBound.GetMin().GetComponent(axis) - delta
		bExt[1][axis] = nodeBound.GetMax().GetComponent(axis) + delta
	}

	bd := bound.New(vector.New(bExt[0][0], bExt[0][1], bExt[0][2]), vector.New(bExt[1][0], bExt[1][1], bExt[1][2]))

	clipVals = make([]Value, 0, len(vals))
	clipData = make([]ClipInfo, 0, len(vals))
	nClip := 0
	for i, v := range vals {
		if clipv, ok := v.(primitive.Clipper); ok {
			info := state.getClipInfo(i)
			if info.Bound == nil {
				info.Bound = state.getBound(v)
			}
			newBound, newData := clipv.ClipToBound(bd, state.ClipAxis, info.InternalData)
			if newBound != nil {
				// Polygon clipped and still with us.
				info.Bound, info.InternalData = newBound, newData
				clipVals = append(clipVals, v)
				clipData = append(clipData, info)

				// Update bound
				if clipBox == nil {
					clipBox = newBound
				} else {
					clipBox = bound.Union(clipBox, newBound)
				}
			} else {
				// TODO: Handle "error in clipping" case differently?
				nClip++
			}
		} else {
			if clipBox == nil {
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

// GetRoot returns the root of the kd-tree.
func (tree *Tree) GetRoot() Node { return tree.root }

// GetBound returns a bounding box that encloses all objects in the tree.
func (tree *Tree) GetBound() *bound.Bound { return bound.New(tree.bound.Get()) }
