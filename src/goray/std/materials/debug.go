//
//	goray/std/materials/debug.go
//	goray
//
//	Created by Ross Light on 2010-06-10.
//

package debug

import (
	"os"
	"goray/core/color"
	"goray/core/material"
	"goray/core/render"
	"goray/core/surface"
	"goray/core/vector"
	"goray/std/materials/common"
	yamldata "goyaml.googlecode.com/hg/data"
)

type debugMaterial struct {
	Color color.Color
}

func New(col color.Color) material.Material { return &debugMaterial{col} }

func (mat *debugMaterial) InitBSDF(state *render.State, sp surface.Point) material.BSDF {
	return material.BSDFDiffuse
}

func (mat *debugMaterial) Eval(state *render.State, sp surface.Point, wo, wl vector.Vector3D, types material.BSDF) color.Color {
	return mat.Color
}

func (mat *debugMaterial) Sample(state *render.State, sp surface.Point, wo vector.Vector3D, s *material.Sample) (color.Color, vector.Vector3D) {
	s.Pdf = 1.0
	return mat.Color, vector.Reflect(wo, sp.Normal)
}

func (mat *debugMaterial) Pdf(state *render.State, sp surface.Point, wo, wi vector.Vector3D, bsdfs material.BSDF) float64 {
	return 0.0
}

func (mat *debugMaterial) GetSpecular(state *render.State, sp surface.Point, wo vector.Vector3D) (reflect, refract bool, dir [2]vector.Vector3D, col [2]color.Color) {
	return
}

func (mat *debugMaterial) GetReflectivity(state *render.State, sp surface.Point, flags material.BSDF) color.Color {
	return common.GetReflectivity(mat, state, sp, flags)
}

func (mat *debugMaterial) GetAlpha(state *render.State, sp surface.Point, wo vector.Vector3D) float64 {
	return 1.0
}

func (mat *debugMaterial) ScatterPhoton(state *render.State, sp surface.Point, wi vector.Vector3D, s *material.PhotonSample) (wo vector.Vector3D, scattered bool) {
	return common.ScatterPhoton(mat, state, sp, wi, s)
}

func (mat *debugMaterial) GetFlags() material.BSDF {
	return material.BSDFDiffuse
}

func Construct(m yamldata.Map) (data interface{}, err os.Error) {
	col, ok := m["color"].(color.Color)
	if !ok {
		err = os.NewError("Color must be an RGB")
		return
	}

	data = New(col)
	return
}
