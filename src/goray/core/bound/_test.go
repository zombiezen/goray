package bound

import "testing"

import (
	"math"
	"goray/core/vector"
)

type crossTest struct {
	From, Dir vector.Vector3D
	Expected  bool
}

func TestCross(t *testing.T) {
	box := Bound{vector.Vector3D{-1, -1, -1}, vector.Vector3D{1, 1, 1}}

	tests := []crossTest{
		crossTest{vector.Vector3D{0, 0, 0}, vector.Vector3D{1, 0, 0}, true},
		crossTest{vector.Vector3D{0, 0, 0}, vector.Vector3D{0, 1, 0}, true},
		crossTest{vector.Vector3D{0, 0, 0}, vector.Vector3D{0, 0, 1}, true},

		crossTest{vector.Vector3D{2, 0, 0}, vector.Vector3D{-1, 0, 0}, true},
		crossTest{vector.Vector3D{0, 2, 0}, vector.Vector3D{0, -1, 0}, true},
		crossTest{vector.Vector3D{0, 0, 2}, vector.Vector3D{0, 0, -1}, true},
		crossTest{vector.Vector3D{-2, 0, 0}, vector.Vector3D{1, 0, 0}, true},
		crossTest{vector.Vector3D{0, -2, 0}, vector.Vector3D{0, 1, 0}, true},
		crossTest{vector.Vector3D{0, 0, -2}, vector.Vector3D{0, 0, 1}, true},

		crossTest{vector.Vector3D{2, 0, 0}, vector.Vector3D{1, 0, 0}, false},
		crossTest{vector.Vector3D{0, 2, 0}, vector.Vector3D{0, 1, 0}, false},
		crossTest{vector.Vector3D{0, 0, 2}, vector.Vector3D{0, 0, 1}, false},
		crossTest{vector.Vector3D{-2, 0, 0}, vector.Vector3D{-1, 0, 0}, false},
		crossTest{vector.Vector3D{0, -2, 0}, vector.Vector3D{0, -1, 0}, false},
		crossTest{vector.Vector3D{0, 0, -2}, vector.Vector3D{0, 0, -1}, false},

		crossTest{vector.Vector3D{2, 2, 2}, vector.Vector3D{-1, -1, -1}, true},
		crossTest{vector.Vector3D{2, 2, 2}, vector.Vector3D{1, 1, 1}, false},
		crossTest{vector.Vector3D{-2, -2, -2}, vector.Vector3D{-1, -1, -1}, false},
		crossTest{vector.Vector3D{-2, -2, -2}, vector.Vector3D{1, 1, 1}, true},
	}

	for _, ct := range tests {
		if _, _, result := box.Cross(ct.From, ct.Dir, math.Inf(1)); result != ct.Expected {
			t.Errorf("Failed for From=%v Dir=%v (got %t)", ct.From, ct.Dir, result)
		}
	}
}

func TestRealCross(t *testing.T) {
	ct := crossTest{vector.Vector3D{0, 0, 5}, vector.Vector3D{-0.23640189135082473, 0.2234736629175765, -0.9456075654032989}, true}
	box := Bound{vector.Vector3D{-1.367188, -0.046875, 0.257812}, vector.Vector3D{-0.859375, 0.984375, 0.851562}}
	a, b, hit := box.Cross(ct.From, ct.Dir, math.Inf(1))
	aTarget := 4.387060924402294
	bTarget := 4.404881484235866

	if hit != ct.Expected {
		t.Error("Did not collide")
	} else {
		if a != aTarget {
			t.Errorf("a = %#v (wanted %#v)", a, aTarget)
		}
		if b != bTarget {
			t.Errorf("b = %#v (wanted %#v)", b, bTarget)
		}
	}
}
