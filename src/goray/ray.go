//
//  goray/ray.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

/* The goray/ray package contains data structures for describing rays. */
package ray

import "./goray/vector"

/* Ray defines the path of light. */
type Ray struct {
	From       vector.Vector3D // From stores where the ray originated
	Dir        vector.Vector3D // Dir stores the direction the ray is traveling
	TMin, TMax float           // TMin and TMax are used to indicate how much of the ray to consider for collision detection
	Time       float           // Time stores the relative time frame (values between [0, 1]) at which the ray was generated
}

/*
   DifferentialRay stores additional information about a ray for use in surface intersections.
   For an explanation, see http://www.opticalres.com/white%20papers/DifferentialRayTracing.pdf
*/
type DifferentialRay struct {
	Ray
	FromX, FromY vector.Vector3D
	DirX, DirY   vector.Vector3D
}
