package bound

import "testing"

import (
	"math"
	"goray/core/vector"
)

func TestCross(t *testing.T) {
	type crossTest struct {
		From, Dir vector.Vector3D
		Expected  bool
	}

	box := New(vector.Vector3D{-1, -1, -1}, vector.Vector3D{1, 1, 1})

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
