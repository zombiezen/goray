//
//  goray/camera.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

/*
   The goray/camera package defines a common interface for cameras, along with a
   simple orthographic camera.
*/
package camera

import (
	"goray/ray"
	"goray/vector"
)

/* A viewpoint of a scene */
type Camera interface {
	/*
	   ShootRay calculates the initial ray used for computing a fragment of the
	   output.  U and V are sample coordinates that are only calculated if
	   SampleLens returns true.
	*/
	ShootRay(x, y, u, v float) (ray.Ray, float)
	/* ResolutionX returns the number of fragments wide that the camera is. */
	ResolutionX() int
	/* ResolutionY returns the number of fragments high that the camera is. */
	ResolutionY() int
	/* Project calculates the projection of a ray onto the fragment plane. */
	Project(wo ray.Ray, lu, lv *float) (pdf float, changed bool)
	/*
	   SampleLens returns whether the lens needs to be sampled using the u and v
	   parameters of ShootRay.  This is useful for DOF-like effects.  When this
	   returns false, no lens samples need to be computed.
	*/
	SampleLens() bool
}

/* A simple orthographic camera */
type orthoCam struct {
	resx, resy         int
	position           vector.Vector3D
	vlook, vup, vright vector.Vector3D
}

/* NewOrtho creates a new orthographic camera */
func NewOrtho(pos, look, up vector.Vector3D, resx, resy int, aspect, scale float) Camera {
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

func (c *orthoCam) ShootRay(x, y, u, v float) (r ray.Ray, wt float) {
	wt = 1
	r = ray.New()
	r.SetFrom(vector.Add(c.position, vector.ScalarMul(c.vright, x), vector.ScalarMul(c.vup, y)))
	r.SetDir(c.vlook)
	return
}
func (c *orthoCam) SampleLens() bool { return false }
func (c *orthoCam) ResolutionX() int { return c.resx }
func (c *orthoCam) ResolutionY() int { return c.resy }
func (c *orthoCam) Project(wo ray.Ray, lu, lv *float) (pdf float, changed bool) {
	return 0.0, false
}
