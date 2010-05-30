//
//  goray/integrator.go
//  goray
//
//  Created by Ross Light on 2010-05-29.
//

package integrator

import (
    "os"
    "./goray/color"
    "./goray/ray"
    "./goray/render"
)

const (
    SurfaceType = iota
    VolumeType
)

type Integrator interface {
    GetType() int
    SetScene(interface{})
    Preprocess() (os.Error)
    Render() <-chan render.Fragment
    Integrate(*render.State, *ray.Ray) color.AlphaColor
}

type SurfaceIntegrator interface {
    Integrator
}

type VolumeIntegrator interface {
    Integrator
    Transmittance(*render.State, *ray.Ray) color.AlphaColor
}
