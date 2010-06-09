//
//  goray/core/ray.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

/* The goray/ray package contains data structures for describing rays. */
package ray

import "fmt"
import "goray/core/vector"

type Ray interface {
	// From returns where the ray originated
	From() vector.Vector3D
	SetFrom(vector.Vector3D)
	// Dir returns the direction the ray is traveling
	Dir() vector.Vector3D
	SetDir(vector.Vector3D)
	// TMin returns the minimum distance from the ray's origin to consider
	TMin() float
	SetTMin(float)
	// TMax returns the maximum distance from the ray's origin to consider
	TMax() float
	SetTMax(float)
	// Time returns the relative time frame (values between [0, 1]) at which the ray was generated
	Time() float
	SetTime(float)
}

/* Ray defines the path of light. */
type simpleRay struct {
	from       vector.Vector3D
	dir        vector.Vector3D
	tMin, tMax float
	time       float
}

func New() Ray { return &simpleRay{tMax: -1.0} }
func Copy(r Ray) Ray {
	return &simpleRay{from: r.From(), dir: r.Dir(), tMin: r.TMin(), tMax: r.TMax(), time: r.Time()}
}

func (r *simpleRay) From() vector.Vector3D     { return r.from }
func (r *simpleRay) SetFrom(v vector.Vector3D) { r.from = v }
func (r *simpleRay) Dir() vector.Vector3D      { return r.dir }
func (r *simpleRay) SetDir(v vector.Vector3D)  { r.dir = v }
func (r *simpleRay) TMin() float               { return r.tMin }
func (r *simpleRay) SetTMin(t float)           { r.tMin = t }
func (r *simpleRay) TMax() float               { return r.tMax }
func (r *simpleRay) SetTMax(t float)           { r.tMax = t }
func (r *simpleRay) Time() float               { return r.time }
func (r *simpleRay) SetTime(t float)           { r.time = t }

func (r *simpleRay) String() string {
	return fmt.Sprintf("Ray{From: %v, Dir: %v, TMin: %.2f, TMax: %.2f, Time: %.2f}", r.from, r.dir, r.tMin, r.tMax, r.time)
}

/*
   DifferentialRay stores additional information about a ray for use in surface intersections.
   For an explanation, see http://www.opticalres.com/white%20papers/DifferentialRayTracing.pdf
*/
type DifferentialRay struct {
	*simpleRay
	FromX, FromY vector.Vector3D
	DirX, DirY   vector.Vector3D
}

func NewDifferential() *DifferentialRay {
    return &DifferentialRay{simpleRay: New().(*simpleRay)}
}

func (r *DifferentialRay) String() string {
	return fmt.Sprintf("DifferentialRay{%v, FromX: %v, FromY: %v, DirX: %v, DirY: %v}", r.simpleRay, r.FromX, r.FromY, r.DirX, r.DirY)
}
