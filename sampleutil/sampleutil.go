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

// Package sampleutil provides useful sampling functions.
package sampleutil

import (
	"bitbucket.org/zombiezen/goray/vector"
	"math"
	"sort"
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

// Cone uniformly samples a cone.
func Cone(d, u, v vector.Vector3D, maxCosAngle, s1, s2 float64) vector.Vector3D {
	cosAngle := 1 - (1-maxCosAngle)*s2
	sinAngle := math.Sqrt(1 - cosAngle*cosAngle)
	t1 := 2 * math.Pi * s1

	// \sin \theta (\vec{u} \cos t_1 + \vec{v} \cos t_1) + \vec{d} \cos \theta
	return vector.Add(vector.ScalarMul(vector.Add(vector.ScalarMul(u, math.Cos(t1)), vector.ScalarMul(v, math.Sin(t1))), sinAngle), vector.ScalarMul(d, cosAngle))
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
	F, Cdf   []float64
	Integral float64
}

// NewPdf1D creates a new probably distribution function and calculates its
// cumulative distribution function.
func NewPdf1D(f []float64) (p Pdf1D) {
	p.F = f
	p.Cdf = make([]float64, 1, len(f)+1)
	p.Cdf[0] = 0.0
	delta := 1.0 / float64(len(p.F))
	for i := range p.F {
		p.Integral += p.F[i] * delta
		p.Cdf = append(p.Cdf, p.Integral)
	}
	for i := range p.Cdf {
		p.Cdf[i] /= p.Integral
	}
	return
}

func (p Pdf1D) Len() int { return len(p.F) }

func (p Pdf1D) Sample(u float64) (offset, pdf float64) {
	index := sort.Search(len(p.Cdf), func(i int) bool { return p.Cdf[i] >= u }) - 1
	delta := (u - p.Cdf[index]) / (p.Cdf[index+1] - p.Cdf[index])
	return float64(index) + delta, p.F[index] / p.Integral
}

func (p Pdf1D) DiscreteSample(u float64) (index int, pdf float64) {
	if u == 0 {
		pdf = p.F[0] / p.Integral
		return
	}
	index = sort.Search(len(p.Cdf), func(i int) bool { return p.Cdf[i] >= u }) - 1
	pdf = p.F[index] / p.Integral
	return
}
