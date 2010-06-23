//
//	goray/std/materials/debug.go
//	goray
//
//	Created by Ross Light on 2010-06-10.
//

package debug

import (
	"goray/core/color"
	"goray/core/material"
	"goray/core/render"
	"goray/core/surface"
	"goray/core/vector"
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
	return mat.Color, vector.New(0, 0, 0)
}

func (mat *debugMaterial) Pdf(state *render.State, sp surface.Point, wo, wi vector.Vector3D, bsdfs material.BSDF) float {
	return 0.0
}

func (mat *debugMaterial) GetSpecular(state *render.State, sp surface.Point, wo vector.Vector3D) (reflect, refract bool, dir [2]vector.Vector3D, col [2]color.Color) {
	return
}

func (mat *debugMaterial) GetReflectivity(state *render.State, sp surface.Point, flags material.BSDF) color.Color {
	return nil
}

func (mat *debugMaterial) GetAlpha(state *render.State, sp surface.Point, wo vector.Vector3D) float {
	return 1.0
}

func (mat *debugMaterial) ScatterPhoton(state *render.State, sp surface.Point, wi vector.Vector3D, s *material.PhotonSample) (wo vector.Vector3D, scattered bool) {
	return
}

func (mat *debugMaterial) GetFlags() material.BSDF {
	return material.BSDFDiffuse
}
