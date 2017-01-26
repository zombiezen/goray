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

package bound

import (
	"math"
	"testing"

	"bitbucket.org/zombiezen/math3/vec64"
)

type crossTest struct {
	From, Dir vec64.Vector
	Expected  bool
}

func TestCross(t *testing.T) {
	box := Bound{vec64.Vector{-1, -1, -1}, vec64.Vector{1, 1, 1}}

	tests := []crossTest{
		crossTest{vec64.Vector{0, 0, 0}, vec64.Vector{1, 0, 0}, true},
		crossTest{vec64.Vector{0, 0, 0}, vec64.Vector{0, 1, 0}, true},
		crossTest{vec64.Vector{0, 0, 0}, vec64.Vector{0, 0, 1}, true},

		crossTest{vec64.Vector{2, 0, 0}, vec64.Vector{-1, 0, 0}, true},
		crossTest{vec64.Vector{0, 2, 0}, vec64.Vector{0, -1, 0}, true},
		crossTest{vec64.Vector{0, 0, 2}, vec64.Vector{0, 0, -1}, true},
		crossTest{vec64.Vector{-2, 0, 0}, vec64.Vector{1, 0, 0}, true},
		crossTest{vec64.Vector{0, -2, 0}, vec64.Vector{0, 1, 0}, true},
		crossTest{vec64.Vector{0, 0, -2}, vec64.Vector{0, 0, 1}, true},

		crossTest{vec64.Vector{2, 0, 0}, vec64.Vector{1, 0, 0}, false},
		crossTest{vec64.Vector{0, 2, 0}, vec64.Vector{0, 1, 0}, false},
		crossTest{vec64.Vector{0, 0, 2}, vec64.Vector{0, 0, 1}, false},
		crossTest{vec64.Vector{-2, 0, 0}, vec64.Vector{-1, 0, 0}, false},
		crossTest{vec64.Vector{0, -2, 0}, vec64.Vector{0, -1, 0}, false},
		crossTest{vec64.Vector{0, 0, -2}, vec64.Vector{0, 0, -1}, false},

		crossTest{vec64.Vector{2, 2, 2}, vec64.Vector{-1, -1, -1}, true},
		crossTest{vec64.Vector{2, 2, 2}, vec64.Vector{1, 1, 1}, false},
		crossTest{vec64.Vector{-2, -2, -2}, vec64.Vector{-1, -1, -1}, false},
		crossTest{vec64.Vector{-2, -2, -2}, vec64.Vector{1, 1, 1}, true},
	}

	for _, ct := range tests {
		if _, _, result := box.Cross(ct.From, ct.Dir, math.Inf(1)); result != ct.Expected {
			t.Errorf("Failed for From=%v Dir=%v (got %t)", ct.From, ct.Dir, result)
		}
	}
}

func TestRealCross(t *testing.T) {
	ct := crossTest{vec64.Vector{0, 0, 5}, vec64.Vector{-0.23640189135082473, 0.2234736629175765, -0.9456075654032989}, true}
	box := Bound{vec64.Vector{-1.367188, -0.046875, 0.257812}, vec64.Vector{-0.859375, 0.984375, 0.851562}}
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

func TestVolume(t *testing.T) {
	b := Bound{vec64.Vector{15, -27, 3}, vec64.Vector{20, -24, 7}}
	vol := b.Volume()
	if vol != 60 {
		t.Errorf("%#v != 60", vol)
	}
}

func TestSize(t *testing.T) {
	b := Bound{vec64.Vector{15, -27, 3}, vec64.Vector{20, -24, 7}}
	size := b.Size()
	if !(size[0] == 5 && size[1] == 3 && size[2] == 4) {
		t.Errorf("%#v != {5, 3, 4}", size)
	}
}
