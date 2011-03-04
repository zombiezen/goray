// intersect.go - Pure Go implementation of ray intersection

package mesh

import (
	"goray/core/vector"
)

func intersect(a, b, c, rDir, rFrom [3]float64) (rayDepth, u, v float64) {
	// Tomas Möller and Ben Trumbore ray intersection scheme
	// Explanation: <http://softsurfer.com/Archive/algorithm_0105/algorithm_0105.htm#Segment-Triangle>
	rayDepth = -1.0
	edge1 := vector.Vector3D{b[vector.X] - a[vector.X], b[vector.Y] - a[vector.Y], b[vector.Z] - a[vector.Z]}
	edge2 := vector.Vector3D{c[vector.X] - a[vector.X], c[vector.Y] - a[vector.Y], c[vector.Z] - a[vector.Z]}
	pvec := vector.Vector3D{
		rDir[vector.Y]*edge2[vector.Z] - rDir[vector.Z]*edge2[vector.Y],
		rDir[vector.Z]*edge2[vector.X] - rDir[vector.X]*edge2[vector.Z],
		rDir[vector.X]*edge2[vector.Y] - rDir[vector.Y]*edge2[vector.X],
	}
	det := edge1[vector.X]*pvec[vector.X] + edge1[vector.Y]*pvec[vector.Y] + edge1[vector.Z]*pvec[vector.Z]
	if det == 0.0 {
		return
	}
	invDet := 1.0 / det
	tvec := vector.Vector3D{rFrom[vector.X] - a[vector.X], rFrom[vector.Y] - a[vector.Y], rFrom[vector.Z] - a[vector.Z]}
	u = (pvec[vector.X]*tvec[vector.X] + pvec[vector.Y]*tvec[vector.Y] + pvec[vector.Z]*tvec[vector.Z]) * invDet
	if u < 0.0 || u > 1.0 {
		return
	}
	qvec := vector.Vector3D{
		tvec[vector.Y]*edge1[vector.Z] - tvec[vector.Z]*edge1[vector.Y],
		tvec[vector.Z]*edge1[vector.X] - tvec[vector.X]*edge1[vector.Z],
		tvec[vector.X]*edge1[vector.Y] - tvec[vector.Y]*edge1[vector.X],
	}
	v = (rDir[vector.X]*qvec[vector.X] + rDir[vector.Y]*qvec[vector.Y] + rDir[vector.Z]*qvec[vector.Z]) * invDet
	if v < 0.0 || u+v > 1.0 {
		return
	}
	rayDepth = (edge2[vector.X]*qvec[vector.X] + edge2[vector.Y]*qvec[vector.Y] + edge2[vector.Z]*qvec[vector.Z]) * invDet
	return
}