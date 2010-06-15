//
//	goray/std/integrators/util.go
//	goray
//
//	Created by Ross Light on 2010-06-10.
//

package util

import (
	"math"
	"goray/fmath"
	"goray/core/color"
	"goray/core/light"
	"goray/core/material"
	"goray/core/photon"
	"goray/core/render"
	"goray/core/scene"
	"goray/core/surface"
	"goray/core/vector"
)

/* EstimateDirectPH computes an estimate of direct lighting with multiple importance sampling using the power heuristic with exponent=2. */
func EstimateDirectPH(state *render.State, sp surface.Point, lights []light.Light, sc *scene.Scene, wo vector.Vector3D, trShad bool, sDepth int) color.Color {
	return nil
}

func EstimatePhotons(state *render.State, sp surface.Point, m *photon.Map, wo vector.Vector3D, nSearch int, radius float) (sum color.Color) {
	sum = color.NewRGB(0.0, 0.0, 0.0)
	if !m.Ready() {
		return
	}
	gathered := m.Gather(sp.Position, nSearch, radius)

	if len(gathered) > 0 {
		mat := sp.Material.(material.Material)
		for _, gResult := range gathered {
			phot := gResult.Photon
			surfCol := mat.Eval(state, sp, wo, phot.GetDirection(), material.BSDFAll)
			k := kernel(gResult.Distance, radius)
			sum = color.Add(sum, color.Mul(surfCol, color.ScalarMul(phot.GetColor(), k)))
		}
		sum = color.ScalarMul(sum, 1.0/float(m.GetNumPaths()))
	}
	return
}

func kernel(phot, gather float) float {
	s := 1 - phot/gather
	return 3.0 / (gather * math.Pi) * s * s
}

func ckernel(phot, gather float) float {
	p, g := fmath.Sqrt(phot), fmath.Sqrt(gather)
	return 3.0 * (1.0 - p/g) / (gather * math.Pi)
}
