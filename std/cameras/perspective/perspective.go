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

package perspective

import (
	"bitbucket.org/zombiezen/goray"
	"bitbucket.org/zombiezen/goray/std/yamlscene"
	"bitbucket.org/zombiezen/goray/vector"
	yamldata "bitbucket.org/zombiezen/goray/yaml/data"
	"errors"
	"math"
)

type Bokeh int

const (
	Disk1Bokeh Bokeh = 0
	Disk2Bokeh       = 1
	RingBokeh        = 7
)

const (
	TriangleBokeh Bokeh = 3 + iota
	SquareBokeh
	PentagonBokeh
	HexagonBokeh
)

type BokehBias int

const (
	NoBias BokehBias = iota
	CenterBias
	EdgeBias
)

func biasDist(bias BokehBias, r float64) float64 {
	switch bias {
	case CenterBias:
		return math.Sqrt(math.Sqrt(r) * r)
	case EdgeBias:
		return math.Sqrt(1 - r*r)
	}
	return math.Sqrt(r)
}

// shirleyDisk maps a square to a disk using P. Shirley's concentric disk algorithm.
func shirleyDisk(r1, r2 float64) (u, v float64) {
	var phi, r float64
	a, b := 2*r1-1, 2*r2-1

	switch {
	case a > -b && a > b:
		r = a
		phi = math.Pi / 4 * (b / a)
	case a > -b && a <= b:
		r = b
		phi = math.Pi / 4 * (2 - a/b)
	case a <= -b && a < b:
		r = -a
		phi = math.Pi / 4 * (4 + b/a)
	default:
		r = -b
		if b != 0 {
			phi = math.Pi / 4 * (6 - a/b)
		} else {
			phi = 0
		}
	}

	return r * math.Cos(phi), r * math.Sin(phi)
}

type Camera struct {
	resx, resy    int
	eye           vector.Vector3D // eye is the camera position
	focalDistance float64
	dofDistance   float64
	aspectRatio   float64 // aspectRatio is the aspect of the camera (not the image)

	look, up, right vector.Vector3D
	dofUp, dofRight vector.Vector3D
	x, y, z         vector.Vector3D

	aperture  float64
	aPix      float64
	bokeh     Bokeh
	bokehBias BokehBias
	lens      []float64
}

var _ goray.Camera = &Camera{}

func New(pos, look, up vector.Vector3D, resx, resy int, aspect, focalDist, aperture float64, bokeh Bokeh, bias BokehBias, bokehRot float64) (cam *Camera) {
	cam = new(Camera)
	cam.eye = pos
	cam.aperture = aperture
	cam.dofDistance = 0
	cam.resx, cam.resy = resx, resy

	cam.up = vector.Sub(up, pos)
	cam.look = vector.Sub(look, pos)
	cam.right = vector.Cross(cam.up, cam.look)
	cam.up = vector.Cross(cam.right, cam.look)

	cam.up = cam.up.Normalize()
	cam.right = cam.right.Normalize()
	cam.right = cam.right.Negate() // Due to the order of vectors, we need to flip the vector to get "right"

	cam.look = cam.look.Normalize()
	cam.x = cam.right
	cam.y = cam.up
	cam.z = cam.look

	// For DOF, premultiply values with aperture
	cam.dofRight = cam.right.Scale(cam.aperture)
	cam.dofUp = cam.up.Scale(cam.aperture)

	cam.aspectRatio = aspect * float64(resy) / float64(resx)
	cam.up = cam.up.Scale(cam.aspectRatio)

	cam.focalDistance = focalDist
	cam.look = vector.Sub(cam.look.Scale(cam.focalDistance), vector.Add(cam.up, cam.right).Scale(0.5))
	cam.up = cam.up.Scale(1.0 / float64(resy))
	cam.right = cam.right.Scale(1.0 / float64(resx))
	cam.aPix = cam.aspectRatio / (cam.focalDistance * cam.focalDistance)

	// Set up bokeh
	cam.bokeh = bokeh
	cam.bokehBias = bias

	if cam.bokeh >= TriangleBokeh && cam.bokeh <= HexagonBokeh {
		w := bokehRot * math.Pi / 180
		wi := 2.0 * math.Pi / float64(cam.bokeh)
		cam.lens = make([]float64, 0, (int(cam.bokeh)+2)*2)
		for i := 0; i < int(cam.bokeh)+2; i++ {
			cam.lens = append(cam.lens, math.Cos(w), math.Sin(w))
			w += wi
		}
	}
	return
}

func (cam *Camera) ResolutionX() int { return cam.resx }
func (cam *Camera) ResolutionY() int { return cam.resy }

func (cam *Camera) sampleTSD(r1, r2 float64) (u, v float64) {
	idx := int(r1 * float64(cam.bokeh))
	r1 = biasDist(cam.bokehBias, (r1-float64(idx)/float64(cam.bokeh))*float64(cam.bokeh))
	b1 := r1 * r2
	b0 := r1 - b1
	idx <<= 1 // multiply by two

	u = cam.lens[idx]*b0 + cam.lens[idx+2]*b1
	v = cam.lens[idx+1]*b0 + cam.lens[idx+3]*b1
	return
}

func (cam *Camera) getLensUV(r1, r2 float64) (u, v float64) {
	switch cam.bokeh {
	case TriangleBokeh, SquareBokeh, PentagonBokeh, HexagonBokeh:
		return cam.sampleTSD(r1, r2)
	case Disk2Bokeh, RingBokeh:
		w := 2 * math.Pi * r2
		if cam.bokeh == RingBokeh {
			r1 = 1.0
		} else {
			r1 = biasDist(cam.bokehBias, r1)
		}
		u, v = r1*math.Cos(w), r2*math.Sin(w)
		return
	}

	return shirleyDisk(r1, r2)
}

func (cam *Camera) ShootRay(x, y, u, v float64) (r goray.Ray, wt float64) {
	wt = 1.0 // for now, always 1, except 0 for probe when outside sphere

	r = goray.Ray{
		From: cam.eye,
		Dir:  vector.Sum(cam.right.Scale(x), cam.up.Scale(y), cam.look).Normalize(),
		TMax: -1.0,
	}

	if cam.SampleLens() {
		u, v = cam.getLensUV(u, v)
		li := vector.Add(cam.dofRight.Scale(u), cam.dofUp.Scale(v))
		r.From = vector.Add(r.From, li)
		r.Dir = vector.Sub(r.Dir.Scale(cam.dofDistance), li).Normalize()
	}
	return
}

func (cam *Camera) Project(wo goray.Ray, lu, lv *float64) (pdf float64, changed bool) {
	// TODO
	return
}

func (cam *Camera) SampleLens() bool { return cam.aperture != 0 }

func init() {
	yamlscene.Constructor[yamlscene.StdPrefix+"cameras/perspective"] = yamlscene.MapConstruct(Construct)
}

func Construct(m yamldata.Map) (data interface{}, err error) {
	m = m.Copy()

	if !m.HasKeys("position", "look", "up", "width", "height") {
		err = errors.New("Missing required camera key")
		return
	}

	pos := m["position"].(vector.Vector3D)
	look := m["look"].(vector.Vector3D)
	up := m["up"].(vector.Vector3D)
	width, _ := yamldata.AsInt(m["width"])
	height, _ := yamldata.AsInt(m["height"])

	aspect, _ := yamldata.AsFloat(m.SetDefault("aspect", 1.0))
	focalDistance, _ := yamldata.AsFloat(m.SetDefault("focalDistance", 1.0))
	dofDistance, _ := yamldata.AsFloat(m.SetDefault("dofDistance", 0.0))
	aperture, _ := yamldata.AsFloat(m.SetDefault("aperture", 0.0))
	btype := m.SetDefault("bokehType", "disk1").(string)
	bbias, _ := m.SetDefault("bokehBias", "uniform").(string)
	bokehRot, _ := yamldata.AsFloat(m.SetDefault("bokehRotation", 1.0))

	bokehType := Disk1Bokeh
	switch btype {
	case "disk2":
		bokehType = Disk2Bokeh
	case "triangle":
		bokehType = TriangleBokeh
	case "square":
		bokehType = SquareBokeh
	case "pentagon":
		bokehType = PentagonBokeh
	case "hexagon":
		bokehType = HexagonBokeh
	case "ring":
		bokehType = RingBokeh
	}

	bokehBias := NoBias
	switch bbias {
	case "center":
		bokehBias = CenterBias
	case "edge":
		bokehBias = EdgeBias
	}

	cam := New(pos, look, up, width, height, aspect, focalDistance, aperture, bokehType, bokehBias, bokehRot)
	cam.dofDistance = dofDistance
	return cam, nil
}
