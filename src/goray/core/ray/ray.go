//
//	goray/core/ray/ray.go
//	goray
//
//	Created by Ross Light on 2010-05-23.
//

// The ray package contains data structures for describing rays.
package ray

import "fmt"
import "goray/core/vector"

// Ray defines the path of light.
type Ray struct {
	From       vector.Vector3D
	Dir        vector.Vector3D
	TMin, TMax float
	Time       float
}

func (r Ray) String() string {
	return fmt.Sprintf("Ray{From: %v, Dir: %v, TMin: %.2f, TMax: %.2f, Time: %.2f}", r.From, r.Dir, r.TMin, r.TMax, r.Time)
}

// DifferentialRay stores additional information about a ray for use in surface intersections.
// For an explanation, see http://www.opticalres.com/white%20papers/DifferentialRayTracing.pdf
type DifferentialRay struct {
	Ray
	FromX, FromY vector.Vector3D
	DirX, DirY   vector.Vector3D
}

func (r DifferentialRay) String() string {
	return fmt.Sprintf("DifferentialRay{Ray: %v, FromX: %v, FromY: %v, DirX: %v, DirY: %v}", r.Ray, r.FromX, r.FromY, r.DirX, r.DirY)
}
