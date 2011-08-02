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

import (
	"goray/bound"
	"goray/vector"
)

func dim(v Value, axis vector.Axis) (min, max float64) {
	switch val := v.(type) {
	case vector.Vector3D:
		comp := val[axis]
		return comp, comp
	case bound.Bound:
		return val.Min[axis], val.Max[axis]
	}
	return
}

func newPointTree(pts []vector.Vector3D, opts Options) *Tree {
	vals := make([]Value, len(pts))
	for i, p := range pts {
		vals[i] = p
	}
	return New(vals, opts)
}

func newBoxTree(boxes []bound.Bound, opts Options) *Tree {
	vals := make([]Value, len(boxes))
	for i, b := range boxes {
		vals[i] = b
	}
	return New(vals, opts)
}

func TestLeafTree(t *testing.T) {
	opts := MakeOptions(dim, nil)
	opts.LeafSize = 2
	tree := newPointTree([]vector.Vector3D{{-1, 0, 0}, {1, 0, 0}}, opts)
	if tree.Depth() != 0 {
		t.Error("Simple leaf tree creation fails")
	}
}

func TestBound(t *testing.T) {
	ptA, ptB := vector.Vector3D{1, 2, 3}, vector.Vector3D{4, 5, 6}

	opts := MakeOptions(dim, nil)
	b := newBoxTree([]bound.Bound{{ptA, ptB}}, opts).Bound()
	for axis := vector.X; axis <= vector.Z; axis++ {
		if b.Min[axis] != ptA[axis] {
			t.Errorf("Box tree %v minimum expects %.2f, got %.2f", axis, b.Min[axis], ptA[axis])
		}
	}
	for axis := vector.X; axis <= vector.Z; axis++ {
		if b.Max[axis] != ptB[axis] {
			t.Errorf("Box tree %v maximum expects %.2f, got %.2f", axis, b.Max[axis], ptB[axis])
		}
	}

	b = newPointTree([]vector.Vector3D{ptA, ptB}, opts).Bound()
	for axis := vector.X; axis <= vector.Z; axis++ {
		if b.Min[axis] != ptA[axis] {
			t.Errorf("Point tree %v minimum expects %.2f, got %.2f", axis, b.Min[axis], ptA[axis])
		}
	}
	for axis := vector.X; axis <= vector.Z; axis++ {
		if b.Max[axis] != ptB[axis] {
			t.Errorf("Point tree %v maximum expects %.2f, got %.2f", axis, b.Max[axis], ptB[axis])
		}
	}
}

func TestSimpleTree(t *testing.T) {
	opts := MakeOptions(dim, nil)
	opts.SplitFunc = SimpleSplit
	tree := newPointTree(
		[]vector.Vector3D{
			{-1, 0, 0},
			{1, 0, 0},
			{-2, 0, 0},
			{2, 0, 0},
		},
		opts,
	)
	if tree.Depth() != 1 {
		t.Error("Creation failed")
	}

	if tree.root.Leaf() {
		t.Fatal("Tree root is not an interior node")
	}
	if tree.root.Axis() != 0 {
		t.Errorf("Wrong split axis (got %d)", tree.root.Axis())
	}
	if tree.root.Pivot() != 1 {
		t.Errorf("Wrong pivot value (got %.2f)", tree.root.Pivot())
	}

	if tree.root.Left() != nil {
		if tree.root.Left().Leaf() {
			if len(tree.root.Left().Values()) != 2 {
				t.Error("Wrong number of values in left")
			}
		} else {
			t.Error("Left child is not a leaf")
		}
	} else {
		t.Error("Left child is nil")
	}

	if tree.root.Right() != nil {
		if tree.root.Right().Leaf() {
			if len(tree.root.Right().Values()) != 2 {
				t.Error("Wrong number of values in right")
			}
		} else {
			t.Error("Right child is not a leaf")
		}
	} else {
		t.Error("Right child is nil")
	}
}
