//
//  goray/integrator.go
//  goray
//
//  Created by Ross Light on 2010-05-29.
//

/* The goray/integrator package provides the interface for rendering methods. */
package integrator

import (
	"os"
    "./goray/camera"
	"./goray/color"
	"./goray/ray"
	"./goray/render"
)

/* A rendering system */
type Integrator interface {
	SetScene(interface{})
	Preprocess() os.Error
	Integrate(*render.State, ray.Ray) color.AlphaColor
}

type SurfaceIntegrator interface {
	Integrator
	SurfaceIntegrator()
}

type VolumeIntegrator interface {
	Integrator
	VolumeIntegrator()
	Transmittance(*render.State, ray.Ray) color.AlphaColor
}
