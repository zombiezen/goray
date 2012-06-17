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
	"bitbucket.org/zombiezen/goray/bound"
	"bitbucket.org/zombiezen/goray/vector"
	"reflect"
	"testing"
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
	expected := newInterior(vector.X, -1,
		newLeaf([]int{2}),
		newInterior(vector.X, 1,
			newLeaf([]int{0}),
			newLeaf([]int{1, 3})))
	if !reflect.DeepEqual(tree.root, expected) {
		t.Errorf("got %v\nexpected %v", tree, &Tree{root: expected})
	}
}
