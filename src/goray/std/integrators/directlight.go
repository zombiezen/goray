//
//  goray/std/integrators/directlight.go
//  goray
//
//  Created by Ross Light on 2010-06-06.
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
	"goray/core/vector"
	"goray/std/integrators/util"
	yamldata "yaml/data"
	"yaml/parser"
)

type directLighting struct {
	background            background.Background
	transparentShadows    bool
	shadowDepth, rayDepth int
	numPhotons, numSearch int

	caustics       bool
	causticsDepth  int
	causticsRadius float

	doAO      bool
	aoSamples int
	aoDist    float
	aoColor   color.Color

	lights []light.Light
}

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
	sceneLights := sc.GetLights()
	dl.lights = make([]light.Light, len(sceneLights), len(sceneLights)+1)
	copy(dl.lights, sceneLights)
	// Set up background
	dl.background = sc.GetBackground()
	if dl.background != nil {
		if bgLight := dl.background.GetLight(); bgLight != nil {
			dl.lights = dl.lights[0 : len(dl.lights)+1]
			dl.lights[len(dl.lights)-1] = bgLight
		}
	}
	return
}

func (dl *directLighting) Integrate(sc *scene.Scene, state *render.State, r ray.Ray) color.AlphaColor {
	col, alpha := color.Black, 0.0

	defer func(il bool) {
		state.IncludeLights = il
	}(state.IncludeLights)

	if coll := sc.Intersect(r, -1); coll.Hit() {
		sp := coll.GetSurface()
		// Camera ray
		if state.RayLevel == 0 {
			state.IncludeLights = true
		}

		mat := sp.Material.(material.Material)
		bsdfs := mat.InitBSDF(state, sp)
		wo := vector.ScalarMul(r.Dir(), -1)
		lightRay := ray.New()
		lightRay.SetFrom(sp.Position)

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

				reflect, refract, dir, rcol := mat.GetSpecular(state, sp, wo)
				if reflect {
					refRay := ray.NewDifferential()
					refRay.SetFrom(sp.Position)
					refRay.SetDir(dir[0])
					refRay.SetTMin(0.0005)

					integ := dl.Integrate(sc, state, refRay)
					// TODO: Multiply by volume integrator result
					col = color.Add(col, color.Mul(integ, rcol[0]))
				}
				if refract {
					refRay := ray.NewDifferential()
					refRay.SetFrom(sp.Position)
					refRay.SetDir(dir[1])
					refRay.SetTMin(0.0005)

					integ := dl.Integrate(sc, state, refRay)
					// TODO: Multiply by volume integrator result
					col, alpha = color.Add(col, color.Mul(integ, rcol[1])), integ.GetA()
				}
			}
		}
		state.RayLevel--

		matAlpha := mat.GetAlpha(state, sp, wo)
		alpha = matAlpha + (1-matAlpha)*alpha
	} else {
		// Nothing was hit, use the background.
		if dl.background != nil {
			col = color.Add(col, dl.background.GetColor(r, state, false))
		}
	}
	return color.NewRGBAFromColor(col, alpha)
}

func Construct(n parser.Node) (data interface{}, err os.Error) {
	node, ok := n.(*parser.Mapping)
	if !ok {
		err = os.NewError("Direct light integrator must have a mapping")
		return
	}
	
	m := node.Map()
	trShad, _ := yamldata.AsBool(m["transparentShadows"])
	shadowDepth, _ := yamldata.AsInt(m["shadowDepth"])
	rayDepth, _ := yamldata.AsInt(m["rayDepth"])
	data = New(trShad, int(shadowDepth), int(rayDepth))
	return
}
