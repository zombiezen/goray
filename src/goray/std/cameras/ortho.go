//
//	goray/std/cameras/ortho.go
//	goray
//
//	Created by Ross Light on 2010-06-22.
//

// The ortho package provides an orthographic camera.
package ortho

import (
	"os"
	"goray/core/camera"
	"goray/core/ray"
	"goray/core/vector"
	yamldata "goyaml.googlecode.com/hg/data"
)

// orthoCam is a simple orthographic camera.
type orthoCam struct {
	resx, resy         int
	position           vector.Vector3D
	vlook, vup, vright vector.Vector3D
}

// NewOrtho creates a new orthographic camera.
func New(pos, look, up vector.Vector3D, resx, resy int, aspect, scale float64) camera.Camera {
	c := new(orthoCam)
	c.resx, c.resy = resx, resy
	c.vup = vector.Sub(up, pos)
	c.vlook = vector.Sub(look, pos).Normalize()
	c.vright = vector.Cross(c.vup, c.vlook)
	c.vup = vector.Cross(c.vright, c.vlook)

	// Normalize separately
	c.vup = c.vup.Normalize()
	c.vright = c.vright.Normalize()

	c.vright = vector.ScalarMul(c.vright, -1.0)
	c.vup = vector.ScalarMul(c.vup, aspect*float64(resy)/float64(resx))

	c.position = vector.Sub(pos, vector.ScalarMul(vector.Add(c.vup, c.vright), 0.5*scale))

	c.vup = vector.ScalarMul(c.vup, scale/float64(resy))
	c.vright = vector.ScalarMul(c.vright, scale/float64(resx))
	return c
}

func (c *orthoCam) SampleLens() bool { return false }
func (c *orthoCam) ResolutionX() int { return c.resx }
func (c *orthoCam) ResolutionY() int { return c.resy }

func (c *orthoCam) ShootRay(x, y, u, v float64) (r ray.Ray, wt float64) {
	wt = 1
	r = ray.Ray{
		From: vector.Add(c.position, vector.ScalarMul(c.vright, x), vector.ScalarMul(c.vup, y)),
		Dir:  c.vlook,
		TMax: -1.0,
	}
	return
}

func (c *orthoCam) Project(wo ray.Ray, lu, lv *float64) (pdf float64, changed bool) {
	return 0.0, false
}

func Construct(m yamldata.Map) (data interface{}, err os.Error) {
	// Obtain position, look, and up
	pos, posOk := m["position"].(vector.Vector3D)
	look, lookOk := m["look"].(vector.Vector3D)
	up, upOk := m["up"].(vector.Vector3D)
	if !posOk || !lookOk || !upOk {
		err = os.NewError("Ortho nodes must have position, look, and up vectors")
		return
	}
	// Width and height
	width, widthOk := yamldata.AsInt(m["width"])
	height, heightOk := yamldata.AsInt(m["height"])
	if !widthOk || !heightOk {
		err = os.NewError("Ortho must have width and height")
		return
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
	data = New(pos, look, up, int(width), int(height), aspect, scale)
	return
}
