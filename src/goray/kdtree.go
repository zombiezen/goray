//
//  goray/kdtree.go
//  goray
//
//  Created by Ross Light on 2010-06-02.
//

package kdtree

import (
	"fmt"
	"sort"
	"./fmath"
)

import (
	"./goray/bound"
	"./goray/vector"
)

type Tree struct {
	root  Node
	bound *bound.Bound
}

type DimensionFunc func(v Value, axis int) (min, max float)

type buildParams struct {
	GetDimension DimensionFunc
	MaxDepth     int
	LeafSize     int
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

func New(vals []Value, getDim DimensionFunc) (tree *Tree) {
	tree = new(Tree)
	params := buildParams{getDim, 16, 2}
	if len(vals) > 0 {
		tree.bound = bound.New(getBound(vals[0], getDim).Get())
		for _, v := range vals[1:] {
			tree.bound = bound.Union(tree.bound, params.getBound(v))
		}
	} else {
		tree.bound = bound.New(vector.New(0, 0, 0), vector.New(0, 0, 0))
	}
	tree.root = build(vals, tree.bound, params)
	fmt.Printf("Tree is %d levels deep\n", tree.depth())
	return tree
}

func (tree *Tree) depth() int {
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
	axis, pivot := pigeonSplit(vals, bd, params)
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

func simpleSplit(vals []Value, bd *bound.Bound, params buildParams) (axis int, pivot float) {
	axis = bd.GetLargestAxis()
	data := make([]float, len(vals))
	for i, v := range vals {
		min, max := params.GetDimension(v, axis)
		data[i] = (min + max) / 2
	}
	sort.SortFloats(data)
	pivot = data[len(data)/2]
	return
}

type pigeonBin struct {
	n           int
	left, right int
	bleft, both int
	t           float
}

func (b pigeonBin) empty() bool { return b.n == 0 }

func pigeonSplit(vals []Value, bd *bound.Bound, params buildParams) (bestAxis int, bestPivot float) {
	const numBins = 1024
	const emptyBonus = 0.33
	const costRatio = 0.35

	var bins [numBins + 1]pigeonBin
	d := [3]float{bd.GetXLength(), bd.GetYLength(), bd.GetZLength()}
	bestCost := fmath.Inf
	totalSA := d[0]*d[1] + d[0]*d[2] + d[1]*d[2]
	invTotalSA := 0.0
	if !fmath.Eq(totalSA, 0.0) {
		invTotalSA = 1.0 / totalSA
	}

	for axis := 0; axis < 3; axis++ {
		s := numBins / d[axis]
		min := bd.GetMin().GetComponent(axis)

		for _, v := range vals {
			tLow, tHigh := params.GetDimension(v, axis)
			bLeft, bRight := int((tLow-min)*s), int((tHigh-min)*s)
			if bLeft < 0 {
				bLeft = 0
			} else if bLeft > numBins {
				bLeft = numBins
			}
			if bRight < 0 {
				bRight = 0
			} else if bRight > numBins {
				bRight = numBins
			}

			if tLow == tHigh {
				if bins[bLeft].empty() || tLow >= bins[bLeft].t {
					bins[bLeft].t = tLow
					bins[bLeft].both++
				} else {
					bins[bLeft].left++
					bins[bLeft].right++
				}
				bins[bLeft].n += 2
			} else {
				if bins[bLeft].empty() || tLow > bins[bLeft].t {
					bins[bLeft].t = tLow
					bins[bLeft].left += bins[bLeft].both + bins[bLeft].bleft
					bins[bLeft].right += bins[bLeft].both
					bins[bLeft].both, bins[bLeft].bleft = 0, 0
					bins[bLeft].bleft++
				} else if tLow == bins[bLeft].t {
					bins[bLeft].bleft++
				} else {
					bins[bLeft].left++
				}

				bins[bLeft].n++
				bins[bRight].right++
				if bins[bRight].empty() || tHigh > bins[bRight].t {
					bins[bRight].t = tHigh
					bins[bRight].left += bins[bRight].both + bins[bRight].bleft
					bins[bRight].right += bins[bRight].both
					bins[bRight].both, bins[bRight].bleft = 0, 0
				}
				bins[bRight].n++
			}
		}

		capArea := d[(axis+1)%3] * d[(axis+2)%3]
		capPerim := d[(axis+1)%3] + d[(axis+2)%3]

		nBelow, nAbove := 0, len(vals)
		// Cumulate values and evaluate cost
		for _, b := range bins {
			if !b.empty() {
				nBelow += b.left
				nAbove -= b.right
				// Cost:
				edget := b.t
				if edget > bd.GetMin().GetComponent(axis) && edget < bd.GetMax().GetComponent(axis) {
					l1, l2 := edget-bd.GetMin().GetComponent(axis), bd.GetMax().GetComponent(axis)-edget
					belowSA, aboveSA := capArea+l1*capPerim, capArea+l2*capPerim
					rawCosts := belowSA*float(nBelow) + aboveSA*float(nAbove)

					eb := 0.0
					if nAbove == 0 {
						eb = (0.1 + l2/d[axis]) * emptyBonus * rawCosts
					} else if nBelow == 0 {
						eb = (0.1 + l1/d[axis]) * emptyBonus * rawCosts
					}

					cost := costRatio + invTotalSA*(rawCosts-eb)
					if cost < bestCost {
						bestAxis, bestPivot = axis, edget
					}
				}

				nBelow += b.both + b.bleft
				nAbove -= b.both
			}
		}

		if nBelow != len(vals) || nAbove != 0 {
			// SCREWED.
			panic("Cost function mismatch")
		}

		// Reset all bins
		for i, _ := range bins {
			bins[i] = pigeonBin{}
		}
	}

	return
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
