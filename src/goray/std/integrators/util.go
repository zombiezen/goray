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

const (
	raySelfBias = 0.0005 // TODO: Make this not fixed
	pdfCutoff   = 1e-6
)

type colorFunc func(int) color.Color

// colorSum returns the sum of all of the colors returned by the function.
//
// The function given is called with [0, n) and it should return a color.
// If concurrent is true, then these calls happen concurrently.  The result of
// each function call is summed to get the final return value.
func colorSum(n int, concurrent bool, f colorFunc) (col color.Color) {
	col = color.Black
	if concurrent {
		// Set up channels
		channels := make([]chan color.Color, n)
		for i, _ := range channels {
			channels[i] = make(chan color.Color, 1)
		}
		// Start sampling
		for i := 0; i < n; i++ {
			go func(i int) {
				defer close(channels[i])
				channels[i] <- f(i)
			}(i)
		}
		// Collect samples
		for _, ch := range channels {
			col = color.Add(col, <-ch)
		}
	} else {
		for i := 0; i < n; i++ {
			col = color.Add(col, f(i))
		}
	}
	return
}

func sample(n int, f colorFunc) color.Color {
	return color.ScalarDiv(colorSum(n, true, f), float(n))
}

func halSeq(n int, base, start uint) (seq []float) {
	seq = make([]float, n)
	hal := montecarlo.NewHalton(base)
	hal.SetStart(start)
	for i, _ := range seq {
		seq[i] = hal.Float()
	}
	return
}

// EstimateDirectPH computes an estimate of direct lighting with multiple importance sampling using the power heuristic with exponent=2.
func EstimateDirectPH(state *render.State, sp surface.Point, lights []light.Light, sc *scene.Scene, wo vector.Vector3D, trShad bool, sDepth int) (col color.Color) {
	params := directParams{state, sp, lights, sc, wo, trShad, sDepth}

	return colorSum(len(lights), false, func(i int) (col color.Color) {
		switch l := lights[i].(type) {
		case light.DiracLight:
			// Light with delta distribution
			col = estimateDiracDirect(params, l)
		default:
			// Area light, etc.
			col = estimateAreaDirect(params, l)
		}
		return
	})
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
	r.TMin = raySelfBias
	if params.TrShad {
		// TODO
	}
	return params.Scene.IsShadowed(r, fmath.Inf)
}

func estimateDiracDirect(params directParams, l light.DiracLight) color.Color {
	sp := params.Surf
	lightRay := ray.Ray{
		From: sp.Position,
		TMax: -1.0,
	}
	mat := sp.Material.(material.Material)

	lcol, lightRay, ok := l.Illuminate(sp, lightRay)
	if ok {
		if shadowed := checkShadow(params, lightRay); !shadowed {
			if params.TrShad {
				//lcol = color.Mul(lcol, scol)
			}
			surfCol := mat.Eval(params.State, sp, params.Wo, lightRay.Dir, material.BSDFAll)
			//TODO: transmitCol
			return color.ScalarMul(
				color.Mul(surfCol, lcol),
				fmath.Abs(vector.Dot(sp.Normal, lightRay.Dir)),
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

	// Sample from light
	hals1 := halSeq(n, 3, offset-1)
	ccol = sample(n, func(i int) color.Color {
		lightSamp := light.Sample{
			S1: montecarlo.VanDerCorput(uint32(offset)+uint32(i), 0),
			S2: hals1[i],
		}
		return sampleLight(params, l, canIntersect, lightSamp)
	})

	// Sample from BSDF
	hals1 = halSeq(n, 5, offset)
	hals2 := halSeq(n, 7, offset)
	if canIntersect {
		ccol2 := sample(n, func(i int) color.Color {
			s1, s2 := hals1[i], hals2[i]
			return sampleBSDF(params, isect, s1, s2)
		})
		ccol = color.Add(ccol, ccol2)
	}

	return
}

func sampleLight(params directParams, l light.Light, canIntersect bool, lightSamp light.Sample) (col color.Color) {
	col = color.Black
	sp := params.Surf
	mat := sp.Material.(material.Material)

	if params.State.RayDivision > 1 {
		lightSamp.S1 = addMod1(lightSamp.S1, params.State.Dc1)
		lightSamp.S2 = addMod1(lightSamp.S2, params.State.Dc2)
	}

	lightRay := ray.Ray{ // Illuminate will fill in most of the ray
		From: sp.Position,
		TMax: -1.0,
	}
	lightRay, ok := l.IlluminateSample(sp, lightRay, &lightSamp)
	if ok {
		if shadowed := checkShadow(params, lightRay); !shadowed && lightSamp.Pdf > pdfCutoff {
			// TODO: if trShad
			// TODO: transmitCol
			surfCol := mat.Eval(params.State, sp, params.Wo, lightRay.Dir, material.BSDFAll)
			col = color.ScalarMul(
				color.Mul(surfCol, lightSamp.Color),
				fmath.Abs(vector.Dot(sp.Normal, lightRay.Dir)),
			)
			if canIntersect {
				mPdf := mat.Pdf(
					params.State, sp, params.Wo, lightRay.Dir,
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
	bRay := ray.Ray{
		From: sp.Position,
		TMin: raySelfBias,
		TMax: -1.0,
	}

	if params.State.RayDivision > 1 {
		s1 = addMod1(s1, params.State.Dc1)
		s2 = addMod1(s2, params.State.Dc2)
	}
	s := material.NewSample(s1, s2)
	s.Flags = material.BSDFGlossy | material.BSDFDiffuse | material.BSDFDispersive | material.BSDFReflect | material.BSDFTransmit

	surfCol, wi := mat.Sample(params.State, sp, params.Wo, &s)
	bRay.Dir = wi

	if dist, lcol, lightPdf, ok := l.Intersect(bRay); s.Pdf > pdfCutoff && ok {
		bRay.TMax = dist
		if !checkShadow(params, bRay) {
			// TODO: if trShad
			// TODO: transmitCol
			lPdf := 1.0 / lightPdf
			l2 := lPdf * lPdf
			m2 := s.Pdf * s.Pdf
			w := m2 / (l2 + m2)
			cos2 := fmath.Abs(vector.Dot(sp.Normal, bRay.Dir))
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

func SampleAO(sc *scene.Scene, state *render.State, sp surface.Point, wo vector.Vector3D, aoSamples int, aoDist float, aoColor color.Color) color.Color {
	mat := sp.Material.(material.Material)

	n := aoSamples
	if state.RayDivision > 1 {
		n /= state.RayDivision
		if n < 1 {
			n = 1
		}
	}

	offset := uint(n*state.PixelSample) + state.SamplingOffset

	hals := halSeq(n, 3, offset-1)
	return sample(n, func(i int) color.Color {
		s1 := montecarlo.VanDerCorput(uint32(offset)+uint32(i), 0)
		s2 := hals[i]
		if state.RayDivision > 1 {
			s1 = addMod1(s1, state.Dc1)
			s2 = addMod1(s2, state.Dc2)
		}
		lightRay := ray.Ray{
			From: sp.Position,
			TMin: raySelfBias,
			TMax: aoDist,
		}

		s := material.NewSample(s1, s2)
		s.Flags = material.BSDFDiffuse | material.BSDFReflect
		surfCol, dir := mat.Sample(state, sp, wo, &s)
		lightRay.Dir = dir

		if s.Pdf <= pdfCutoff || sc.IsShadowed(lightRay, fmath.Inf) {
			return color.Black
		}
		cos := fmath.Abs(vector.Dot(sp.Normal, lightRay.Dir))
		return color.ScalarMul(color.Mul(aoColor, surfCol), cos/s.Pdf)
	})
}
