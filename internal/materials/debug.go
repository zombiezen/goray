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

package materials

import (
	"errors"

	"bitbucket.org/zombiezen/math3/vec64"
	"zombiezen.com/go/goray/internal/color"
	"zombiezen.com/go/goray/internal/goray"
	yamldata "zombiezen.com/go/goray/internal/yaml/data"
	"zombiezen.com/go/goray/internal/yamlscene"
)

type debugMaterial struct {
	Color color.Color
}

var _ goray.Material = &debugMaterial{}

func NewDebug(col color.Color) goray.Material {
	return &debugMaterial{col}
}

func (mat *debugMaterial) InitBSDF(state *goray.RenderState, sp goray.SurfacePoint) goray.BSDF {
	return goray.BSDFDiffuse
}

func (mat *debugMaterial) Eval(state *goray.RenderState, sp goray.SurfacePoint, wo, wl vec64.Vector, types goray.BSDF) color.Color {
	return mat.Color
}

func (mat *debugMaterial) Sample(state *goray.RenderState, sp goray.SurfacePoint, wo vec64.Vector, s *goray.MaterialSample) (color.Color, vec64.Vector) {
	s.Pdf = 1.0
	return mat.Color, vec64.Reflect(wo, sp.Normal)
}

func (mat *debugMaterial) Pdf(state *goray.RenderState, sp goray.SurfacePoint, wo, wi vec64.Vector, bsdfs goray.BSDF) float64 {
	return 0.0
}

func (mat *debugMaterial) Specular(state *goray.RenderState, sp goray.SurfacePoint, wo vec64.Vector) (reflect, refract bool, dir [2]vec64.Vector, col [2]color.Color) {
	return
}

func (mat *debugMaterial) Reflectivity(state *goray.RenderState, sp goray.SurfacePoint, flags goray.BSDF) color.Color {
	return getReflectivity(mat, state, sp, flags)
}

func (mat *debugMaterial) Alpha(state *goray.RenderState, sp goray.SurfacePoint, wo vec64.Vector) float64 {
	return 1.0
}

func (mat *debugMaterial) ScatterPhoton(state *goray.RenderState, sp goray.SurfacePoint, wi vec64.Vector, s *goray.PhotonSample) (wo vec64.Vector, scattered bool) {
	return scatterPhoton(mat, state, sp, wi, s)
}

func (mat *debugMaterial) MaterialFlags() goray.BSDF {
	return goray.BSDFDiffuse
}

func init() {
	yamlscene.Constructor[yamlscene.StdPrefix+"materials/debug"] = yamlscene.MapConstruct(constructDebug)
}

func constructDebug(m yamldata.Map) (interface{}, error) {
	col, ok := m["color"].(color.Color)
	if !ok {
		return nil, errors.New("Color must be an RGB")
	}
	return NewDebug(col), nil
}
