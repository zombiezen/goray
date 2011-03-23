//
//	goray/std/lights/point.go
//	goray
//
//	Created by Ross Light on 2010-06-02.
//

package point

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

type pointLight struct {
	position  vector.Vector3D
	color     color.Color
	intensity float64
}

func New(pos vector.Vector3D, col color.Color, intensity float64) light.Light {
	pl := pointLight{position: pos, color: color.ScalarMul(col, intensity)}
	pl.intensity = color.Energy(pl.color)
	return &pl
}

func (l *pointLight) NumSamples() int  { return 1 }
func (l *pointLight) LightFlags() uint { return light.TypeSingular }

func (l *pointLight) SetScene(scene interface{}) {}

func (l *pointLight) TotalEnergy() color.Color {
	return color.ScalarMul(l.color, 4*math.Pi)
}

func (l *pointLight) EmitPhoton(s1, s2, s3, s4 float64) (col color.Color, r ray.Ray, ipdf float64) {
	r = ray.Ray{
		From: l.position,
		Dir:  sampleutil.Sphere(s1, s2),
	}
	ipdf = 4.0 * math.Pi
	col = l.color
	return
}

func (l *pointLight) EmitSample(s *light.Sample) (wo vector.Vector3D, col color.Color) {
	s.Point.Position = l.position
	s.Flags = l.LightFlags()
	s.DirPdf, s.AreaPdf = 0.25, 1.0
	wo = sampleutil.Sphere(s.S1, s.S2)
	col = l.color
	return
}

func (l *pointLight) CanIlluminate(pt vector.Vector3D) bool { return true }

func (l *pointLight) IlluminateSample(sp surface.Point, wi ray.Ray, s *light.Sample) (wo ray.Ray, ok bool) {
	_, wo, ok = l.Illuminate(sp, wi)
	if ok {
		s.Flags = l.LightFlags()
		s.Color = l.color
		s.Pdf = vector.Sub(l.position, sp.Position).LengthSqr()
	}
	return
}

func (l *pointLight) Illuminate(sp surface.Point, wi ray.Ray) (col color.Color, wo ray.Ray, ok bool) {
	ldir := vector.Sub(l.position, sp.Position)
	distSqr := ldir.LengthSqr()
	dist := math.Sqrt(distSqr)
	if dist == 0.0 {
		return
	}

	ok = true
	idistSqr := 1.0 / distSqr
	ldir = vector.ScalarMul(ldir, 1.0/dist)

	wo = wi
	wo.TMax = dist
	wo.Dir = ldir

	col = color.ScalarMul(l.color, idistSqr)
	return
}

func (l *pointLight) IlluminatePdf(sp, spLight surface.Point) float64 { return 0.0 }

func (l *pointLight) EmitPdf(sp surface.Point, wo vector.Vector3D) (areaPdf, dirPdf, cosWo float64) {
	return 1.0, 0.25, 1.0
}

func Construct(m yamldata.Map) (data interface{}, err os.Error) {
	pos := m["position"].(vector.Vector3D)
	col := m["color"].(color.Color)
	intensity, _ := yamldata.AsFloat(m["intensity"])
	data = New(pos, col, intensity)
	return
}
