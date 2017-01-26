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

package integrators

import (
	"zombiezen.com/go/goray/internal/color"
	"zombiezen.com/go/goray/internal/goray"
	yamldata "zombiezen.com/go/goray/internal/yaml/data"
	"zombiezen.com/go/goray/internal/yamlscene"
)

type directLighting struct {
	background            goray.Background
	transparentShadows    bool
	shadowDepth, rayDepth int
	numPhotons, numSearch int

	caustics       bool
	causticsDepth  int
	causticsRadius float64

	doAO      bool
	aoSamples int
	aoDist    float64
	aoColor   color.Color

	lights []goray.Light
}

// NewDirectLight creates a new direct lighting integrator.
func NewDirectLight(transparentShadows bool, shadowDepth, rayDepth int) goray.SurfaceIntegrator {
	return &directLighting{
		transparentShadows: transparentShadows,
		shadowDepth:        shadowDepth,
		rayDepth:           rayDepth,
		causticsRadius:     0.25,
		causticsDepth:      10,
		numPhotons:         100000,
		numSearch:          100,
	}
}

func (dl *directLighting) SurfaceIntegrator() {}

func (dl *directLighting) Preprocess(sc *goray.Scene) {
	// Add lights
	sceneLights := sc.Lights()
	dl.lights = make([]goray.Light, len(sceneLights), len(sceneLights)+1)
	copy(dl.lights, sceneLights)
	// Set up background
	dl.background = sc.Background()
	if dl.background != nil {
		if bgLight := dl.background.Light(); bgLight != nil {
			dl.lights = append(dl.lights, bgLight)
		}
	}
	return
}

func (dl *directLighting) Integrate(sc *goray.Scene, state *goray.RenderState, r goray.DifferentialRay) color.AlphaColor {
	col, alpha := color.Black, 0.0

	defer func(il bool) {
		state.IncludeLights = il
	}(state.IncludeLights)

	if coll := sc.Intersect(r.Ray, -1); coll.Hit() {
		sp := coll.Surface()

		// Camera ray
		if state.RayLevel == 0 {
			state.IncludeLights = true
		}

		mat := sp.Material.(goray.Material)
		bsdfs := mat.InitBSDF(state, sp)
		wo := r.Dir.Negate()

		// Contribution of light-emitting surfaces
		if emat, ok := mat.(goray.EmitMaterial); ok {
			col = color.Add(col, emat.Emit(state, sp, wo))
		}

		// Normal lighting
		if bsdfs&(goray.BSDFGlossy|goray.BSDFDiffuse|goray.BSDFDispersive) != 0 {
			col = color.Add(col, estimateDirectPH(state, sp, dl.lights, sc, wo, dl.transparentShadows, dl.shadowDepth))
		}
		if bsdfs&(goray.BSDFDiffuse|goray.BSDFGlossy) != 0 {
			// TODO: estimatePhotons
		}
		if bsdfs&goray.BSDFDiffuse != 0 && dl.doAO {
			col = color.Add(col, sampleAO(sc, state, sp, wo, dl.aoSamples, dl.aoDist, dl.aoColor))
		}

		state.RayLevel++
		if state.RayLevel <= dl.rayDepth {
			// Dispersive effects with recursive raytracing
			if bsdfs&goray.BSDFDispersive != 0 && state.Chromatic {
				// TODO
			}

			// Glossy reflection with recursive raytracing
			if bsdfs&goray.BSDFGlossy != 0 {
				// TODO
			}

			// Perfect specular reflection/refraction with recursive raytracing
			{
				state.IncludeLights = true

				reflect, refract, dir, rcol := mat.Specular(state, sp, wo)
				if reflect {
					refRay := goray.DifferentialRay{
						Ray: goray.Ray{
							From: sp.Position,
							Dir:  dir[0],
							TMin: 0.0005,
							TMax: -1.0,
						},
					}

					integ := dl.Integrate(sc, state, refRay)
					// TODO: Multiply by volume integrator result
					col = color.Add(col, color.Mul(integ, rcol[0]))
				}
				if refract {
					refRay := goray.DifferentialRay{
						Ray: goray.Ray{
							From: sp.Position,
							Dir:  dir[1],
							TMin: 0.0005,
							TMax: -1.0,
						},
					}

					integ := dl.Integrate(sc, state, refRay)
					// TODO: Multiply by volume integrator result
					col, alpha = color.Add(col, color.Mul(integ, rcol[1])), integ.Alpha()
				}
			}
		}
		state.RayLevel--

		matAlpha := mat.Alpha(state, sp, wo)
		alpha = matAlpha + (1-matAlpha)*alpha
	} else {
		// Nothing was hit, use the background.
		if dl.background != nil {
			col = color.Add(col, dl.background.Color(r.Ray, state, false))
		}
	}
	return color.NewRGBAFromColor(col, alpha)
}

func init() {
	yamlscene.Constructor[yamlscene.StdPrefix+"integrators/directlight"] = yamlscene.MapConstruct(constructDirectLight)
}

func constructDirectLight(m yamldata.Map) (interface{}, error) {
	trShad, _ := yamldata.AsBool(m["transparentShadows"])
	shadowDepth, _ := yamldata.AsInt(m["shadowDepth"])
	rayDepth, _ := yamldata.AsInt(m["rayDepth"])
	return NewDirectLight(trShad, shadowDepth, rayDepth), nil
}
