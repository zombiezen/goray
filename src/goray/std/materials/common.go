//
//	goray/std/materials/common.go
//	goray
//
//	Created by Ross Light on 2011-02-04.
//

package common

import (
	"math"
	"goray/montecarlo"
	"goray/sampleutil"
	"goray/core/color"
	"goray/core/material"
	"goray/core/render"
	"goray/core/surface"
	"goray/core/vector"
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
	Sample(state *render.State, sp surface.Point, wo vector.Vector3D, s *material.Sample) (color.Color, vector.Vector3D)
	GetFlags() material.BSDF
}

func ScatterPhoton(mat Sampler, state *render.State, sp surface.Point, wi vector.Vector3D, s *material.PhotonSample) (wo vector.Vector3D, scattered bool) {
	scol, wo := mat.Sample(state, sp, wi, &s.Sample)
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

func GetReflectivity(mat Sampler, state *render.State, sp surface.Point, flags material.BSDF) (col color.Color) {
	const N = 16

	col = color.Black
	if flags&(material.BSDFTransmit|material.BSDFReflect)&mat.GetFlags() == 0 {
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
		s := material.Sample{S1: s3, S2: s4, Flags: flags}
		c, wi := mat.Sample(state, sp, wo, &s)
		if s.Pdf > 1e-6 {
			col = color.Add(col, color.ScalarMul(c, math.Fabs(vector.Dot(wi, sp.Normal))/s.Pdf))
		}
	}
	return color.ScalarMul(col, 1/N)
}
