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

package materials

import (
	"math"

	"bitbucket.org/zombiezen/math3/vec64"
	"zombiezen.com/go/goray/internal/color"
	"zombiezen.com/go/goray/internal/goray"
	"zombiezen.com/go/goray/internal/montecarlo"
	"zombiezen.com/go/goray/internal/sampleutil"
)

func fresnel(i, n vec64.Vector, ior float64) (kr, kt float64) {
	c := vec64.Dot(i, n)
	if c < 0 {
		n = n.Negate()
		c = -c
	}
	g := ior*ior + c*c - 1

	if g <= 0 {
		g = 0
	} else {
		g = math.Sqrt(g)
	}

	aux := c * (g + c)
	kr = (0.5 * (g - c) * (g - c)) / ((g + c) * (g + c)) * (1 + (aux-1)*(aux-1)/((aux+1)*(aux+1)))
	if kr < 1.0 {
		kt = 1 - kr
	} else {
		kt = 0
	}
	return
}

type sampler interface {
	Sample(state *goray.RenderState, sp goray.SurfacePoint, wo vec64.Vector, s *goray.MaterialSample) (color.Color, vec64.Vector)
	MaterialFlags() goray.BSDF
}

func scatterPhoton(mat sampler, state *goray.RenderState, sp goray.SurfacePoint, wi vec64.Vector, s *goray.PhotonSample) (wo vec64.Vector, scattered bool) {
	scol, wo := mat.Sample(state, sp, wi, &s.MaterialSample)
	if s.Pdf <= 1e-6 {
		return
	}
	cnew := color.ScalarMul(color.Mul(s.LastColor, color.Mul(s.Alpha, scol)), math.Abs(vec64.Dot(wo, sp.Normal))/s.Pdf)
	newMax := math.Max(math.Max(cnew.Red(), cnew.Green()), cnew.Blue())
	oldMax := math.Max(math.Max(s.LastColor.Red(), s.LastColor.Green()), s.LastColor.Blue())
	prob := math.Min(1.0, newMax/oldMax)
	if s.S3 <= prob {
		scattered = true
		s.Color = color.ScalarMul(cnew, 1/prob)
	}
	return
}

func getReflectivity(mat sampler, state *goray.RenderState, sp goray.SurfacePoint, flags goray.BSDF) (col color.Color) {
	const N = 16

	col = color.Black
	if flags&(goray.BSDFTransmit|goray.BSDFReflect)&mat.MaterialFlags() == 0 {
		return
	}

	h1 := montecarlo.NewHalton(3)
	h2 := montecarlo.NewHalton(5)
	for i := 0; i < N; i++ {
		s1 := 1/(N*2) + 1/N*float64(i)
		s2 := montecarlo.VanDerCorput(uint32(i), 0)
		s3 := h1.Float64()
		s4 := h2.Float64()

		wo := sampleutil.CosHemisphere(sp.Normal, sp.NormalU, sp.NormalV, s1, s2)
		s := goray.MaterialSample{S1: s3, S2: s4, Flags: flags}
		c, wi := mat.Sample(state, sp, wo, &s)
		if s.Pdf > 1e-6 {
			col = color.Add(col, color.ScalarMul(c, math.Abs(vec64.Dot(wi, sp.Normal))/s.Pdf))
		}
	}
	return color.ScalarMul(col, 1/N)
}
