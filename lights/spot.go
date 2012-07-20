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

package lights

import (
	"math"

	"bitbucket.org/zombiezen/goray"
	"bitbucket.org/zombiezen/goray/color"
	"bitbucket.org/zombiezen/goray/sampleutil"
	"bitbucket.org/zombiezen/goray/vecutil"
	yamldata "bitbucket.org/zombiezen/goray/yaml/data"
	"bitbucket.org/zombiezen/goray/yamlscene"
	"bitbucket.org/zombiezen/math3/vec64"
)

type spotLight struct {
	position         vec64.Vector
	direction        vec64.Vector
	du, dv           vec64.Vector
	cosStart, cosEnd float64
	icosDiff         float64
	intensity        float64
	color            color.Color
	pdf              sampleutil.Pdf1D
	interv1, interv2 float64
}

var _ goray.DiracLight = &spotLight{}

func NewSpot(from, to vec64.Vector, col color.Color, power, angle, falloff float64) goray.Light {
	newSpot := &spotLight{
		position:  from,
		direction: vec64.Sub(to, from).Normalize(),
		color:     color.ScalarMul(col, power),
		intensity: power,
	}
	newSpot.du, newSpot.dv = vecutil.CreateCS(newSpot.direction)
	radAngle := angle * math.Pi / 180
	radInnerAngle := radAngle * (1 - falloff)
	newSpot.cosStart = math.Cos(radInnerAngle)
	newSpot.cosEnd = math.Cos(radAngle)
	newSpot.icosDiff = 1.0 / (newSpot.cosStart - newSpot.cosEnd)
	f := make([]float64, 65)
	for i := range f {
		v := float64(i) / float64(len(f))
		f[i] = v * v * (3 - 2*v)
	}
	newSpot.pdf = sampleutil.NewPdf1D(f)

	/* the integral of the smoothstep is 0.5, and since it gets applied to the cos, and each delta cos
	corresponds to a constant surface are of the (partial) emitting sphere, we can actually simply
	compute the energy emitted from both areas, the constant and blending one...
	1  cosStart  cosEnd              -1
	|------|--------|-----------------|
	*/
	newSpot.interv1 = 1 - newSpot.cosStart
	newSpot.interv2 = 0.5 * (newSpot.cosStart - newSpot.cosEnd)
	sum := newSpot.interv1 + newSpot.interv2
	if sum > 1e-10 {
		newSpot.interv1 /= sum
		newSpot.interv2 /= sum
	}
	return newSpot
}

func (spot *spotLight) LightFlags() uint {
	return goray.LightTypeSingular
}

func (spot *spotLight) NumSamples() int {
	return 1
}

func (spot *spotLight) SetScene(scene *goray.Scene) {
}

func (spot *spotLight) TotalEnergy() color.Color {
	return color.ScalarMul(spot.color, 2*math.Pi*(1-0.5*(spot.cosStart+spot.cosEnd)))
}

func (spot *spotLight) Illuminate(sp goray.SurfacePoint, wi *goray.Ray) (col color.Color, ok bool) {
	ldir := vec64.Sub(spot.position, sp.Position)
	distSqr := ldir.LengthSqr()
	dist := math.Sqrt(distSqr)
	if dist == 0 {
		return
	}
	ldir = ldir.Scale(1.0 / dist) // normalize
	cosa := vec64.Dot(spot.direction.Negate(), ldir)
	switch {
	case cosa < spot.cosEnd:
		// Outside cone
		return
	case cosa >= spot.cosStart:
		// Not affected by falloff
		col = color.ScalarDiv(spot.color, distSqr)
	default:
		v := (cosa - spot.cosEnd) * spot.icosDiff
		v = v * v * (3 - 2*v)
		col = color.ScalarMul(spot.color, v/distSqr)
	}
	wi.TMax = dist
	wi.Dir = ldir
	ok = true
	return
}

func (spot *spotLight) IlluminateSample(sp goray.SurfacePoint, wi *goray.Ray, s *goray.LightSample) (ok bool) {
	s.Color, ok = spot.Illuminate(sp, wi)
	if ok {
		s.Flags = spot.LightFlags()
		s.Pdf = vec64.Sub(spot.position, sp.Position).LengthSqr()
	}
	return
}

func (spot *spotLight) IlluminatePdf(sp, spLight goray.SurfacePoint) float64 {
	return 0
}

func (spot *spotLight) emit(s1, s2, s3 float64) (col color.Color, wo vec64.Vector, pdf float64) {
	col = spot.color
	if s3 <= spot.interv1 {
		// Sample from cone not affected by falloff
		col = spot.color
		wo = sampleutil.Cone(spot.direction, spot.du, spot.dv, spot.cosStart, s1, s2)
		pdf = spot.interv1 / (2 * (1 - spot.cosStart))
		return
	}
	sm2, spdf := spot.pdf.Sample(s2)
	sm2 /= float64(spot.pdf.Len())
	pdf = (spot.interv2 * spdf) / (2.0 * (spot.cosStart - spot.cosEnd))
	cosAngle := spot.cosEnd + (spot.cosStart-spot.cosEnd)*sm2
	sinAngle := math.Sqrt(1 - cosAngle*cosAngle)
	t1 := 2 * math.Pi * s1
	wo = vec64.Add(vec64.Add(spot.du.Scale(math.Cos(t1)), spot.dv.Scale(math.Sin(t1))).Scale(sinAngle), spot.direction.Scale(cosAngle))
	col = color.ScalarMul(spot.color, spdf*spot.pdf.Integral) // color is scaled by falloff
	return
}

func (spot *spotLight) EmitPhoton(s1, s2, s3, s4 float64) (col color.Color, r goray.Ray, ipdf float64) {
	col, r.Dir, ipdf = spot.emit(s1, s2, s3)
	ipdf = math.Pi / ipdf
	r.From = spot.position
	return
}

func (spot *spotLight) EmitSample(s *goray.LightSample) (wo vec64.Vector, col color.Color) {
	col, wo, s.DirPdf = spot.emit(s.S1, s.S2, s.S3)
	s.Point.Position = spot.position
	s.AreaPdf = 1.0
	s.Flags = spot.LightFlags()
	return
}

func (spot *spotLight) EmitPdf(sp goray.SurfacePoint, wo vec64.Vector) (areaPdf, dirPdf, cosWo float64) {
	areaPdf, cosWo = 1, 1
	cosa := vec64.Dot(spot.direction, wo)
	switch {
	case cosa < spot.cosEnd:
		dirPdf = 0
	case cosa >= spot.cosStart:
		dirPdf = spot.interv1 / (2 * (1 - spot.cosStart))
	default:
		v := (cosa - spot.cosEnd) * spot.icosDiff
		v = v * v * (3 - 2*v)
		dirPdf = spot.interv2 * v / (spot.cosStart - spot.cosEnd)
	}
	return
}

func (spot *spotLight) CanIlluminate(pt vec64.Vector) bool {
	ldir := vec64.Sub(spot.position, pt)
	dist := ldir.Length()
	if dist == 0 {
		return false
	}
	ldir = ldir.Scale(1.0 / dist) // normalize
	cosa := vec64.Dot(spot.direction.Negate(), ldir)
	return cosa >= spot.cosEnd
}

func init() {
	yamlscene.Constructor[yamlscene.StdPrefix+"lights/spot"] = yamlscene.MapConstruct(constructSpot)
}

func constructSpot(m yamldata.Map) (interface{}, error) {
	pos := m["position"].(vec64.Vector)
	look := m["look"].(vec64.Vector)
	col := m["color"].(color.Color)
	power, _ := yamldata.AsFloat(m["intensity"])
	angle, _ := yamldata.AsFloat(m["coneAngle"])
	falloff, _ := yamldata.AsFloat(m["falloff"])
	return NewSpot(pos, look, col, power, angle, falloff), nil
}
