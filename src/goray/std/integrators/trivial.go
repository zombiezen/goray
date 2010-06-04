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
	"./goray/vector"
)

type trivial struct {
	sc *scene.Scene
}

func New() integrator.SurfaceIntegrator    { return &trivial{} }
func (ti *trivial) SurfaceIntegrator()     {}
func (ti *trivial) SetScene(s interface{}) { ti.sc = s.(*scene.Scene) }
func (ti *trivial) Preprocess() os.Error   { return nil }

func (ti *trivial) Integrate(s *render.State, r ray.Ray) color.AlphaColor {
	if _, hit, _ := ti.sc.Intersect(r); hit {
		return color.NewRGBA(1.0, 1.0, 1.0, 1.0)
	}
	return color.NewRGBA(0.1, 0.1, 0.1, 0.0)
}

func (ti *trivial) Render() <-chan render.Fragment {
	cam := ti.sc.GetCamera()
	w, h := cam.ResolutionX(), cam.ResolutionY()
	ch := make(chan render.Fragment, 1000)
	control := make(chan bool)
	go func() {
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				go func(x, y int) {
					// Set up state
					state := &render.State{}
					state.Init(nil)
					state.PixelNumber = y*w + x
					state.ScreenPos = vector.New(2.0*float(x)/float(w)-1.0, -2.0*float(y)/float(h)+1.0, 0.0)
					state.Time = 0.0
					// Shoot ray
					r, _ := cam.ShootRay(float(x), float(y), 0, 0)
					// Integrate
					color := ti.Integrate(state, r)
					ch <- render.Fragment{X: x, Y: y, Color: color}
					control <- true
				}(x, y)
			}
		}
	}()
	go func() {
		pixrendered := 0
		pixcount := w * h
		for _ = range control {
			pixrendered++
			if pixrendered >= pixcount {
				close(control)
				close(ch)
			}
		}
	}()
	return ch
}
