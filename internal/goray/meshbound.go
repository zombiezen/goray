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
	"math"

	"zombiezen.com/go/goray/internal/vecutil"
)

// Triangle bound intersection methods
// Note that a lot of functionality gets inlined for efficiency reasons.

func planeBoxOverlap(normal, vert, maxbox [3]float64) bool {
	var vmin, vmax [3]float64

	for q := 0; q < 3; q++ {
		v := vert[q]
		if normal[q] > 0.0 {
			vmin[q] = -maxbox[q] - v
			vmax[q] = maxbox[q] - v
		} else {
			vmin[q] = maxbox[q] - v
			vmax[q] = -maxbox[q] - v
		}
	}

	if normal[0]*vmin[0]+normal[1]*vmin[1]+normal[2]*vmin[2] > 0 {
		return false
	}
	if normal[0]*vmax[0]+normal[1]*vmax[1]+normal[2]*vmax[2] >= 0 {
		return true
	}

	return false
}

func triBoxOverlap(boxcenter, boxhalfsize [3]float64, verts [3][3]float64) bool {
	var normal, f [3]float64
	var v, e [3][3]float64
	var min, max float64

	axisTest := func(axis vecutil.Axis, i, j, edgeNum int) bool {
		var axis1, axis2 vecutil.Axis
		var a, b, fa, fb float64
		var p1, p2, rad, min, max float64
		var v1, v2 []float64

		switch axis {
		case vecutil.X:
			axis1, axis2 = vecutil.Y, vecutil.Z
		case vecutil.Y:
			axis1, axis2 = vecutil.X, vecutil.Z
		case vecutil.Z:
			axis1, axis2 = vecutil.X, vecutil.Y
		}

		v1, v2 = v[i][:], v[j][:]
		a, b = e[edgeNum][axis2], e[edgeNum][axis1]
		fa, fb = f[axis2], f[axis1]

		p1 = -a*v1[axis1] + b*v1[axis2]
		p2 = -a*v2[axis1] + b*v2[axis2]
		if p1 < p2 {
			min, max = p1, p2
		} else {
			min, max = p2, p1
		}
		rad = fa*boxhalfsize[axis1] + fb*boxhalfsize[axis2]
		return min <= rad && max >= -rad
	}

	// Move everything so that the boxcenter is in (0, 0, 0)
	for i := 0; i < 3; i++ {
		for axis := vecutil.X; axis <= vecutil.Z; axis++ {
			v[i][axis] = verts[i][axis] - boxcenter[axis]
		}
	}

	// Compute triangle edges
	for axis := vecutil.X; axis <= vecutil.Z; axis++ {
		e[0][axis] = v[1][axis] - v[0][axis]
		e[1][axis] = v[2][axis] - v[1][axis]
		e[2][axis] = v[0][axis] - v[2][axis]
	}

	// Run the nine tests
	f = [3]float64{math.Abs(e[0][vecutil.X]), math.Abs(e[0][vecutil.Y]), math.Abs(e[0][vecutil.Z])}
	if !axisTest(vecutil.X, 0, 1, 0) {
		return false
	}
	if !axisTest(vecutil.Y, 0, 2, 0) {
		return false
	}
	if !axisTest(vecutil.Z, 1, 2, 0) {
		return false
	}

	f = [3]float64{math.Abs(e[1][vecutil.X]), math.Abs(e[1][vecutil.Y]), math.Abs(e[1][vecutil.Z])}
	if !axisTest(vecutil.X, 0, 1, 1) {
		return false
	}
	if !axisTest(vecutil.Y, 0, 2, 1) {
		return false
	}
	if !axisTest(vecutil.Z, 0, 1, 1) {
		return false
	}

	f = [3]float64{math.Abs(e[2][vecutil.X]), math.Abs(e[2][vecutil.Y]), math.Abs(e[2][vecutil.Z])}
	if !axisTest(vecutil.X, 0, 2, 2) {
		return false
	}
	if !axisTest(vecutil.Y, 0, 1, 2) {
		return false
	}
	if !axisTest(vecutil.Z, 1, 2, 2) {
		return false
	}

	// First test overlap in the x,y,z directions
	// This is equivalent to testing a minimal AABB
	min, max = v[0][vecutil.X], v[0][vecutil.X]
	if v[1][vecutil.X] < min {
		min = v[1][vecutil.X]
	}
	if v[1][vecutil.X] > max {
		max = v[1][vecutil.X]
	}
	if v[2][vecutil.X] < min {
		min = v[2][vecutil.X]
	}
	if v[2][vecutil.X] > max {
		max = v[2][vecutil.X]
	}
	if min > boxhalfsize[vecutil.X] || max < -boxhalfsize[vecutil.X] {
		return false
	}

	min, max = v[0][vecutil.Y], v[0][vecutil.Y]
	if v[1][vecutil.Y] < min {
		min = v[1][vecutil.Y]
	}
	if v[1][vecutil.Y] > max {
		max = v[1][vecutil.Y]
	}
	if v[2][vecutil.Y] < min {
		min = v[2][vecutil.Y]
	}
	if v[2][vecutil.Y] > max {
		max = v[2][vecutil.Y]
	}
	if min > boxhalfsize[vecutil.Y] || max < -boxhalfsize[vecutil.Y] {
		return false
	}

	min, max = v[0][vecutil.Z], v[0][vecutil.Z]
	if v[1][vecutil.Z] < min {
		min = v[1][vecutil.Z]
	}
	if v[1][vecutil.Z] > max {
		max = v[1][vecutil.Z]
	}
	if v[2][vecutil.Z] < min {
		min = v[2][vecutil.Z]
	}
	if v[2][vecutil.Z] > max {
		max = v[2][vecutil.Z]
	}
	if min > boxhalfsize[vecutil.Z] || max < -boxhalfsize[vecutil.Z] {
		return false
	}

	// Test if the box intersects the plane of the triangle
	// Plane equation of triangle: normal * x + d = 0
	normal[vecutil.X] = e[0][1]*e[1][2] - e[0][2]*e[1][1]
	normal[vecutil.Y] = e[0][2]*e[1][0] - e[0][0]*e[1][2]
	normal[vecutil.Z] = e[0][0]*e[1][1] - e[0][1]*e[1][0]
	return planeBoxOverlap(normal, v[0], boxhalfsize)
}
