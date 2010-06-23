//
//  goray/core/camera/camera.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

/* The camera package defines a common interface for cameras. */
package camera

import (
	"goray/core/ray"
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
