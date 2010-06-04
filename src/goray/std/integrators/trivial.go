//
//  goray/std/integrators/trivial.go
//  goray
//
//  Created by Ross Light on 2010-06-03.
//

package trivial

import (
    "os"
    "./goray/color"
    "./goray/integrator"
    "./goray/ray"
    "./goray/render"
    "./goray/scene"
)

type trivial struct {
    sc *scene.Scene
}

func New() integrator.SurfaceIntegrator { return &trivial{} }
func (ti *trivial) SurfaceIntegrator() {}
func (ti *trivial) SetScene(s interface{}) { ti.sc = s.(*scene.Scene) }
func (ti *trivial) Preprocess() os.Error { return nil }

func (ti *trivial) Integrate(s *render.State, r ray.Ray) color.AlphaColor {
    if _, hit, _ := ti.sc.Intersect(r); hit {
        return color.NewRGBA(1.0, 1.0, 1.0, 1.0)
    }
    return color.NewRGBA(0.1, 0.1, 0.1, 0.0)
}

func (ti *trivial) Render() <-chan render.Fragment {
    cam := ti.sc.GetCamera()
    ch := make(chan render.Fragment)
    for x := 0; x < cam.ResolutionX(); x++ {
        for y := 0; y < cam.ResolutionY(); y++ {
            go func() {
                r, _ := cam.ShootRay(float(x), float(y), 0, 0)
                color := ti.Integrate(nil, r)
                ch <- render.Fragment{X: x, Y: y, Color: color}
            }()
        }
    }
    return ch
}
