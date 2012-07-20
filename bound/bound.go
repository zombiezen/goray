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

// Package bound provides a bounding box type, along with various manipulation operations.
package bound

import (
	"bitbucket.org/zombiezen/goray/vecutil"
	"bitbucket.org/zombiezen/math3/vec64"
	"fmt"
	"math"
)

// Bound is a simple bounding box.
type Bound struct {
	Min, Max vec64.Vector
}

func (b Bound) String() string {
	return fmt.Sprintf("Bound{min: %v, max: %v}", b.Min, b.Max)
}

// IsZero returns whether the bound is the zero bound.  This is not the same as being empty.
func (b Bound) IsZero() bool { return b.Min.IsZero() && b.Max.IsZero() }

// Union creates a new bounding box that contains the two bounds.
func Union(b1, b2 Bound) Bound {
	newBound := Bound{}
	for axis := range newBound.Min {
		newBound.Min[axis] = math.Min(b1.Min[axis], b2.Min[axis])
		newBound.Max[axis] = math.Max(b1.Max[axis], b2.Max[axis])
	}
	return newBound
}

// Cross checks whether a given ray crosses the bound.
// from specifies a point where the ray starts.
// ray specifies the direction the ray is in.
// dist is the maximum distance that this method will check.  Pass in math.Inf(1) to remove the check.
func (b Bound) Cross(from, ray vec64.Vector, dist float64) (lmin, lmax float64, crosses bool) {
	p := vec64.Sub(from, b.Min)
	lmin, lmax = math.Inf(-1), math.Inf(1)
	crosses = true

	for axis := range b.Min {
		if ray[axis] != 0 {
			tmp1 := -p[axis] / ray[axis]
			tmp2 := (b.Max[axis] - b.Min[axis] - p[axis]) / ray[axis]
			if tmp1 > tmp2 {
				tmp1, tmp2 = tmp2, tmp1
			}
			if tmp1 > lmin {
				lmin = tmp1
			}
			if tmp2 < lmax {
				lmax = tmp2
			}
			if lmax < 0 || lmin > dist {
				return 0, 0, false
			}
		}
	}

	if lmin > lmax || lmax < 0 || lmin > dist {
		return 0, 0, false
	}
	return
}

// Volume calculates the volume of the bounding box
func (b Bound) Volume() float64 {
	return (b.Max[0] - b.Min[0]) * (b.Max[1] - b.Min[1]) * (b.Max[2] - b.Min[2])
}

func (b Bound) Size() [3]float64 {
	return [3]float64(vec64.Sub(b.Max, b.Min))
}

func (b Bound) LargestAxis() vecutil.Axis {
	s := b.Size()
	return vecutil.LargestAxis(s[0], s[1], s[2])
}

func (b Bound) HalfSize() [3]float64 {
	return [3]float64(vec64.Vector(b.Size()).Scale(0.5))
}

// Include modifies the bounding box so that it contains the specified point.
func (b Bound) Include(p vec64.Vector) Bound {
	for axis := range b.Min {
		b.Min[axis] = math.Min(b.Min[axis], p[axis])
		b.Max[axis] = math.Max(b.Max[axis], p[axis])
	}
	return b
}

// Includes returns whether a given point is in the bounding box.
func (b Bound) Includes(p vec64.Vector) bool {
	for axis := range b.Min {
		if p[axis] < b.Min[axis] || p[axis] > b.Max[axis] {
			return false
		}
	}
	return true
}

func (b Bound) Center() vec64.Vector {
	return vec64.Add(b.Max, b.Min).Scale(0.5)
}

// Grow increases the size of the bounding box by d on all sides.  The center will remain the same.
func (b Bound) Grow(d float64) Bound {
	for axis := range b.Min {
		b.Min[axis] -= d
		b.Max[axis] += d
	}
	return b
}
