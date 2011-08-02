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

package common

import (
	"math"

	"goray"
	"goray/color"
	"goray/montecarlo"
	"goray/sampleutil"
	"goray/vector"
)

func Fresnel(i, n vector.Vector3D, ior float64) (kr, kt float64) {
	c := vector.Dot(i, n)
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

type Sampler interface {
	Sample(state *goray.RenderState, sp goray.SurfacePoint, wo vector.Vector3D, s *goray.MaterialSample) (color.Color, vector.Vector3D)
	MaterialFlags() goray.BSDF
}

func ScatterPhoton(mat Sampler, state *goray.RenderState, sp goray.SurfacePoint, wi vector.Vector3D, s *goray.PhotonSample) (wo vector.Vector3D, scattered bool) {
	scol, wo := mat.Sample(state, sp, wi, &s.MaterialSample)
	if s.Pdf <= 1e-6 {
		return
	}
	cnew := color.ScalarMul(color.Mul(s.LastColor, color.Mul(s.Alpha, scol)), math.Fabs(vector.Dot(wo, sp.Normal))/s.Pdf)
	newMax := math.Fmax(math.Fmax(cnew.Red(), cnew.Green()), cnew.Blue())
	oldMax := math.Fmax(math.Fmax(s.LastColor.Red(), s.LastColor.Green()), s.LastColor.Blue())
	prob := math.Fmin(1.0, newMax/oldMax)
	if s.S3 <= prob {
		scattered = true
		s.Color = color.ScalarMul(cnew, 1/prob)
	}
	return
}

func GetReflectivity(mat Sampler, state *goray.RenderState, sp goray.SurfacePoint, flags goray.BSDF) (col color.Color) {
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
			col = color.Add(col, color.ScalarMul(c, math.Fabs(vector.Dot(wi, sp.Normal))/s.Pdf))
		}
	}
	return color.ScalarMul(col, 1/N)
}
