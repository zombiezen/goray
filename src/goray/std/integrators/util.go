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
	"goray/montecarlo"
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
	col = color.Black

	for _, l := range lights {
		var newCol color.Color
		switch realLight := l.(type) {
		case light.DiracLight:
			// Light with delta distribution
			newCol = estimateDiracDirect(state, sp, realLight, sc, wo, trShad, sDepth)
		default:
			// Area light, etc.
			newCol = estimateAreaDirect(state, sp, l, sc, wo, trShad, sDepth)
		}
		col = color.Add(col, newCol)
	}
	return
}

func checkShadow(state *render.State, sc *scene.Scene, r ray.Ray, trShad bool, sDepth int) bool {
	r.SetTMin(0.0005) // TODO: Add a smart self-bias value
	if trShad {
		// TODO
	}
	return sc.IsShadowed(r, fmath.Inf)
}

func estimateDiracDirect(state *render.State, sp surface.Point, diracLight light.DiracLight, sc *scene.Scene, wo vector.Vector3D, trShad bool, sDepth int) color.Color {
	lightRay := ray.New()
	lightRay.SetFrom(sp.Position)
	mat := sp.Material.(material.Material)

	if lcol, ok := diracLight.Illuminate(sp, lightRay); ok {
		if shadowed := checkShadow(state, sc, lightRay, trShad, sDepth); !shadowed {
			if trShad {
				//lcol = color.Mul(lcol, scol)
			}
			surfCol := mat.Eval(state, sp, wo, lightRay.Dir(), material.BSDFAll)
			//TODO: transmitCol
			return color.ScalarMul(color.Mul(surfCol, lcol), fmath.Abs(vector.Dot(sp.Normal, lightRay.Dir())))
		}
	}

	return color.Black
}

func addMod1(a, b float) (s float) {
	s = a + b
	if s > 1 {
		s -= 1
	}
	return
}

func estimateAreaDirect(state *render.State, sp surface.Point, l light.Light, sc *scene.Scene, wo vector.Vector3D, trShad bool, sDepth int) (ccol color.Color) {
	ccol = color.Black
	lightRay := ray.New()
	lightRay.SetFrom(sp.Position)
	mat := sp.Material.(material.Material)

	n := l.NumSamples()
	if state.RayDivision > 1 {
		n /= state.RayDivision
		if n < 1 {
			n = 1
		}
	}
	// TODO: Add a unique offset for every light
	offset := uint(n*state.PixelSample) + state.SamplingOffset

	isect, canIntersect := l.(light.Intersecter)
	hal := montecarlo.NewHalton(3)
	hal.SetStart(offset - 1)

	for i := 0; i < n; i++ {
		lightSamp := light.Sample{
			S1: montecarlo.VanDerCorput(uint32(offset)+uint32(i), 0),
			S2: hal.Float(),
		}
		if state.RayDivision > 1 {
			lightSamp.S1 = addMod1(lightSamp.S1, state.Dc1)
			lightSamp.S2 = addMod1(lightSamp.S2, state.Dc2)
		}

		if l.IlluminateSample(sp, lightRay, &lightSamp) {
			if shadowed := checkShadow(state, sc, lightRay, trShad, sDepth); !shadowed && lightSamp.Pdf > 1e-6 {
				// TODO
			}
		}
	}
	// TODO: Remove when finished
	_, _, _ = mat, canIntersect, isect
	return
}

func EstimatePhotons(state *render.State, sp surface.Point, m *photon.Map, wo vector.Vector3D, nSearch int, radius float) (sum color.Color) {
	sum = color.Black
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
