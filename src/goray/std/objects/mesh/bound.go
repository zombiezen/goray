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

package mesh

import (
	"math"
	"goray/vector"
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

	axisTest := func(axis vector.Axis, i, j, edgeNum int) bool {
		var axis1, axis2 vector.Axis
		var a, b, fa, fb float64
		var p1, p2, rad, min, max float64
		var v1, v2 []float64

		switch axis {
		case vector.X:
			axis1, axis2 = vector.Y, vector.Z
		case vector.Y:
			axis1, axis2 = vector.X, vector.Z
		case vector.Z:
			axis1, axis2 = vector.X, vector.Y
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
		for axis := vector.X; axis <= vector.Z; axis++ {
			v[i][axis] = verts[i][axis] - boxcenter[axis]
		}
	}

	// Compute triangle edges
	for axis := vector.X; axis <= vector.Z; axis++ {
		e[0][axis] = v[1][axis] - v[0][axis]
		e[1][axis] = v[2][axis] - v[1][axis]
		e[2][axis] = v[0][axis] - v[2][axis]
	}

	// Run the nine tests
	f = [3]float64{math.Fabs(e[0][vector.X]), math.Fabs(e[0][vector.Y]), math.Fabs(e[0][vector.Z])}
	if !axisTest(vector.X, 0, 1, 0) {
		return false
	}
	if !axisTest(vector.Y, 0, 2, 0) {
		return false
	}
	if !axisTest(vector.Z, 1, 2, 0) {
		return false
	}

	f = [3]float64{math.Fabs(e[1][vector.X]), math.Fabs(e[1][vector.Y]), math.Fabs(e[1][vector.Z])}
	if !axisTest(vector.X, 0, 1, 1) {
		return false
	}
	if !axisTest(vector.Y, 0, 2, 1) {
		return false
	}
	if !axisTest(vector.Z, 0, 1, 1) {
		return false
	}

	f = [3]float64{math.Fabs(e[2][vector.X]), math.Fabs(e[2][vector.Y]), math.Fabs(e[2][vector.Z])}
	if !axisTest(vector.X, 0, 2, 2) {
		return false
	}
	if !axisTest(vector.Y, 0, 1, 2) {
		return false
	}
	if !axisTest(vector.Z, 1, 2, 2) {
		return false
	}

	// First test overlap in the x,y,z directions
	// This is equivalent to testing a minimal AABB
	min, max = v[0][vector.X], v[0][vector.X]
	if v[1][vector.X] < min {
		min = v[1][vector.X]
	}
	if v[1][vector.X] > max {
		max = v[1][vector.X]
	}
	if v[2][vector.X] < min {
		min = v[2][vector.X]
	}
	if v[2][vector.X] > max {
		max = v[2][vector.X]
	}
	if min > boxhalfsize[vector.X] || max < -boxhalfsize[vector.X] {
		return false
	}

	min, max = v[0][vector.Y], v[0][vector.Y]
	if v[1][vector.Y] < min {
		min = v[1][vector.Y]
	}
	if v[1][vector.Y] > max {
		max = v[1][vector.Y]
	}
	if v[2][vector.Y] < min {
		min = v[2][vector.Y]
	}
	if v[2][vector.Y] > max {
		max = v[2][vector.Y]
	}
	if min > boxhalfsize[vector.Y] || max < -boxhalfsize[vector.Y] {
		return false
	}

	min, max = v[0][vector.Z], v[0][vector.Z]
	if v[1][vector.Z] < min {
		min = v[1][vector.Z]
	}
	if v[1][vector.Z] > max {
		max = v[1][vector.Z]
	}
	if v[2][vector.Z] < min {
		min = v[2][vector.Z]
	}
	if v[2][vector.Z] > max {
		max = v[2][vector.Z]
	}
	if min > boxhalfsize[vector.Z] || max < -boxhalfsize[vector.Z] {
		return false
	}

	// Test if the box intersects the plane of the triangle
	// Plane equation of triangle: normal * x + d = 0
	normal[vector.X] = e[0][1]*e[1][2] - e[0][2]*e[1][1]
	normal[vector.Y] = e[0][2]*e[1][0] - e[0][0]*e[1][2]
	normal[vector.Z] = e[0][0]*e[1][1] - e[0][1]*e[1][0]
	return planeBoxOverlap(normal, v[0], boxhalfsize)
}
