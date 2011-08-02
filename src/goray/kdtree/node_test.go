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

import "testing"

func TestLeaf(t *testing.T) {
	myLeaf := newLeaf([]Value{1, 2, 3})
	if !myLeaf.Leaf() {
		t.Error("Leaf nodes claim that they are not leaves")
	}
	vals := myLeaf.Values()
	if len(vals) != 3 {
		t.Error("Leaf nodes don't store the right number of values")
	}
	for i := 0; i < 3; i++ {
		if vals[i].(int) != i+1 {
			t.Errorf("Leaf value %d is corrupted", i)
		}
	}
}

func TestInterior(t *testing.T) {
	leafA, leafB := newLeaf([]Value{}), newLeaf([]Value{})
	myInt := newInterior(2, 3.14, leafA, leafB)
	if myInt.Leaf() {
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
