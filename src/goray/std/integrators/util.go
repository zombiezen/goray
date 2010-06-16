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
	"goray/core/ray"
	"goray/core/render"
	"goray/core/scene"
	"goray/core/surface"
	"goray/core/vector"
)

/* EstimateDirectPH computes an estimate of direct lighting with multiple importance sampling using the power heuristic with exponent=2. */
func EstimateDirectPH(state *render.State, sp surface.Point, lights []light.Light, sc *scene.Scene, wo vector.Vector3D, trShad bool, sDepth int) (col color.Color) {
	col = color.NewRGB(0, 0, 0)
	lightRay := ray.New()
	lightRay.SetFrom(sp.Position)
	mat := sp.Material.(material.Material)
	
	for _, l := range lights {
		if diracLight, ok := l.(light.DiracLight); ok {
			// Light with delta distribution
			if lcol, ok := diracLight.Illuminate(sp, lightRay); ok {
				// Shadowed
				lightRay.SetTMin(0.0005) // TODO: Add a smart self-bias value
				var shadowed bool
				if trShad {
					// TODO
				} else {
					shadowed = sc.IsShadowed(state, lightRay)
				}
				if !shadowed {
					if trShad {
						//lcol = color.Mul(lcol, scol)
					}
					surfCol := mat.Eval(state, sp, wo, lightRay.Dir(), material.BSDFAll)
					//TODO: transmitCol
					col = color.Add(col, color.ScalarMul(color.Mul(surfCol, lcol), fmath.Abs(vector.Dot(sp.Normal, lightRay.Dir()))))
				}
			}
		} else {
			// Area light, etc.
			// TODO
		}
	}
	return
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
