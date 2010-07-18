//
//	goray/core/kdtree/node.go
//	goray
//
//	Created by Ross Light on 2010-06-21.
//

package kdtree

// Value is a type for the individual elements stored in the leaves of the tree.
type Value interface{}

// Node is the common interface for leaf and interior nodes.
type Node interface {
	IsLeaf() bool
}

// Leaf is the node type that actually stores values.
type Leaf struct {
	values []Value
}

func newLeaf(vals []Value) *Leaf      { return &Leaf{vals} }
func (leaf *Leaf) IsLeaf() bool       { return true }
func (leaf *Leaf) GetValues() []Value { return leaf.values }

// Interior is represents a planar split.
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
