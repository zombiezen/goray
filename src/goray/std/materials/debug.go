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

package debug

import (
	"os"

	"goray"
	"goray/color"
	"goray/vector"
	"goray/std/materials/common"
	"goray/std/yamlscene"

	yamldata "goyaml.googlecode.com/hg/data"
)

type debugMaterial struct {
	Color color.Color
}

var _ goray.Material = &debugMaterial{}

func New(col color.Color) goray.Material { return &debugMaterial{col} }

func (mat *debugMaterial) InitBSDF(state *goray.RenderState, sp goray.SurfacePoint) goray.BSDF {
	return goray.BSDFDiffuse
}

func (mat *debugMaterial) Eval(state *goray.RenderState, sp goray.SurfacePoint, wo, wl vector.Vector3D, types goray.BSDF) color.Color {
	return mat.Color
}

func (mat *debugMaterial) Sample(state *goray.RenderState, sp goray.SurfacePoint, wo vector.Vector3D, s *goray.MaterialSample) (color.Color, vector.Vector3D) {
	s.Pdf = 1.0
	return mat.Color, vector.Reflect(wo, sp.Normal)
}

func (mat *debugMaterial) Pdf(state *goray.RenderState, sp goray.SurfacePoint, wo, wi vector.Vector3D, bsdfs goray.BSDF) float64 {
	return 0.0
}

func (mat *debugMaterial) Specular(state *goray.RenderState, sp goray.SurfacePoint, wo vector.Vector3D) (reflect, refract bool, dir [2]vector.Vector3D, col [2]color.Color) {
	return
}

func (mat *debugMaterial) Reflectivity(state *goray.RenderState, sp goray.SurfacePoint, flags goray.BSDF) color.Color {
	return common.GetReflectivity(mat, state, sp, flags)
}

func (mat *debugMaterial) Alpha(state *goray.RenderState, sp goray.SurfacePoint, wo vector.Vector3D) float64 {
	return 1.0
}

func (mat *debugMaterial) ScatterPhoton(state *goray.RenderState, sp goray.SurfacePoint, wi vector.Vector3D, s *goray.PhotonSample) (wo vector.Vector3D, scattered bool) {
	return common.ScatterPhoton(mat, state, sp, wi, s)
}

func (mat *debugMaterial) MaterialFlags() goray.BSDF {
	return goray.BSDFDiffuse
}

func init() {
	yamlscene.Constructor[yamlscene.StdPrefix+"materials/debug"] = yamlscene.MapConstruct(Construct)
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
