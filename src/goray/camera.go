//
//  goray/camera.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

package camera

import (
	"./goray/ray"
	"./goray/vector"
)

type Camera interface {
	ShootRay(x, y, u, v float) (ray.Ray, float)
	ResolutionX() int
	ResolutionY() int
	Project(wo ray.Ray, lu, lv float) (ok bool, pu, v float, pdf float)
    // SampleLens indicates whether the lens needs to be sampled using the u and v
    // parameters of ShootRay.  This is useful for DOF-like effects.  When this
    // returns false, no lens samples need to be computed.
	SampleLens() bool
}

type orthoCam struct {
	resx, resy         int
	position           vector.Vector3D
	vlook, vup, vright vector.Vector3D
}

func NewOrtho(pos, look, up vector.Vector3D, resx, resy int, aspect, scale float) Camera {
	c := new(orthoCam)
	c.resx, c.resy = resx, resy
	c.vup = vector.Sub(up, pos)
	c.vlook = vector.Sub(look, pos).Normalize()
	c.vright = vector.Cross(c.vup, c.vlook)
	c.vup = vector.Cross(c.vright, c.vlook).Normalize()

	c.vright = vector.ScalarMul(c.vright.Normalize(), -1.0)
	c.vup = vector.ScalarMul(c.vup, aspect*float(resy)/float(resx))

	c.position = vector.Sub(pos, vector.ScalarMul(vector.Add(c.vup, c.vright), 0.5*scale))

	c.vup = vector.ScalarMul(c.vup, scale/float(resy))
	c.vright = vector.ScalarMul(c.vup, scale/float(resx))
	return c
}

func (c *orthoCam) ShootRay(x, y, u, v float) (r ray.Ray, wt float) {
	wt = 1
	r.From = vector.Add(c.position, vector.ScalarMul(c.vright, x), vector.ScalarMul(c.vup, y))
	r.Dir = c.vlook
	return
}
func (c *orthoCam) SampleLens() bool { return false }
func (c *orthoCam) ResolutionX() int { return c.resx }
func (c *orthoCam) ResolutionY() int { return c.resy }
func (c *orthoCam) Project(wo ray.Ray, lu, lv float) (ok bool, u, v float, pdf float) {
	ok = false
	return
}
