package kdtree

import "testing"

import (
	"goray/core/bound"
	"goray/core/vector"
)

func dim(v Value, axis int) (min, max float64) {
	switch val := v.(type) {
	case vector.Vector3D:
		comp := val[axis]
		return comp, comp
	case *bound.Bound:
		return val.GetMin()[axis], val.GetMax()[axis]
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

func newBoxTree(boxes []*bound.Bound, opts Options) *Tree {
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
	b := newBoxTree([]*bound.Bound{bound.New(ptA, ptB)}, opts).GetBound()
	for axis := vector.X; axis <= vector.Z; axis++ {
		if b.GetMin()[axis] != ptA[axis] {
			t.Errorf("Box tree %v minimum expects %.2f, got %.2f", b.GetMin()[axis], ptA[axis])
		}
	}
	for axis := vector.X; axis <= vector.Z; axis++ {
		if b.GetMin()[axis] != ptA[axis] {
			t.Errorf("Box tree %v maximum expects %.2f, got %.2f", b.GetMax()[axis], ptB[axis])
		}
	}

	b = newPointTree([]vector.Vector3D{ptA, ptB}, opts).GetBound()
	for axis := vector.X; axis <= vector.Z; axis++ {
		if b.GetMin()[axis] != ptA[axis] {
			t.Errorf("Point tree %v minimum expects %.2f, got %.2f", b.GetMin()[axis], ptA[axis])
		}
	}
	for axis := vector.X; axis <= vector.Z; axis++ {
		if b.GetMin()[axis] != ptA[axis] {
			t.Errorf("Point tree %v maximum expects %.2f, got %.2f", b.GetMax()[axis], ptB[axis])
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

	if tree.root.IsLeaf() {
		t.Fatal("Tree root is not an interior node")
	}
	root := tree.root.(*Interior)
	if root.GetAxis() != 0 {
		t.Errorf("Wrong split axis (got %d)", root.GetAxis())
	}
	if root.GetPivot() != 1 {
		t.Errorf("Wrong pivot value (got %.2f)", root.GetPivot())
	}

	if root.GetLeft() != nil {
		if leaf, ok := root.GetLeft().(*Leaf); ok {
			if len(leaf.GetValues()) != 2 {
				t.Error("Wrong number of values in left")
			}
		} else {
			t.Error("Left child is not a leaf")
		}
	} else {
		t.Error("Left child is nil")
	}

	if root.GetRight() != nil {
		if leaf, ok := root.GetRight().(*Leaf); ok {
			if len(leaf.GetValues()) != 2 {
				t.Error("Wrong number of values in right")
			}
		} else {
			t.Error("Right child is not a leaf")
		}
	} else {
		t.Error("Right child is nil")
	}
}
