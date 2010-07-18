//
//	goray/std/integrators/trivial.go
//	goray
//
//	Created by Ross Light on 2010-06-03.
//

package trivial

import (
	"goray/core/color"
	"goray/core/integrator"
	"goray/core/ray"
	"goray/core/render"
	"goray/core/scene"
)

type trivial struct{}

func New() integrator.SurfaceIntegrator        { return new(trivial) }
func (ti *trivial) SurfaceIntegrator()         {}
func (ti *trivial) Preprocess(sc *scene.Scene) {}

func (ti *trivial) Integrate(sc *scene.Scene, s *render.State, r ray.Ray) color.AlphaColor {
	if coll := sc.Intersect(r, -1); coll.Hit() {
		return color.NewRGBAFromColor(color.White, 1.0)
	}
	return color.NewRGBAFromColor(color.Gray(0.1), 0.0)
}
