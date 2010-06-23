//
//  goray/std/cameras/ortho.go
//  goray
//
//  Created by Ross Light on 2010-06-22.
//

/* The ortho package provides an orthographic camera. */
package ortho

import (
	"goray/core/camera"
	"goray/core/ray"
	"goray/core/vector"
)

/* A simple orthographic camera */
type orthoCam struct {
	resx, resy         int
	position           vector.Vector3D
	vlook, vup, vright vector.Vector3D
}

/* NewOrtho creates a new orthographic camera */
func New(pos, look, up vector.Vector3D, resx, resy int, aspect, scale float) camera.Camera {
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
	c.vup = vector.ScalarMul(c.vup, aspect*float(resy)/float(resx))

	c.position = vector.Sub(pos, vector.ScalarMul(vector.Add(c.vup, c.vright), 0.5*scale))

	c.vup = vector.ScalarMul(c.vup, scale/float(resy))
	c.vright = vector.ScalarMul(c.vright, scale/float(resx))
	return c
}

func (c *orthoCam) SampleLens() bool { return false }
func (c *orthoCam) ResolutionX() int { return c.resx }
func (c *orthoCam) ResolutionY() int { return c.resy }

func (c *orthoCam) ShootRay(x, y, u, v float) (r ray.Ray, wt float) {
	wt = 1
	r = ray.New()
	r.SetFrom(vector.Add(c.position, vector.ScalarMul(c.vright, x), vector.ScalarMul(c.vup, y)))
	r.SetDir(c.vlook)
	return
}

func (c *orthoCam) Project(wo ray.Ray, lu, lv *float) (pdf float, changed bool) {
	return 0.0, false
}
