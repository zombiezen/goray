//
//	goray/sampleutil.go
//	goray
//
//	Created by Ross Light on 2011-02-15.
//

// The sampleutil package provides useful sampling functions.
package sampleutil

import (
	"math"
	"goray/core/vector"
)

// CosHemisphere samples a cosine-weighted hemisphere given the coordinate system built by n, ru, and rv.
func CosHemisphere(n, ru, rv vector.Vector3D, s1, s2 float64) (v vector.Vector3D) {
	z1 := s1
	z2 := s2 * 2 * math.Pi
	v = vector.ScalarMul(vector.Add(vector.ScalarMul(ru, math.Cos(z2)), vector.ScalarMul(rv, math.Sin(z2))), math.Sqrt(1-z1))
	v = vector.Add(v, vector.ScalarMul(n, math.Sqrt(z1)))
	return
}

// AddMod1 performs an floating-point addition modulo 1. Both values must be in the range [0,1].
func AddMod1(a, b float64) (s float64) {
	s = a + b
	if s > 1 {
		s -= 1
	}
	return
}
