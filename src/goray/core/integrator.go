//
//  goray/core/integrator.go
//  goray
//
//  Created by Ross Light on 2010-05-29.
//

/* The integrator package provides the interface for rendering methods. */
package integrator

import (
	"goray/core/color"
	"goray/core/ray"
	"goray/core/render"
	"goray/core/scene"
)

/* An Integrator renders rays. */
type Integrator interface {
	Preprocess(scene *scene.Scene)
	Integrate(scene *scene.Scene, state *render.State, r ray.Ray) color.AlphaColor
}

/* A SurfaceIntegrator renders rays by casting onto solid objects. */
type SurfaceIntegrator interface {
	Integrator
	SurfaceIntegrator()
}

/* A VolumeIntegrator renders rays by casting onto volumetric regions. */
type VolumeIntegrator interface {
	Integrator
	VolumeIntegrator()
	Transmittance(scene *scene.Scene, state *render.State, r ray.Ray) color.AlphaColor
}
