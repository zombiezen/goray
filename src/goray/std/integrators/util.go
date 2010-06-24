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
	params := directParams{state, sp, lights, sc, wo, trShad, sDepth}

	for _, l := range lights {
		var newCol color.Color
		switch realLight := l.(type) {
		case light.DiracLight:
			// Light with delta distribution
			newCol = estimateDiracDirect(params, realLight)
		default:
			// Area light, etc.
			newCol = estimateAreaDirect(params, realLight)
		}
		col = color.Add(col, newCol)
	}
	return
}

type directParams struct {
	State  *render.State
	Surf   surface.Point
	Lights []light.Light
	Scene  *scene.Scene
	Wo     vector.Vector3D
	TrShad bool
	SDepth int
}

func checkShadow(params directParams, r ray.Ray) bool {
	r.SetTMin(0.0005) // TODO: Add a smart self-bias value
	if params.TrShad {
		// TODO
	}
	return params.Scene.IsShadowed(r, fmath.Inf)
}

func estimateDiracDirect(params directParams, l light.DiracLight) color.Color {
	sp := params.Surf
	lightRay := ray.New()
	lightRay.SetFrom(sp.Position)
	mat := sp.Material.(material.Material)

	if lcol, ok := l.Illuminate(sp, lightRay); ok {
		if shadowed := checkShadow(params, lightRay); !shadowed {
			if params.TrShad {
				//lcol = color.Mul(lcol, scol)
			}
			surfCol := mat.Eval(params.State, sp, params.Wo, lightRay.Dir(), material.BSDFAll)
			//TODO: transmitCol
			return color.ScalarMul(
				color.Mul(surfCol, lcol),
				fmath.Abs(vector.Dot(sp.Normal, lightRay.Dir())),
			)
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

func estimateAreaDirect(params directParams, l light.Light) (ccol color.Color) {
	ccol = color.Black
	sp := params.Surf
	lightRay := ray.New()
	lightRay.SetFrom(sp.Position)

	n := l.NumSamples()
	if params.State.RayDivision > 1 {
		n /= params.State.RayDivision
		if n < 1 {
			n = 1
		}
	}
	// TODO: Add a unique offset for every light
	offset := uint(n*params.State.PixelSample) + params.State.SamplingOffset

	isect, canIntersect := l.(light.Intersecter)
	hal := montecarlo.NewHalton(3)
	hal.SetStart(offset - 1)

	// Sample from light
	for i := 0; i < n; i++ {
		lightSamp := light.Sample{
			S1: montecarlo.VanDerCorput(uint32(offset)+uint32(i), 0),
			S2: hal.Float(),
		}
		ccol = color.Add(ccol, sampleLight(params, l, canIntersect, lightRay, lightSamp))
	}
	ccol = color.ScalarDiv(ccol, float(n))

	// Sample from BSDF
	if canIntersect {
		ccol2 := color.Black
		for i := 0; i < n; i++ {
			hal1 := montecarlo.NewHalton(5)
			hal1.SetStart(offset + uint(i))
			s1 := hal1.Float()
			hal2 := montecarlo.NewHalton(7)
			hal2.SetStart(offset + uint(i))
			s2 := hal2.Float()

			ccol2 = color.Add(ccol2, sampleBSDF(params, isect, s1, s2))
		}
		ccol2 = color.ScalarDiv(ccol2, float(n))
		ccol = color.Add(ccol, ccol2)
	}

	return
}

const pdfCutoff = 1e-6

func sampleLight(params directParams, l light.Light, canIntersect bool, lightRay ray.Ray, lightSamp light.Sample) (col color.Color) {
	col = color.Black
	sp := params.Surf
	mat := sp.Material.(material.Material)

	if params.State.RayDivision > 1 {
		lightSamp.S1 = addMod1(lightSamp.S1, params.State.Dc1)
		lightSamp.S2 = addMod1(lightSamp.S2, params.State.Dc2)
	}

	if l.IlluminateSample(sp, lightRay, &lightSamp) {
		if shadowed := checkShadow(params, lightRay); !shadowed && lightSamp.Pdf > pdfCutoff {
			// TODO: if trShad
			// TODO: transmitCol
			surfCol := mat.Eval(params.State, sp, params.Wo, lightRay.Dir(), material.BSDFAll)
			col = color.ScalarMul(
				color.Mul(surfCol, lightSamp.Color),
				fmath.Abs(vector.Dot(sp.Normal, lightRay.Dir())),
			)
			if canIntersect {
				mPdf := mat.Pdf(
					params.State, sp, params.Wo, lightRay.Dir(),
					material.BSDFGlossy|material.BSDFDiffuse|material.BSDFDispersive|material.BSDFReflect|material.BSDFTransmit,
				)
				l2 := lightSamp.Pdf * lightSamp.Pdf
				m2 := mPdf * mPdf
				w := l2 / (l2 + m2)
				col = color.ScalarMul(col, w/lightSamp.Pdf)
			} else {
				col = color.ScalarDiv(col, lightSamp.Pdf)
			}
		}
	}
	return
}

func sampleBSDF(params directParams, l light.Intersecter, s1, s2 float) (col color.Color) {
	sp := params.Surf
	mat := sp.Material.(material.Material)
	bRay := ray.New()
	bRay.SetTMin(0.0005)
	bRay.SetFrom(sp.Position)

	if params.State.RayDivision > 1 {
		s1 = addMod1(s1, params.State.Dc1)
		s2 = addMod1(s2, params.State.Dc2)
	}
	s := material.NewSample(s1, s2)
	s.Flags = material.BSDFGlossy | material.BSDFDiffuse | material.BSDFDispersive | material.BSDFReflect | material.BSDFTransmit

	surfCol, wi := mat.Sample(params.State, sp, params.Wo, &s)
	bRay.SetDir(wi)

	if dist, lcol, lightPdf, ok := l.Intersect(bRay); s.Pdf > pdfCutoff && ok {
		bRay.SetTMax(dist)
		if !checkShadow(params, bRay) {
			// TODO: if trShad
			// TODO: transmitCol
			lPdf := 1.0 / lightPdf
			l2 := lPdf * lPdf
			m2 := s.Pdf * s.Pdf
			w := m2 / (l2 + m2)
			cos2 := fmath.Abs(vector.Dot(sp.Normal, bRay.Dir()))
			if s.Pdf > pdfCutoff {
				col = color.ScalarMul(
					color.Mul(surfCol, lcol),
					cos2*w/s.Pdf,
				)
			}
		}
	}

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
