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
	"sort"
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

// Sphere uniformly samples a sphere.
func Sphere(s1, s2 float64) (dir vector.Vector3D) {
	dir[vector.Z] = 1.0 - 2.0*s1
	r := 1.0 - dir[vector.Z]*dir[vector.Z]
	if r > 0.0 {
		r = math.Sqrt(r)
		a := 2 * math.Pi * s2
		dir[vector.X], dir[vector.Y] = math.Cos(a)*r, math.Sin(a)*r
	} else {
		dir[vector.X], dir[vector.Y] = 0.0, 0.0
	}
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

// Pdf1D stores a 1-dimensonal probability distribution function.
type Pdf1D struct {
	f, cdf   []float64
	integral float64
}

// NewPdf1D creates a new probably distribution function and calculates its
// cumulative distribution function.
func NewPdf1D(f []float64) (p Pdf1D) {
	p.f = f
	p.cdf = make([]float64, 1, len(f)+1)
	p.cdf[0] = 0.0
	delta := 1.0 / float64(len(p.f))
	for i := range p.f {
		p.integral += p.f[i] * delta
		p.cdf = append(p.cdf, p.integral)
	}
	for i := range p.cdf {
		p.cdf[i] /= p.integral
	}
	return
}

func (p Pdf1D) Sample(u float64) (offset, pdf float64) {
	index := sort.Search(len(p.cdf), func(i int) bool { return p.cdf[i] >= u }) - 1
	delta := (u - p.cdf[index]) / (p.cdf[index+1] - p.cdf[index])
	return float64(index) + delta, p.f[index] / p.integral
}

func (p Pdf1D) DiscreteSample(u float64) (index int, pdf float64) {
	if u == 0 {
		pdf = p.f[0] / p.integral
		return
	}
	index = sort.Search(len(p.cdf), func(i int) bool { return p.cdf[i] >= u }) - 1
	pdf = p.f[index] / p.integral
	return
}
