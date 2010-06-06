//
//  goray/kdtree.go
//  goray
//
//  Created by Ross Light on 2010-06-02.
//

package kdtree

type Tree struct {
	root Node
}

type CompareFunc func(Value, float) bool
type AxisFunc func(Value, axis int) float

type buildParams struct {
	GetAxis           AxisFunc
	LeftCmp, RightCmp CompareFunc
	MaxDepth          int
	LeafSize          int
}

func New(vals []Value, getAxis AxisFunc, left, right CompareFunc) (tree *Tree) {
	params := buildParams{getAxis, left, right, 64, 2}
	tree.root = build(vals, params)
	return tree
}

func build(vals []Value, params buildParams) Node {
	// If we're within acceptable bounds (or we're just sick of building the tree),
	// then make a leaf.
	if len(vals) <= params.LeafSize || params.MaxDepth <= 0 {
		return newLeaf(vals)
	}
	// TODO: Pick a pivot
	axis := 0
	pivot := 0.0
	// Sort out values
	left, right := make([]Value, 0, len(vals)), make([]Value, 0, len(vals))
	for _, v := range vals {
		if params.LeftCmp(v, pivot) {
			left = left[0 : len(left)+1]
			left[len(left)-1] = v
		}
		if params.RightCmp(v, pivot) {
			right = right[0 : len(right)+1]
			right[len(right)-1] = v
		}
	}
	// Build subtrees
	leftChan, rightChan := make(chan Node), make(chan Node)
	params.MaxDepth--
	go func() {
		leftChan <- build(left, params)
	}()
	go func() {
		rightChan <- build(right, params)
	}()
	// Return interior node
	return newInterior(axis, pivot, <-leftChan, <-rightChan)
}

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
