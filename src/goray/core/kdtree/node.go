//
//	goray/core/kdtree/node.go
//	goray
//
//	Created by Ross Light on 2010-06-21.
//

package kdtree

import (
	"goray/core/vector"
)

// Value is a type for the individual elements stored in the leaves of the tree.
type Value interface{}

// Node represents nodes in a kd-tree, both interior and leaf.
type Node struct {
	axis   int8
	pivot  float64
	values []Value
}

func newLeaf(vals []Value) *Node {
	return &Node{
		axis:   -1,
		pivot:  0.0,
		values: vals,
	}
}

func newInterior(axis vector.Axis, pivot float64, left, right *Node) *Node {
	return &Node{
		axis:   int8(axis),
		pivot:  pivot,
		values: []Value{left, right},
	}
}

func (n *Node) Leaf() bool        { return n.axis == -1 }
func (n *Node) Axis() vector.Axis { return vector.Axis(n.axis) }
func (n *Node) Pivot() float64    { return n.pivot }
func (n *Node) Left() *Node       { return n.values[0].(*Node) }
func (n *Node) Right() *Node      { return n.values[1].(*Node) }
func (n *Node) Values() []Value   { return n.values }
