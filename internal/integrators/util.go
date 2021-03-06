/*
	Copyright (c) 2011 Ross Light.
	Copyright (c) 2005 Mathias Wein, Alejandro Conty, and Alfredo de Greef.

	This file is part of goray.

	goray is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	goray is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with goray.  If not, see <http://www.gnu.org/licenses/>.
*/

package integrators

import (
	"math"

	"bitbucket.org/zombiezen/math3/vec64"
	"zombiezen.com/go/goray/internal/color"
	"zombiezen.com/go/goray/internal/goray"
	"zombiezen.com/go/goray/internal/montecarlo"
	"zombiezen.com/go/goray/internal/sampleutil"
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
	return color.ScalarDiv(colorSum(n, true, f), float64(n))
}

func halSeq(n int, base, start uint) (seq []float64) {
	seq = make([]float64, n)
	hal := montecarlo.NewHalton(base)
	hal.SetStart(start)
	for i, _ := range seq {
		seq[i] = hal.Float64()
	}
	return
}

// estimateDirectPH computes an estimate of direct lighting with multiple importance sampling using the power heuristic with exponent=2.
func estimateDirectPH(state *goray.RenderState, sp goray.SurfacePoint, lights []goray.Light, sc *goray.Scene, wo vec64.Vector, trShad bool, sDepth int) (col color.Color) {
	params := directParams{state, sp, lights, sc, wo, trShad, sDepth}

	return colorSum(len(lights), false, func(i int) (col color.Color) {
		switch l := lights[i].(type) {
		case goray.DiracLight:
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
	State  *goray.RenderState
	Surf   goray.SurfacePoint
	Lights []goray.Light
	Scene  *goray.Scene
	Wo     vec64.Vector
	TrShad bool
	SDepth int
}

func checkShadow(params directParams, r goray.Ray) bool {
	r.TMin = raySelfBias
	if params.TrShad {
		// TODO
	}
	return params.Scene.Shadowed(r, math.Inf(1))
}

func estimateDiracDirect(params directParams, l goray.DiracLight) color.Color {
	sp := params.Surf
	lightRay := goray.Ray{
		From: sp.Position,
		TMax: -1.0,
	}
	mat := sp.Material.(goray.Material)

	lcol, ok := l.Illuminate(sp, &lightRay)
	if ok {
		if shadowed := checkShadow(params, lightRay); !shadowed {
			if params.TrShad {
				//lcol = color.Mul(lcol, scol)
			}
			surfCol := mat.Eval(params.State, sp, params.Wo, lightRay.Dir, goray.BSDFAll)
			//TODO: transmitCol
			return color.ScalarMul(
				color.Mul(surfCol, lcol),
				math.Abs(vec64.Dot(sp.Normal, lightRay.Dir)),
			)
		}
	}

	return color.Black
}

func estimateAreaDirect(params directParams, l goray.Light) (ccol color.Color) {
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

	isect, canIntersect := l.(goray.LightIntersecter)

	// Sample from light
	hals1 := halSeq(n, 3, offset-1)
	ccol = sample(n, func(i int) color.Color {
		lightSamp := goray.LightSample{
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

func sampleLight(params directParams, l goray.Light, canIntersect bool, lightSamp goray.LightSample) (col color.Color) {
	col = color.Black
	sp := params.Surf
	mat := sp.Material.(goray.Material)

	if params.State.RayDivision > 1 {
		lightSamp.S1 = sampleutil.AddMod1(lightSamp.S1, params.State.Dc1)
		lightSamp.S2 = sampleutil.AddMod1(lightSamp.S2, params.State.Dc2)
	}

	lightRay := goray.Ray{ // Illuminate will fill in most of the ray
		From: sp.Position,
		TMax: -1.0,
	}
	if ok := l.IlluminateSample(sp, &lightRay, &lightSamp); ok {
		if shadowed := checkShadow(params, lightRay); !shadowed && lightSamp.Pdf > pdfCutoff {
			// TODO: if trShad
			// TODO: transmitCol
			surfCol := mat.Eval(params.State, sp, params.Wo, lightRay.Dir, goray.BSDFAll)
			col = color.ScalarMul(
				color.Mul(surfCol, lightSamp.Color),
				math.Abs(vec64.Dot(sp.Normal, lightRay.Dir)),
			)
			if canIntersect {
				mPdf := mat.Pdf(
					params.State, sp, params.Wo, lightRay.Dir,
					goray.BSDFGlossy|goray.BSDFDiffuse|goray.BSDFDispersive|goray.BSDFReflect|goray.BSDFTransmit,
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

func sampleBSDF(params directParams, l goray.LightIntersecter, s1, s2 float64) (col color.Color) {
	sp := params.Surf
	mat := sp.Material.(goray.Material)
	bRay := goray.Ray{
		From: sp.Position,
		TMin: raySelfBias,
		TMax: -1.0,
	}

	if params.State.RayDivision > 1 {
		s1 = sampleutil.AddMod1(s1, params.State.Dc1)
		s2 = sampleutil.AddMod1(s2, params.State.Dc2)
	}
	s := goray.NewMaterialSample(s1, s2)
	s.Flags = goray.BSDFGlossy | goray.BSDFDiffuse | goray.BSDFDispersive | goray.BSDFReflect | goray.BSDFTransmit

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
			cos2 := math.Abs(vec64.Dot(sp.Normal, bRay.Dir))
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

func estimatePhotons(state *goray.RenderState, sp goray.SurfacePoint, m *goray.PhotonMap, wo vec64.Vector, nSearch int, radius float64) (sum color.Color) {
	sum = color.Black
	if !m.Ready() {
		return
	}
	gathered := m.Gather(sp.Position, nSearch, radius)

	if len(gathered) > 0 {
		mat := sp.Material.(goray.Material)
		for _, gResult := range gathered {
			phot := gResult.Photon
			surfCol := mat.Eval(state, sp, wo, phot.Direction, goray.BSDFAll)
			k := kernel(gResult.Distance, radius)
			sum = color.Add(sum, color.Mul(surfCol, color.ScalarMul(phot.Color, k)))
		}
		sum = color.ScalarMul(sum, 1.0/float64(m.NumPaths()))
	}
	return
}

func kernel(phot, gather float64) float64 {
	s := 1 - phot/gather
	return 3.0 / (gather * math.Pi) * s * s
}

func ckernel(phot, gather float64) float64 {
	p, g := math.Sqrt(phot), math.Sqrt(gather)
	return 3.0 * (1.0 - p/g) / (gather * math.Pi)
}

func sampleAO(sc *goray.Scene, state *goray.RenderState, sp goray.SurfacePoint, wo vec64.Vector, aoSamples int, aoDist float64, aoColor color.Color) color.Color {
	mat := sp.Material.(goray.Material)

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
			s1 = sampleutil.AddMod1(s1, state.Dc1)
			s2 = sampleutil.AddMod1(s2, state.Dc2)
		}
		lightRay := goray.Ray{
			From: sp.Position,
			TMin: raySelfBias,
			TMax: aoDist,
		}

		s := goray.NewMaterialSample(s1, s2)
		s.Flags = goray.BSDFDiffuse | goray.BSDFReflect
		surfCol, dir := mat.Sample(state, sp, wo, &s)
		lightRay.Dir = dir

		if s.Pdf <= pdfCutoff || sc.Shadowed(lightRay, math.Inf(1)) {
			return color.Black
		}
		cos := math.Abs(vec64.Dot(sp.Normal, lightRay.Dir))
		return color.ScalarMul(color.Mul(aoColor, surfCol), cos/s.Pdf)
	})
}
