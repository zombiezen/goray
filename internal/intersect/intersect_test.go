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

package intersect

import (
	"math"
	"testing"

	"bitbucket.org/zombiezen/math3/vec64"
	"zombiezen.com/go/goray/internal/goray"
	"zombiezen.com/go/goray/internal/primitives/sphere"
)

func TestDepth(t *testing.T) {
	type depthTestCase struct {
		Name      string
		Intersect goray.Intersecter
	}

	r := goray.Ray{
		From: vec64.Vector{2, 0, 0},
		Dir:  vec64.Vector{-1, 0, 0},
		TMax: -1.0,
	}

	sphereA := sphere.New(vec64.Vector{1, 0, 0}, 0.25, nil)
	sphereB := sphere.New(vec64.Vector{0, 0, 0}, 0.25, nil)
	sphereC := sphere.New(vec64.Vector{-1, 0, 0}, 0.25, nil)

	cases := []depthTestCase{
		{"Simple", NewSimple(
			[]goray.Primitive{sphereB, sphereC, sphereA},
		)},
		{"kd-tree", NewKD(
			[]goray.Primitive{sphereB, sphereC, sphereA},
			nil,
		)},
		{"kd-tree leaf", NewKD(
			[]goray.Primitive{sphereB, sphereA},
			nil,
		)},
	}

	for _, c := range cases {
		coll := c.Intersect.Intersect(r, math.Inf(1))
		if coll.Hit() {
			if coll.Primitive != sphereA {
				t.Errorf("%s intersect fails depth test", c.Name)
			}
		} else {
			t.Errorf("%s intersect won't collide; depth test skipped.", c.Name)
		}
	}
}
