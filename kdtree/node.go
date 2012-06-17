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

package kdtree

import (
	"bitbucket.org/zombiezen/goray/vector"
)

// TODO: This could be more memory-efficient.

// Node represents nodes in a kd-tree, both interior and leaf.
type Node struct {
	axis    int8
	pivot   float64
	indices []int
	left    *Node
	right   *Node
}

func newLeaf(indices []int) *Node {
	return &Node{
		axis:    -1,
		indices: indices,
	}
}

func newInterior(axis vector.Axis, pivot float64, left, right *Node) *Node {
	return &Node{
		axis:  int8(axis),
		pivot: pivot,
		left:  left,
		right: right,
	}
}

func (n *Node) IsLeaf() bool {
	return n.axis == -1
}

func (n *Node) Axis() vector.Axis {
	return vector.Axis(n.axis)
}

func (n *Node) Pivot() float64 {
	return n.pivot
}

func (n *Node) Left() *Node {
	return n.left
}

func (n *Node) Right() *Node {
	return n.right
}

func (n *Node) Indices() []int {
	return n.indices
}
