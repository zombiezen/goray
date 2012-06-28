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
	"bitbucket.org/zombiezen/goray/vector"
	"bitbucket.org/zombiezen/goray/yamlscene"

	yamldata "bitbucket.org/zombiezen/goray/yaml/data"
)

type pointLight struct {
	position  vector.Vector3D
	color     color.Color
	intensity float64
}

var _ goray.DiracLight = &pointLight{}

func NewPoint(pos vector.Vector3D, col color.Color, intensity float64) goray.Light {
	pl := pointLight{position: pos, color: color.ScalarMul(col, intensity)}
	pl.intensity = color.Energy(pl.color)
	return &pl
}

func (l *pointLight) NumSamples() int {
	return 1
}

func (l *pointLight) LightFlags() uint {
	return goray.LightTypeSingular
}

func (l *pointLight) SetScene(scene *goray.Scene) {
}

func (l *pointLight) TotalEnergy() color.Color {
	return color.ScalarMul(l.color, 4*math.Pi)
}

func (l *pointLight) EmitPhoton(s1, s2, s3, s4 float64) (col color.Color, r goray.Ray, ipdf float64) {
	r = goray.Ray{
		From: l.position,
		Dir:  sampleutil.Sphere(s1, s2),
	}
	ipdf = 4.0 * math.Pi
	col = l.color
	return
}

func (l *pointLight) EmitSample(s *goray.LightSample) (wo vector.Vector3D, col color.Color) {
	s.Point.Position = l.position
	s.Flags = l.LightFlags()
	s.DirPdf, s.AreaPdf = 0.25, 1.0
	wo = sampleutil.Sphere(s.S1, s.S2)
	col = l.color
	return
}

func (l *pointLight) CanIlluminate(pt vector.Vector3D) bool {
	return true
}

func (l *pointLight) IlluminateSample(sp goray.SurfacePoint, wi *goray.Ray, s *goray.LightSample) (ok bool) {
	_, ok = l.Illuminate(sp, wi)
	if ok {
		s.Flags = l.LightFlags()
		s.Color = l.color
		s.Pdf = vector.Sub(l.position, sp.Position).LengthSqr()
	}
	return
}

func (l *pointLight) Illuminate(sp goray.SurfacePoint, wi *goray.Ray) (col color.Color, ok bool) {
	ldir := vector.Sub(l.position, sp.Position)
	distSqr := ldir.LengthSqr()
	dist := math.Sqrt(distSqr)
	if dist == 0.0 {
		return
	}

	ok = true
	idistSqr := 1.0 / distSqr
	ldir = ldir.Scale(1.0 / dist)

	wi.TMax = dist
	wi.Dir = ldir

	col = color.ScalarMul(l.color, idistSqr)
	return
}

func (l *pointLight) IlluminatePdf(sp, spLight goray.SurfacePoint) float64 {
	return 0.0
}

func (l *pointLight) EmitPdf(sp goray.SurfacePoint, wo vector.Vector3D) (areaPdf, dirPdf, cosWo float64) {
	return 1.0, 0.25, 1.0
}

func init() {
	yamlscene.Constructor[yamlscene.StdPrefix+"lights/point"] = yamlscene.MapConstruct(constructPoint)
}

func constructPoint(m yamldata.Map) (interface{}, error) {
	pos := m["position"].(vector.Vector3D)
	col := m["color"].(color.Color)
	intensity, _ := yamldata.AsFloat(m["intensity"])
	return NewPoint(pos, col, intensity), nil
}
