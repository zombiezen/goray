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
	"bitbucket.org/zombiezen/goray/bound"
	"bitbucket.org/zombiezen/goray/vector"
)

type pointTree []vector.Vector3D

func (t pointTree) Len() int {
	return len(t)
}

func (t pointTree) Dimension(i int, axis vector.Axis) (float64, float64) {
	return t[i][axis], t[i][axis]
}

type boxTree []bound.Bound

func (t boxTree) Len() int {
	return len(t)
}

func (t boxTree) Dimension(i int, axis vector.Axis) (float64, float64) {
	return t[i].Min[axis], t[i].Max[axis]
}

func TestLeafTree(t *testing.T) {
	tree := New(pointTree{{-1, 0, 0}, {1, 0, 0}}, Options{
		MaxDepth:       DefaultOptions.MaxDepth,
		LeafSize:       2,
		FaultTolerance: DefaultOptions.FaultTolerance,
		ClipThreshold:  DefaultOptions.ClipThreshold,
	})
	if tree.Depth() != 0 {
		t.Error("Simple leaf tree creation fails")
	}
}

func TestBound(t *testing.T) {
	ptA, ptB := vector.Vector3D{1, 2, 3}, vector.Vector3D{4, 5, 6}
	b := New(boxTree{{ptA, ptB}}, DefaultOptions).Bound()
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

	b = New(pointTree{ptA, ptB}, DefaultOptions).Bound()
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

func TestTree(t *testing.T) {
	tree := New(pointTree{
		{-1, 0, 0},
		{1, 0, 0},
		{-2, 0, 0},
		{2, 0, 0},
	}, DefaultOptions)
	if tree.Depth() != 1 {
		t.Error("Creation failed")
	}
	if tree.root.IsLeaf() {
		t.Fatal("Tree root is not an interior node")
	}
	if tree.root.Axis() != 0 {
		t.Errorf("Wrong split axis (got %d)", tree.root.Axis())
	}
	if tree.root.Pivot() != -1 {
		t.Errorf("Wrong pivot value (got %.2f)", tree.root.Pivot())
	}
	if tree.root.Left() != nil {
		if tree.root.Left().IsLeaf() {
			if len(tree.root.Left().Indices()) != 1 {
				t.Error("Wrong number of values in left")
			}
		} else {
			t.Error("Left child is not a leaf")
		}
	} else {
		t.Error("Left child is nil")
	}
	if tree.root.Right() != nil {
		if tree.root.Right().IsLeaf() {
			if len(tree.root.Right().Indices()) != 3 {
				t.Error("Wrong number of values in right")
			}
		} else {
			t.Error("Right child is not a leaf")
		}
	} else {
		t.Error("Right child is nil")
	}
}
