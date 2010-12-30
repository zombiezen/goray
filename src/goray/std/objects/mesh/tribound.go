//
//	goray/std/objects/mesh/tribound.go
//	goray
//
//	Created by Ross Light on 2010-12-26.
//

package mesh

import (
	"math"
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
	const (
		X = iota
		Y
		Z
	)
	var normal, f [3]float64
	var v, e [3][3]float64
	var min, max float64

	axisTest := func(axis, i, j, edgeNum int) bool {
		var axis1, axis2 int
		var a, b, fa, fb float64
		var p1, p2, rad, min, max float64
		var v1, v2 []float64

		switch axis {
		case X:
			axis1, axis2 = Y, Z
		case Y:
			axis1, axis2 = X, Z
		case Z:
			axis1, axis2 = X, Y
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
	v[0][X] = verts[0][X] - boxcenter[X]
	v[0][Y] = verts[0][Y] - boxcenter[Y]
	v[0][Z] = verts[0][Z] - boxcenter[Z]

	v[1][X] = verts[1][X] - boxcenter[X]
	v[1][Y] = verts[1][Y] - boxcenter[Y]
	v[1][Z] = verts[1][Z] - boxcenter[Z]

	v[2][X] = verts[2][X] - boxcenter[X]
	v[2][Y] = verts[2][Y] - boxcenter[Y]
	v[2][Z] = verts[2][Z] - boxcenter[Z]

	// Compute triangle edges
	e[0][X] = v[1][X] - v[0][X]
	e[0][Y] = v[1][Y] - v[0][Y]
	e[0][Z] = v[1][Z] - v[0][Z]

	e[1][X] = v[2][X] - v[1][X]
	e[1][Y] = v[2][Y] - v[1][Y]
	e[1][Z] = v[2][Z] - v[1][Z]

	e[2][X] = v[0][X] - v[2][X]
	e[2][Y] = v[0][Y] - v[2][Y]
	e[2][Z] = v[0][Z] - v[2][Z]

	// Run the nine tests
	f = [3]float64{math.Fabs(e[0][X]), math.Fabs(e[0][Y]), math.Fabs(e[0][Z])}
	if !axisTest(X, 0, 1, 0) {
		return false
	}
	if !axisTest(Y, 0, 2, 0) {
		return false
	}
	if !axisTest(Z, 1, 2, 0) {
		return false
	}

	f = [3]float64{math.Fabs(e[1][X]), math.Fabs(e[1][Y]), math.Fabs(e[1][Z])}
	if !axisTest(X, 0, 1, 1) {
		return false
	}
	if !axisTest(Y, 0, 2, 1) {
		return false
	}
	if !axisTest(Z, 0, 1, 1) {
		return false
	}

	f = [3]float64{math.Fabs(e[2][X]), math.Fabs(e[2][Y]), math.Fabs(e[2][Z])}
	if !axisTest(X, 0, 2, 2) {
		return false
	}
	if !axisTest(Y, 0, 1, 2) {
		return false
	}
	if !axisTest(Z, 1, 2, 2) {
		return false
	}

	// First test overlap in the x,y,z directions
	// This is equivalent to testing a minimal AABB
	min, max = v[0][X], v[0][X]
	if v[1][X] < min {
		min = v[1][X]
	}
	if v[1][X] > max {
		max = v[1][X]
	}
	if v[2][X] < min {
		min = v[2][X]
	}
	if v[2][X] > max {
		max = v[2][X]
	}
	if min > boxhalfsize[X] || max < -boxhalfsize[X] {
		return false
	}

	min, max = v[0][Y], v[0][Y]
	if v[1][Y] < min {
		min = v[1][Y]
	}
	if v[1][Y] > max {
		max = v[1][Y]
	}
	if v[2][Y] < min {
		min = v[2][Y]
	}
	if v[2][Y] > max {
		max = v[2][Y]
	}
	if min > boxhalfsize[Y] || max < -boxhalfsize[Y] {
		return false
	}

	min, max = v[0][Z], v[0][Z]
	if v[1][Z] < min {
		min = v[1][Z]
	}
	if v[1][Z] > max {
		max = v[1][Z]
	}
	if v[2][Z] < min {
		min = v[2][Z]
	}
	if v[2][Z] > max {
		max = v[2][Z]
	}
	if min > boxhalfsize[Z] || max < -boxhalfsize[Z] {
		return false
	}

	// Test if the box intersects the plane of the triangle
	// Plane equation of triangle: normal * x + d = 0
	normal[0] = e[0][1]*e[1][2] - e[0][2]*e[1][1]
	normal[1] = e[0][2]*e[1][0] - e[0][0]*e[1][2]
	normal[2] = e[0][0]*e[1][1] - e[0][1]*e[1][0]
	return planeBoxOverlap(normal, v[0], boxhalfsize)
}
