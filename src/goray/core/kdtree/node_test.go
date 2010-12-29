package kdtree

import "testing"

func TestLeaf(t *testing.T) {
	myLeaf := newLeaf([]Value{1, 2, 3})
	if !myLeaf.IsLeaf() {
		t.Error("Leaf nodes claim that they are not leaves")
	}
	vals := myLeaf.GetValues()
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
	if myInt.IsLeaf() {
		t.Error("Interior node claims that it is a leaf")
	}
	if myInt.GetAxis() != 2 {
		t.Error("Interior node stores wrong axis")
	}
	if myInt.GetPivot() != 3.14 {
		t.Error("Interior node stores wrong pivot")
	}
	if child, _ := myInt.GetLeft().(*Leaf); child != leafA {
		t.Error("Left child not stored")
	}
	if child, _ := myInt.GetRight().(*Leaf); child != leafB {
		t.Error("Right child not stored")
	}
}
