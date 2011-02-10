//
//	goray/core/kdtree/node.go
//	goray
//
//	Created by Ross Light on 2010-06-21.
//

package kdtree

import (
	"sync"
	"goray/core/vector"
)

type nodePool struct {
	nodes    []Node
	nodeIdx  int
	values   []Value
	valueIdx int
	lock     sync.Mutex
}

func newNodePool(nNodes, nValues int) (pool *nodePool) {
	if nNodes < 1 {
		nNodes = 1
	}
	return &nodePool{
		nodes:  make([]Node, nNodes),
		values: make([]Value, nValues),
	}
}

func (pool *nodePool) nextNode() (n *Node) {
	if pool.nodeIdx == len(pool.nodes) {
		pool.nodes = make([]Node, len(pool.nodes))
		pool.nodeIdx = 0
	}
	n = &pool.nodes[pool.nodeIdx]
	pool.nodeIdx++
	return
}

func (pool *nodePool) nextValues(size int) (vals []Value) {
	if len(pool.values) < pool.valueIdx+size {
		if len(pool.values) < size {
			pool.values = make([]Value, size)
		} else {
			pool.values = make([]Value, len(pool.values))
		}
		pool.valueIdx = 0
	}
	vals = pool.values[pool.valueIdx : pool.valueIdx+size]
	pool.valueIdx += size
	return
}

func (pool *nodePool) NewLeaf(vals []Value) (n *Node) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	n = pool.nextNode()
	n.axis = -1
	n.pivot = 0.0
	n.values = pool.nextValues(len(vals))
	copy(n.values, vals)
	return
}

func (pool *nodePool) NewInterior(axis vector.Axis, pivot float64, left, right *Node) (n *Node) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	n = pool.nextNode()
	n.axis = int8(axis)
	n.pivot = pivot
	n.values = pool.nextValues(2)
	n.values[0] = left
	n.values[1] = right
	return
}

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
