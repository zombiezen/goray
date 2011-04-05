//
//	goray/std/integrators/directlight.go
//	goray
//
//	Created by Ross Light on 2010-06-06.
//

package directlight

import (
	"os"
	"goray/core/background"
	"goray/core/color"
	"goray/core/integrator"
	"goray/core/light"
	"goray/core/material"
	"goray/core/ray"
	"goray/core/render"
	"goray/core/scene"
	"goray/std/integrators/util"
	"goray/std/yamlscene"
	yamldata "goyaml.googlecode.com/hg/data"
)

type directLighting struct {
	background            background.Background
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

	lights []light.Light
}

var _ integrator.SurfaceIntegrator = &directLighting{}

func New(transparentShadows bool, shadowDepth, rayDepth int) integrator.SurfaceIntegrator {
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

func (dl *directLighting) Preprocess(sc *scene.Scene) {
	// Add lights
	sceneLights := sc.Lights()
	dl.lights = make([]light.Light, len(sceneLights), len(sceneLights)+1)
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

func (dl *directLighting) Integrate(sc *scene.Scene, state *render.State, r ray.DifferentialRay) color.AlphaColor {
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

		mat := sp.Material.(material.Material)
		bsdfs := mat.InitBSDF(state, sp)
		wo := r.Dir.Negate()

		// Contribution of light-emitting surfaces
		if emat, ok := mat.(material.EmitMaterial); ok {
			col = color.Add(col, emat.Emit(state, sp, wo))
		}
		// Normal lighting
		if bsdfs&(material.BSDFGlossy|material.BSDFDiffuse|material.BSDFDispersive) != 0 {
			col = color.Add(col, util.EstimateDirectPH(state, sp, dl.lights, sc, wo, dl.transparentShadows, dl.shadowDepth))
		}
		if bsdfs&(material.BSDFDiffuse|material.BSDFGlossy) != 0 {
			// TODO: EstimatePhotons
		}
		if bsdfs&material.BSDFDiffuse != 0 && dl.doAO {
			col = color.Add(col, util.SampleAO(sc, state, sp, wo, dl.aoSamples, dl.aoDist, dl.aoColor))
		}

		state.RayLevel++
		if state.RayLevel <= dl.rayDepth {
			// Dispersive effects with recursive raytracing
			if bsdfs&material.BSDFDispersive != 0 && state.Chromatic {
				// TODO
			}

			// Glossy reflection with recursive raytracing
			if bsdfs&material.BSDFGlossy != 0 {
				// TODO
			}

			// Perfect specular reflection/refraction with recursive raytracing
			{
				state.IncludeLights = true

				reflect, refract, dir, rcol := mat.Specular(state, sp, wo)
				if reflect {
					refRay := ray.DifferentialRay{
						Ray: ray.Ray{
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
					refRay := ray.DifferentialRay{
						Ray: ray.Ray{
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
	yamlscene.Constructor[yamlscene.StdPrefix+"integrators/directlight"] = yamlscene.MapConstruct(Construct)
}

func Construct(m yamldata.Map) (data interface{}, err os.Error) {
	trShad, _ := yamldata.AsBool(m["transparentShadows"])
	shadowDepth, _ := yamldata.AsInt(m["shadowDepth"])
	rayDepth, _ := yamldata.AsInt(m["rayDepth"])
	data = New(trShad, int(shadowDepth), int(rayDepth))
	return
}
