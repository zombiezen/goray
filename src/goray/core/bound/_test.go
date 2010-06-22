package bound

import "testing"

import (
	"goray/fmath"
	"goray/core/vector"
)

func TestCross(t *testing.T) {
	type crossTest struct {
		From, Dir vector.Vector3D
		Expected  bool
	}

	box := New(vector.New(-1, -1, -1), vector.New(1, 1, 1))

	tests := []crossTest{
		crossTest{vector.New(0, 0, 0), vector.New(1, 0, 0), true},
		crossTest{vector.New(0, 0, 0), vector.New(0, 1, 0), true},
		crossTest{vector.New(0, 0, 0), vector.New(0, 0, 1), true},

		crossTest{vector.New(2, 0, 0), vector.New(-1, 0, 0), true},
		crossTest{vector.New(0, 2, 0), vector.New(0, -1, 0), true},
		crossTest{vector.New(0, 0, 2), vector.New(0, 0, -1), true},
		crossTest{vector.New(-2, 0, 0), vector.New(1, 0, 0), true},
		crossTest{vector.New(0, -2, 0), vector.New(0, 1, 0), true},
		crossTest{vector.New(0, 0, -2), vector.New(0, 0, 1), true},

		crossTest{vector.New(2, 0, 0), vector.New(1, 0, 0), false},
		crossTest{vector.New(0, 2, 0), vector.New(0, 1, 0), false},
		crossTest{vector.New(0, 0, 2), vector.New(0, 0, 1), false},
		crossTest{vector.New(-2, 0, 0), vector.New(-1, 0, 0), false},
		crossTest{vector.New(0, -2, 0), vector.New(0, -1, 0), false},
		crossTest{vector.New(0, 0, -2), vector.New(0, 0, -1), false},

		crossTest{vector.New(2, 2, 2), vector.New(-1, -1, -1), true},
		crossTest{vector.New(2, 2, 2), vector.New(1, 1, 1), false},
		crossTest{vector.New(-2, -2, -2), vector.New(-1, -1, -1), false},
		crossTest{vector.New(-2, -2, -2), vector.New(1, 1, 1), true},
	}

	for _, ct := range tests {
		if _, _, result := box.Cross(ct.From, ct.Dir, fmath.Inf); result != ct.Expected {
			t.Errorf("Failed for From=%v Dir=%v (got %t)", ct.From, ct.Dir, result)
		}
	}
}
