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

	"goray"
	"goray/color"
	"goray/std/primitives/sphere"
	"goray/vector"
)

type TestMat struct {
	Transp color.Color
}

func (mat TestMat) InitBSDF(state *goray.RenderState, sp goray.SurfacePoint) goray.BSDF {
	return goray.BSDFNone
}

func (mat TestMat) MaterialFlags() goray.BSDF { return goray.BSDFNone }

func (mat TestMat) Eval(state *goray.RenderState, sp goray.SurfacePoint, wo, wl vector.Vector3D, types goray.BSDF) color.Color {
	return color.Black
}

func (mat TestMat) Sample(state *goray.RenderState, sp goray.SurfacePoint, wo vector.Vector3D, s *goray.MaterialSample) (color.Color, vector.Vector3D) {
	return color.Black, vector.Vector3D{0, 0, 0}
}

func (mat TestMat) Pdf(state *goray.RenderState, sp goray.SurfacePoint, wo, wi vector.Vector3D, bsdfs goray.BSDF) float64 {
	return 0
}

func (mat TestMat) Specular(state *goray.RenderState, sp goray.SurfacePoint, wo vector.Vector3D) (reflect, refract bool, dir [2]vector.Vector3D, col [2]color.Color) {
	return
}

func (mat TestMat) Reflectivity(state *goray.RenderState, sp goray.SurfacePoint, flags goray.BSDF) color.Color {
	return color.Black
}

func (mat TestMat) Alpha(state *goray.RenderState, sp goray.SurfacePoint, wo vector.Vector3D) float64 {
	if mat.Transp == nil {
		return 1.0
	}
	return 0
}

func (mat TestMat) ScatterPhoton(state *goray.RenderState, sp goray.SurfacePoint, wi vector.Vector3D, s *goray.PhotonSample) (wo vector.Vector3D, scattered bool) {
	return
}

func (mat TestMat) Transparency(state *goray.RenderState, sp goray.SurfacePoint, wo vector.Vector3D) color.Color {
	if mat.Transp == nil {
		return color.Black
	}
	return mat.Transp
}

func TestTransparentShadow(t *testing.T) {
	type tsTestCase struct {
		Name    string
		Filters []color.Color
		Depth   int

		Expected  color.Color
		ShouldHit bool
	}

	cases := []tsTestCase{
		{
			Name:      "Pass-through",
			Filters:   []color.Color{},
			Depth:     3,
			Expected:  color.RGB{1.0, 1.0, 1.0},
			ShouldHit: false,
		},
		{
			Name: "Red filter",
			Filters: []color.Color{
				color.RGB{1.0, 0.0, 0.0},
			},
			Depth:     3,
			Expected:  color.RGB{1.0, 0.0, 0.0},
			ShouldHit: false,
		},
		{
			Name: "Depth boundary",
			Filters: []color.Color{
				color.RGB{0.0, 0.0, 1.0},
			},
			Depth:     1,
			Expected:  color.RGB{0.0, 0.0, 1.0},
			ShouldHit: false,
		},
		{
			Name: "Depth clamp",
			Filters: []color.Color{
				color.RGB{1.0, 0.0, 0.0},
				color.RGB{1.0, 0.0, 0.0},
			},
			Depth:     1,
			Expected:  nil,
			ShouldHit: true,
		},
		{
			Name: "Opaque",
			Filters: []color.Color{
				color.Black,
			},
			Depth:     3,
			Expected:  nil,
			ShouldHit: true,
		},
		{
			Name: "3-Filter",
			Filters: []color.Color{
				color.RGB{1.0, 0.5, 0.5},
				color.RGB{0.5, 1.0, 0.5},
				color.RGB{0.5, 0.5, 1.0},
			},
			Depth:     5,
			Expected:  color.RGB{0.25, 0.25, 0.25},
			ShouldHit: false,
		},
	}

	for _, c := range cases {
		primitives := make([]goray.Primitive, 0, len(c.Filters))
		for i, f := range c.Filters {
			primitives = append(primitives, sphere.New(vector.Vector3D{float64(i + 1), 0, 0}, 0.5, TestMat{f}))
		}
		intersect := NewKD(primitives, nil)
		r := goray.Ray{
			From: vector.Vector3D{0, 0, 0},
			Dir:  vector.Vector3D{1, 0, 0},
			TMin: 0,
			TMax: math.Inf(1),
		}
		col, hit := intersect.TransparentShadow(nil, r, c.Depth, r.TMax)
		switch {
		case hit != c.ShouldHit:
			t.Errorf("%s intersect hit mismatch", c.Name)
		case !c.ShouldHit && (col.Red() != c.Expected.Red() || col.Green() != c.Expected.Green() || col.Blue() != c.Expected.Blue()):
			t.Errorf("%s intersect got %v (wanted %v)", c.Name, col, c.Expected)
		case c.ShouldHit && !color.IsBlack(col):
			t.Errorf("%s intersect got %v (wanted black)", c.Name, col)
		}
	}
}
