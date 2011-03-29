//
//	goray/std/lights/spot.go
//	goray
//
//	Created by Ross Light on 2011-03-23.
//

package spot

import (
	"math"
	"os"

	"goray/core/color"
	"goray/core/light"
	"goray/core/ray"
	"goray/core/surface"
	"goray/core/vector"
	"goray/sampleutil"

	yamldata "goyaml.googlecode.com/hg/data"
)

type spotLight struct {
	position         vector.Vector3D
	direction        vector.Vector3D
	du, dv           vector.Vector3D
	cosStart, cosEnd float64
	icosDiff         float64
	intensity        float64
	color            color.Color
	pdf              sampleutil.Pdf1D
	interv1, interv2 float64
}

var _ light.DiracLight = &spotLight{}

func New(from, to vector.Vector3D, col color.Color, power, angle, falloff float64) light.Light {
	newSpot := &spotLight{
		position:  from,
		direction: vector.Sub(to, from).Normalize(),
		color:     color.ScalarMul(col, power),
		intensity: power,
	}
	newSpot.du, newSpot.dv = vector.CreateCS(newSpot.direction)
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

func (spot *spotLight) LightFlags() uint { return light.TypeSingular }
func (spot *spotLight) NumSamples() int  { return 1 }

func (spot *spotLight) SetScene(scene interface{}) {}

func (spot *spotLight) TotalEnergy() color.Color {
	return color.ScalarMul(spot.color, 2*math.Pi*(1-0.5*(spot.cosStart+spot.cosEnd)))
}

func (spot *spotLight) Illuminate(sp surface.Point, wi *ray.Ray) (col color.Color, ok bool) {
	ldir := vector.Sub(spot.position, sp.Position)
	distSqr := ldir.LengthSqr()
	dist := math.Sqrt(distSqr)
	if dist == 0 {
		return
	}
	ldir = vector.ScalarDiv(ldir, dist) // normalize
	cosa := vector.Dot(spot.direction.Negate(), ldir)
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

func (spot *spotLight) IlluminateSample(sp surface.Point, wi *ray.Ray, s *light.Sample) (ok bool) {
	s.Color, ok = spot.Illuminate(sp, wi)
	if ok {
		s.Flags = spot.LightFlags()
		s.Pdf = vector.Sub(spot.position, sp.Position).LengthSqr()
	}
	return
}

func (spot *spotLight) IlluminatePdf(sp, spLight surface.Point) float64 {
	return 0
}

func (spot *spotLight) emit(s1, s2, s3 float64) (col color.Color, wo vector.Vector3D, pdf float64) {
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
	wo = vector.Add(vector.ScalarMul(vector.Add(vector.ScalarMul(spot.du, math.Cos(t1)), vector.ScalarMul(spot.dv, math.Sin(t1))), sinAngle), vector.ScalarMul(spot.direction, cosAngle))
	col = color.ScalarMul(spot.color, spdf*spot.pdf.Integral) // color is scaled by falloff
	return
}

func (spot *spotLight) EmitPhoton(s1, s2, s3, s4 float64) (col color.Color, r ray.Ray, ipdf float64) {
	col, r.Dir, ipdf = spot.emit(s1, s2, s3)
	ipdf = math.Pi / ipdf
	r.From = spot.position
	return
}

func (spot *spotLight) EmitSample(s *light.Sample) (wo vector.Vector3D, col color.Color) {
	col, wo, s.DirPdf = spot.emit(s.S1, s.S2, s.S3)
	s.Point.Position = spot.position
	s.AreaPdf = 1.0
	s.Flags = spot.LightFlags()
	return
}

func (spot *spotLight) EmitPdf(sp surface.Point, wo vector.Vector3D) (areaPdf, dirPdf, cosWo float64) {
	areaPdf, cosWo = 1, 1
	cosa := vector.Dot(spot.direction, wo)
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

func (spot *spotLight) CanIlluminate(pt vector.Vector3D) bool {
	ldir := vector.Sub(spot.position, pt)
	dist := ldir.Length()
	if dist == 0 {
		return false
	}
	ldir = vector.ScalarDiv(ldir, dist) // normalize
	cosa := vector.Dot(spot.direction.Negate(), ldir)
	return cosa >= spot.cosEnd
}

func Construct(m yamldata.Map) (data interface{}, err os.Error) {
	pos := m["position"].(vector.Vector3D)
	look := m["look"].(vector.Vector3D)
	col := m["color"].(color.Color)
	power, _ := yamldata.AsFloat(m["intensity"])
	angle, _ := yamldata.AsFloat(m["coneAngle"])
	falloff, _ := yamldata.AsFloat(m["falloff"])
	data = New(pos, look, col, power, angle, falloff)
	return
}
