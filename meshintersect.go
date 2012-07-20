// +build !amd64

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
	"bitbucket.org/zombiezen/math3/vec64"
)

// Pure Go implementation of ray intersection

func intersect(a, b, c, rDir, rFrom vec64.Vector) (rayDepth, u, v float64) {
	// Tomas Möller and Ben Trumbore ray intersection scheme
	// Explanation: <http://softsurfer.com/Archive/algorithm_0105/algorithm_0105.htm#Segment-Triangle>
	rayDepth = -1.0
	edge1 := vec64.Sub(b, a)
	edge2 := vec64.Sub(c, a)
	pvec := vec64.Cross(rDir, edge2)
	det := vec64.Dot(edge1, pvec)
	if det == 0.0 {
		return
	}
	invDet := 1.0 / det
	tvec := vec64.Sub(rFrom, a)
	u = vec64.Dot(pvec, tvec) * invDet
	if u < 0.0 || u > 1.0 {
		return
	}
	qvec := vec64.Cross(tvec, edge1)
	v = vector.Dot(rDir, qvec) * invDet
	if v < 0.0 || u+v > 1.0 {
		return
	}
	rayDepth = vec64.Dot(edge2, qvec) * invDet
	return
}
