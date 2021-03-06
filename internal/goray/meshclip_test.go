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

package goray

import (
	"testing"

	"bitbucket.org/zombiezen/math3/vec64"
	"zombiezen.com/go/goray/internal/bound"
)

type boxClipTest struct {
	Min, Max        [3]float64
	PolyIn, PolyOut []vec64.Vector
	Region          bound.Bound
}

var boxTests = []boxClipTest{
	{
		Min:     [3]float64{-1, -1, -1},
		Max:     [3]float64{0, 0, 0},
		PolyIn:  []vec64.Vector{{0.1, 0.1, 0.1}, {0.9, 0.9, 0.9}, {0.9, 0.1, 0.9}, {0.1, 0.1, 0.1}},
		PolyOut: nil,
	},
	{
		Min:     [3]float64{0, 0, 0},
		Max:     [3]float64{1, 1, 1},
		PolyIn:  []vec64.Vector{{0.1, 0.1, 0.1}, {0.9, 0.9, 0.9}, {0.9, 0.1, 0.9}, {0.1, 0.1, 0.1}},
		PolyOut: []vec64.Vector{{0.9, 0.1, 0.9}, {0.1, 0.1, 0.1}, {0.1, 0.1, 0.1}, {0.9, 0.9, 0.9}},
		Region:  bound.Bound{vec64.Vector{0.1, 0.1, 0.1}, vec64.Vector{0.9, 0.9, 0.9}},
	},
}

func (test boxClipTest) Run(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("TEST PANIC: %v", err)
		}
	}()

	resultPoly, resultBound := triBoxClip(test.Min, test.Max, test.PolyIn)

	switch {
	case test.PolyOut == nil:
		if resultPoly != nil {
			t.Error("Resulting polygon not nil")
		}
	case resultPoly == nil:
		t.Error("Resulting polygon is nil")
	case len(resultPoly) != len(test.PolyOut):
		t.Errorf("Resulting polygon has %d polys, want %d.", len(resultPoly), len(test.PolyOut))
	default:
		for i, _ := range resultPoly {
			pa, pb := test.PolyOut[i], resultPoly[i]
			if pa[0] != pb[0] || pa[1] != pb[1] || pa[2] != pb[2] {
				t.Errorf("Vertex %d: expected %v, got %v", i, pa, pb)
			}
		}
	}

	if resultBound != test.Region {
		t.Errorf("Bound: %v, wanted %v", resultBound, test.Region)
	}
}

func TestBoxClip(t *testing.T) {
	for i, test := range boxTests {
		t.Logf("** Test [%d]", i)
		test.Run(t)
	}
}
