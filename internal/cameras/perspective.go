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

package cameras

import (
	"errors"
	"math"

	"bitbucket.org/zombiezen/goray"
	"bitbucket.org/zombiezen/math3/vec64"
	yamldata "zombiezen.com/go/goray/internal/yaml/data"
	"zombiezen.com/go/goray/internal/yamlscene"
)

// A Bokeh determines the shape of out-of-focus light.
type Bokeh int

// Bokeh shapes
const (
	Disk1    Bokeh = 0
	Disk2    Bokeh = 1
	Triangle Bokeh = 3
	Square   Bokeh = 4
	Pentagon Bokeh = 5
	Hexagon  Bokeh = 6
	Ring     Bokeh = 7
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

// perspective is a conventional perspective camera.
type perspective struct {
	resx, resy    int
	eye           vec64.Vector // eye is the camera position
	focalDistance float64
	dofDistance   float64
	aspectRatio   float64 // aspectRatio is the aspect of the camera (not the image)

	look, up, right vec64.Vector
	dofUp, dofRight vec64.Vector
	x, y, z         vec64.Vector

	aperture  float64
	aPix      float64
	bokeh     Bokeh
	bokehBias BokehBias
	lens      []float64
}

var _ goray.Camera = &perspective{}

// NewPerspective creates a perspective camera.
// It will not lead you to enlightenment.
func NewPerspective(pos, look, up vec64.Vector,
	resx, resy int,
	aspect, focalDist, aperture float64,
	bokeh Bokeh, bias BokehBias, bokehRot float64) goray.Camera {
	cam := new(perspective)
	cam.eye = pos
	cam.aperture = aperture
	cam.dofDistance = 0
	cam.resx, cam.resy = resx, resy

	cam.up = vec64.Sub(up, pos)
	cam.look = vec64.Sub(look, pos)
	cam.right = vec64.Cross(cam.up, cam.look)
	cam.up = vec64.Cross(cam.right, cam.look)

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
	cam.look = vec64.Sub(cam.look.Scale(cam.focalDistance), vec64.Add(cam.up, cam.right).Scale(0.5))
	cam.up = cam.up.Scale(1.0 / float64(resy))
	cam.right = cam.right.Scale(1.0 / float64(resx))
	cam.aPix = cam.aspectRatio / (cam.focalDistance * cam.focalDistance)

	// Set up bokeh
	cam.bokeh = bokeh
	cam.bokehBias = bias

	if cam.bokeh >= Triangle && cam.bokeh <= Hexagon {
		w := bokehRot * math.Pi / 180
		wi := 2.0 * math.Pi / float64(cam.bokeh)
		cam.lens = make([]float64, 0, (int(cam.bokeh)+2)*2)
		for i := 0; i < int(cam.bokeh)+2; i++ {
			cam.lens = append(cam.lens, math.Cos(w), math.Sin(w))
			w += wi
		}
	}
	return cam
}

func (cam *perspective) ResolutionX() int {
	return cam.resx
}

func (cam *perspective) ResolutionY() int {
	return cam.resy
}

func (cam *perspective) sampleTSD(r1, r2 float64) (u, v float64) {
	idx := int(r1 * float64(cam.bokeh))
	r1 = biasDist(cam.bokehBias, (r1-float64(idx)/float64(cam.bokeh))*float64(cam.bokeh))
	b1 := r1 * r2
	b0 := r1 - b1
	idx <<= 1 // multiply by two

	u = cam.lens[idx]*b0 + cam.lens[idx+2]*b1
	v = cam.lens[idx+1]*b0 + cam.lens[idx+3]*b1
	return
}

func (cam *perspective) getLensUV(r1, r2 float64) (u, v float64) {
	switch cam.bokeh {
	case Triangle, Square, Pentagon, Hexagon:
		return cam.sampleTSD(r1, r2)
	case Disk2, Ring:
		w := 2 * math.Pi * r2
		if cam.bokeh == Ring {
			r1 = 1.0
		} else {
			r1 = biasDist(cam.bokehBias, r1)
		}
		u, v = r1*math.Cos(w), r2*math.Sin(w)
		return
	}

	return shirleyDisk(r1, r2)
}

func (cam *perspective) ShootRay(x, y, u, v float64) (r goray.Ray, wt float64) {
	wt = 1.0 // for now, always 1, except 0 for probe when outside sphere

	r = goray.Ray{
		From: cam.eye,
		Dir:  vec64.Sum(cam.right.Scale(x), cam.up.Scale(y), cam.look).Normalize(),
		TMax: -1.0,
	}

	if cam.SampleLens() {
		u, v = cam.getLensUV(u, v)
		li := vec64.Add(cam.dofRight.Scale(u), cam.dofUp.Scale(v))
		r.From = vec64.Add(r.From, li)
		r.Dir = vec64.Sub(r.Dir.Scale(cam.dofDistance), li).Normalize()
	}
	return
}

func (cam *perspective) Project(wo goray.Ray, lu, lv *float64) (pdf float64, changed bool) {
	// TODO
	return
}

func (cam *perspective) SampleLens() bool {
	return cam.aperture != 0
}

func init() {
	yamlscene.Constructor[yamlscene.StdPrefix+"cameras/perspective"] = yamlscene.MapConstruct(constructPerspective)
}

func constructPerspective(m yamldata.Map) (interface{}, error) {
	m = m.Copy()

	if !m.HasKeys("position", "look", "up", "width", "height") {
		return nil, errors.New("Missing required camera key")
	}

	pos := m["position"].(vec64.Vector)
	look := m["look"].(vec64.Vector)
	up := m["up"].(vec64.Vector)
	width, _ := yamldata.AsInt(m["width"])
	height, _ := yamldata.AsInt(m["height"])

	aspect, _ := yamldata.AsFloat(m.SetDefault("aspect", 1.0))
	focalDistance, _ := yamldata.AsFloat(m.SetDefault("focalDistance", 1.0))
	dofDistance, _ := yamldata.AsFloat(m.SetDefault("dofDistance", 0.0))
	aperture, _ := yamldata.AsFloat(m.SetDefault("aperture", 0.0))
	btype := m.SetDefault("bokehType", "disk1").(string)
	bbias, _ := m.SetDefault("bokehBias", "uniform").(string)
	bokehRot, _ := yamldata.AsFloat(m.SetDefault("bokehRotation", 1.0))

	bokehType := Disk1
	switch btype {
	case "disk2":
		bokehType = Disk2
	case "triangle":
		bokehType = Triangle
	case "square":
		bokehType = Square
	case "pentagon":
		bokehType = Pentagon
	case "hexagon":
		bokehType = Hexagon
	case "ring":
		bokehType = Ring
	}

	bokehBias := NoBias
	switch bbias {
	case "center":
		bokehBias = CenterBias
	case "edge":
		bokehBias = EdgeBias
	}

	cam := NewPerspective(pos, look, up, width, height, aspect, focalDistance, aperture, bokehType, bokehBias, bokehRot)
	cam.(*perspective).dofDistance = dofDistance
	return cam, nil
}
