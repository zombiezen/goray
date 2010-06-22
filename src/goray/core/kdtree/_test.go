package kdtree

import "testing"

import (
	"goray/core/bound"
	"goray/core/vector"
)

func dim(v Value, axis int) (min, max float) {
	switch val := v.(type) {
	case vector.Vector3D:
		comp := val.GetComponent(axis)
		return comp, comp
	case *bound.Bound:
		return val.GetMin().GetComponent(axis), val.GetMax().GetComponent(axis)
	}
	return
}

func newPointTree(pts []vector.Vector3D) *Tree {
	vals := make([]Value, len(pts))
	for i, p := range pts {
		vals[i] = p
	}
	return New(vals, BuildOptions{GetDimension: dim, MaxDepth: 4, LeafSize: 2})
}

func newBoxTree(boxes []*bound.Bound) *Tree {
	vals := make([]Value, len(boxes))
	for i, b := range boxes {
		vals[i] = b
	}
	return New(vals, BuildOptions{GetDimension: dim, MaxDepth: 4, LeafSize: 2})
}

func TestLeafTree(t *testing.T) {
	tree := newPointTree([]vector.Vector3D{vector.New(-1, 0, 0), vector.New(1, 0, 0)})
	if tree.Depth() != 0 {
		t.Error("Simple leaf tree creation fails")
	}
}

func TestBound(t *testing.T) {
	ptA, ptB := vector.New(1, 2, 3), vector.New(4, 5, 6)

	b := newBoxTree([]*bound.Bound{bound.New(ptA, ptB)}).GetBound()
	if b.GetMinX() != ptA.X || b.GetMinY() != ptA.Y || b.GetMinZ() != ptA.Z {
		t.Error("Box tree minimum wrong")
	}
	if b.GetMaxX() != ptB.X || b.GetMaxY() != ptB.Y || b.GetMaxZ() != ptB.Z {
		t.Error("Box tree maximum wrong")
	}

	b = newPointTree([]vector.Vector3D{ptA, ptB}).GetBound()
	if b.GetMinX() != ptA.X || b.GetMinY() != ptA.Y || b.GetMinZ() != ptA.Z {
		t.Error("Point tree minimum wrong")
	}
	if b.GetMaxX() != ptB.X || b.GetMaxY() != ptB.Y || b.GetMaxZ() != ptB.Z {
		t.Error("Point tree maximum wrong")
	}
}

//func TestSmallTree(t *testing.T) {
//	tree := newPointTree([]vector.Vector3D{
//		vector.New(-1, 0, 0),
//		vector.New(1, 0, 0),
//		vector.New(-2, 0, 0),
//		vector.New(2, 0, 0),
//	})
//	if tree.Depth() != 1 {
//		t.Error("Small tree creation failed")
//	}
//
//	if tree.root.IsLeaf() {
//		t.Fatal("Tree root is not an interior node")
//	}
//	root := tree.root.(*Interior)
//	if root.GetAxis() != 0 {
//		t.Errorf("Wrong split axis (got %d)", root.GetAxis())
//	}
//	if root.GetPivot() != 1 {
//		t.Errorf("Wrong pivot value (got %.2f)", root.GetPivot())
//	}
//
//	if root.GetLeft() != nil {
//		if leaf, ok := root.GetLeft().(*Leaf); ok {
//			if len(leaf.GetValues()) != 2 {
//				t.Error("Wrong number of values in left")
//			}
//		} else {
//			t.Error("Left child is not a leaf")
//		}
//	} else {
//		t.Error("Left child is nil")
//	}
//
//	if root.GetRight() != nil {
//		if leaf, ok := root.GetRight().(*Leaf); ok {
//			if len(leaf.GetValues()) != 2 {
//				t.Error("Wrong number of values in right")
//			}
//		} else {
//			t.Error("Right child is not a leaf")
//		}
//	} else {
//		t.Error("Right child is nil")
//	}
//}
