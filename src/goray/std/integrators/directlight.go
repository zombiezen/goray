//
//  goray/std/integrators/directlight.go
//  goray
//
//  Created by Ross Light on 2010-06-06.
//

package directlight

import (
	"goray/core/background"
	"goray/core/color"
	"goray/core/integrator"
	"goray/core/light"
	"goray/core/ray"
	"goray/core/render"
	"goray/core/scene"
)

type directLighting struct {
    background background.Background
	transparentShadows bool
	shadowDepth, rayDepth int
	numPhotons, numSearch int
	
	caustics bool
	causticsDepth int
	causticsRadius float
	
	doAO bool
	aoSamples int
	aoColor color.Color
	
	lights []light.Light
}

func New(transparentShadows bool, shadowDepth, rayDepth int) integrator.SurfaceIntegrator {
	return &directLighting{
		transparentShadows: transparentShadows,
		shadowDepth: shadowDepth,
		rayDepth: rayDepth,
		causticsRadius: 0.25,
		causticsDepth: 10,
		numPhotons: 100000,
		numSearch: 100,
	}
}

func (dl *directLighting) SurfaceIntegrator() {}

func (dl *directLighting) Preprocess(sc *scene.Scene) {
	dl.background = sc.GetBackground()
	// TODO: Get lights
	return
}

func (dl *directLighting) Integrate(sc *scene.Scene, state *render.State, r ray.Ray) color.AlphaColor {
	// TODO
	return color.NewRGBA(0.0, 0.0, 0.0, 0.0)
}
