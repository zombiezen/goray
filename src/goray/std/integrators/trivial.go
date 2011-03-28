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

var _ integrator.SurfaceIntegrator = trivial{}

func New() integrator.SurfaceIntegrator       { return trivial{} }
func (ti trivial) SurfaceIntegrator()         {}
func (ti trivial) Preprocess(sc *scene.Scene) {}

func (ti trivial) Integrate(sc *scene.Scene, s *render.State, r ray.DifferentialRay) color.AlphaColor {
	if coll := sc.Intersect(r.Ray, -1); coll.Hit() {
		return color.NewRGBAFromColor(color.White, 1.0)
	}
	return color.NewRGBAFromColor(color.Gray(0.1), 0.0)
}
