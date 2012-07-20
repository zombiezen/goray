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
	"bitbucket.org/zombiezen/goray"
	yamldata "bitbucket.org/zombiezen/goray/yaml/data"
	"bitbucket.org/zombiezen/goray/yamlscene"
	"bitbucket.org/zombiezen/math3/vec64"
	"errors"
)

// orthographic is a simple orthographic camera.
type orthographic struct {
	resx, resy         int
	position           vec64.Vector
	vlook, vup, vright vec64.Vector
}

var _ goray.Camera = &orthographic{}

// NewOrthographic creates a new orthographic camera.
func NewOrthographic(pos, look, up vec64.Vector, resx, resy int, aspect, scale float64) goray.Camera {
	c := new(orthographic)
	c.resx, c.resy = resx, resy
	c.vup = vec64.Sub(up, pos)
	c.vlook = vec64.Sub(look, pos).Normalize()
	c.vright = vec64.Cross(c.vup, c.vlook)
	c.vup = vec64.Cross(c.vright, c.vlook)

	// Normalize separately
	c.vup = c.vup.Normalize()
	c.vright = c.vright.Normalize()

	c.vright = c.vright.Negate()
	c.vup = c.vup.Scale(aspect * float64(resy) / float64(resx))

	c.position = vec64.Sub(pos, vec64.Add(c.vup, c.vright).Scale(0.5*scale))

	c.vup = c.vup.Scale(scale / float64(resy))
	c.vright = c.vright.Scale(scale / float64(resx))
	return c
}

func (c *orthographic) SampleLens() bool {
	return false
}

func (c *orthographic) ResolutionX() int {
	return c.resx
}

func (c *orthographic) ResolutionY() int {
	return c.resy
}

func (c *orthographic) ShootRay(x, y, u, v float64) (r goray.Ray, wt float64) {
	wt = 1
	r = goray.Ray{
		From: vec64.Sum(c.position, c.vright.Scale(x), c.vup.Scale(y)),
		Dir:  c.vlook,
		TMax: -1.0,
	}
	return
}

func (c *orthographic) Project(wo goray.Ray, lu, lv *float64) (pdf float64, changed bool) {
	return 0.0, false
}

func init() {
	yamlscene.Constructor[yamlscene.StdPrefix+"cameras/ortho"] = yamlscene.MapConstruct(constructOrthographic)
}

func constructOrthographic(m yamldata.Map) (interface{}, error) {
	// Obtain position, look, and up
	pos, posOk := m["position"].(vec64.Vector)
	look, lookOk := m["look"].(vec64.Vector)
	up, upOk := m["up"].(vec64.Vector)
	if !posOk || !lookOk || !upOk {
		return nil, errors.New("Ortho nodes must have position, look, and up vectors")
	}

	// Width and height
	width, widthOk := yamldata.AsInt(m["width"])
	height, heightOk := yamldata.AsInt(m["height"])
	if !widthOk || !heightOk {
		return nil, errors.New("Ortho must have width and height")
	}

	// Aspect and scale
	aspect, ok := yamldata.AsFloat(m["aspect"])
	if !ok {
		aspect = 1.0
	}
	scale, ok := yamldata.AsFloat(m["scale"])
	if !ok {
		scale = 1.0
	}

	// Create camera (finally!)
	return NewOrthographic(pos, look, up, width, height, aspect, scale), nil
}
