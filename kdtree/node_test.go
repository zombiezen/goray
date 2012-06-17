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
	"reflect"
	"testing"
)

func TestLeaf(t *testing.T) {
	expected := []int{1, 2, 3}
	input := make([]int, len(expected))
	copy(input, expected)
	myLeaf := newLeaf(input)
	if !myLeaf.IsLeaf() {
		t.Error("Leaf nodes claim that they are not leaves")
	}
	if indices := myLeaf.Indices(); !reflect.DeepEqual(indices, expected) {
		t.Errorf("myLeaf.Indices() != %v (got %v)", expected, indices)
	}
}

func TestInterior(t *testing.T) {
	leafA, leafB := newLeaf([]int{}), newLeaf([]int{})
	myInt := newInterior(2, 3.14, leafA, leafB)
	if myInt.IsLeaf() {
		t.Error("Interior node claims that it is a leaf")
	}
	if myInt.Axis() != 2 {
		t.Error("Interior node stores wrong axis")
	}
	if myInt.Pivot() != 3.14 {
		t.Error("Interior node stores wrong pivot")
	}
	if myInt.Left() != leafA {
		t.Error("Left child not stored")
	}
	if myInt.Right() != leafB {
		t.Error("Right child not stored")
	}
}
