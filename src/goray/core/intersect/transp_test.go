package intersect

import "testing"

import (
	"math"
	"goray/core/color"
	"goray/core/material"
	"goray/core/primitive"
	"goray/core/ray"
	"goray/core/render"
	"goray/core/surface"
	"goray/core/vector"
	"goray/std/primitives/sphere"
)

type TestMat struct {
	Transp color.Color
}

func (mat TestMat) InitBSDF(state *render.State, sp surface.Point) material.BSDF { return material.BSDFNone }

func (mat TestMat) GetFlags() material.BSDF { return material.BSDFNone }

func (mat TestMat) Eval(state *render.State, sp surface.Point, wo, wl vector.Vector3D, types material.BSDF) color.Color {
	return color.Black
}

func (mat TestMat) Sample(state *render.State, sp surface.Point, wo vector.Vector3D, s *material.Sample) (color.Color, vector.Vector3D) {
	return color.Black, vector.Vector3D{0, 0, 0}
}

func (mat TestMat) Pdf(state *render.State, sp surface.Point, wo, wi vector.Vector3D, bsdfs material.BSDF) float64 { return 0 }

func (mat TestMat) GetSpecular(state *render.State, sp surface.Point, wo vector.Vector3D) (reflect, refract bool, dir [2]vector.Vector3D, col [2]color.Color) {
	return
}

func (mat TestMat) GetReflectivity(state *render.State, sp surface.Point, flags material.BSDF) color.Color {
	return color.Black
}

func (mat TestMat) GetAlpha(state *render.State, sp surface.Point, wo vector.Vector3D) float64 {
	if mat.Transp == nil {
		return 1.0
	}
	return 0
}

func (mat TestMat) ScatterPhoton(state *render.State, sp surface.Point, wi vector.Vector3D, s *material.PhotonSample) (wo vector.Vector3D, scattered bool) { return }

func (mat TestMat) GetTransparency(state *render.State, sp surface.Point, wo vector.Vector3D) color.Color {
	if mat.Transp == nil {
		return color.Black
	}
	return mat.Transp
}

func TestTransparentShadow(t *testing.T) {
	type tsTestCase struct{
		Name string
		Filters []color.Color
		Depth int

		Expected color.Color
		ShouldHit bool
	}

	cases := []tsTestCase{
		{
			Name: "Pass-through",
			Filters: []color.Color{},
			Depth: 3,
			Expected: color.NewRGB(1.0, 1.0, 1.0),
			ShouldHit: false,
		},
		{
			Name: "Red filter",
			Filters: []color.Color{
				color.NewRGB(1.0, 0.0, 0.0),
			},
			Depth: 3,
			Expected: color.NewRGB(1.0, 0.0, 0.0),
			ShouldHit: false,
		},
		{
			Name: "Depth boundary",
			Filters: []color.Color{
				color.NewRGB(0.0, 0.0, 1.0),
			},
			Depth: 1,
			Expected: color.NewRGB(0.0, 0.0, 1.0),
			ShouldHit: false,
		},
		{
			Name: "Depth clamp",
			Filters: []color.Color{
				color.NewRGB(1.0, 0.0, 0.0),
				color.NewRGB(1.0, 0.0, 0.0),
			},
			Depth: 1,
			Expected: nil,
			ShouldHit: true,
		},
		{
			Name: "3-Filter",
			Filters: []color.Color{
				color.NewRGB(1.0, 0.5, 0.5),
				color.NewRGB(0.5, 1.0, 0.5),
				color.NewRGB(0.5, 0.5, 1.0),
			},
			Depth: 5,
			Expected: color.NewRGB(0.25, 0.25, 0.25),
			ShouldHit: false,
		},
	}

	for _, c := range cases {
		primitives := make([]primitive.Primitive, 0, len(c.Filters))
		for i, f := range c.Filters {
			primitives = append(primitives, sphere.New(vector.Vector3D{float64(i + 1), 0, 0}, 0.5, TestMat{f}))
		}
		intersect := NewKD(primitives, nil)
		var col color.Color = color.Gray(1.0)
		r := ray.Ray{
			From: vector.Vector3D{0, 0, 0},
			Dir: vector.Vector3D{1, 0, 0},
			TMin: 0,
			TMax: math.Inf(1),
		}
		hit := intersect.DoTransparentShadows(nil, r, c.Depth, r.TMax, &col)
		switch {
		case hit != c.ShouldHit:
			t.Errorf("%s intersect hit mismatch", c.Name)
		case !c.ShouldHit && (col.GetR() != c.Expected.GetR() || col.GetG() != c.Expected.GetG() || col.GetB() != c.Expected.GetB()):
			t.Errorf("%s intersect got %v (wanted %v)", c.Name, col, c.Expected)
		}
	}
}
