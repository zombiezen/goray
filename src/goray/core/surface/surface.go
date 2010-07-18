//
//	goray/core/surface/surface.go
//	goray
//
//	Created by Ross Light on 2010-05-29.
//

// The surface package provides structures for representing the surface of an object.
package surface

import (
	"goray/core/ray"
	"goray/core/vector"
)

// Point represents a single point on an object's surface.
type Point struct {
	// The associated material, light, 3D object, and primitive.
	// You will have to convert these to the appropriate type, due to dependency
	// issues.  Admittedly, this is an ugly hack, but it fixes the problem.
	Material, Light, Object, Primitive interface{}

	OriginX, OriginY                    int // Only used with "win" texture coordinate mode
	Normal, GeometricNormal, OrcoNormal vector.Vector3D
	Position, OrcoPosition              vector.Vector3D

	HasUV, HasOrco, Available bool
	PrimitiveNumber           int

	U, V               float           // The texture coordinates
	NormalU, NormalV   vector.Vector3D // Vectors building orthogonal shading space with normal
	WorldU, WorldV     vector.Vector3D // U and V axes in world space
	ShadingU, ShadingV vector.Vector3D // U and V axes in shading space
	SurfaceU, SurfaceV float           // Raw surface parametric coordinates; required to evaluate Vmaps
}

// Differentials computes and stores data for surface intersections for differential rays.
// For more information, see http://www.opticalres.com/white%20papers/DifferentialRayTracing.pdf
type Differentials struct {
	X, Y  vector.Vector3D
	Point Point
}

// NewDifferentials creates a new Differentials struct.
func NewDifferentials(p Point, r *ray.DifferentialRay) Differentials {
	d := -vector.Dot(p.Normal, p.Position)
	tx := -(vector.Dot(p.Normal, r.FromX) + d) / vector.Dot(p.Normal, r.DirX)
	px := vector.CompMul(vector.ScalarAdd(r.FromX, tx), r.DirX)
	ty := -(vector.Dot(p.Normal, r.FromY) + d) / vector.Dot(p.Normal, r.DirY)
	py := vector.CompMul(vector.ScalarAdd(r.FromY, ty), r.DirY)
	return Differentials{
		X:     vector.Sub(px, p.Position),
		Y:     vector.Sub(py, p.Position),
		Point: p,
	}
}

// ReflectRay computes differentials for a scattered ray.
// For an explanation, see: http://en.wikipedia.org/wiki/Specular_reflection
func (d Differentials) ReflectRay(in, out *ray.DifferentialRay) {
	// Compute ray differential for specular reflection
	out.FromX = vector.Add(d.Point.Position, d.X)
	out.FromY = vector.Add(d.Point.Position, d.Y)
	// Compute differential reflected directions
	incidenceX, incidenceY := vector.Sub(in.Dir(), in.DirX), vector.Sub(in.Dir(), in.DirY)
	normDx, normDy := vector.Dot(incidenceX, d.Point.Normal), vector.Dot(incidenceY, d.Point.Normal)
	out.DirX = vector.Add(out.Dir(), vector.ScalarMul(incidenceX, -1), vector.ScalarMul(d.Point.Normal, 2*normDx))
	out.DirY = vector.Add(out.Dir(), vector.ScalarMul(incidenceY, -1), vector.ScalarMul(d.Point.Normal, 2*normDy))
}

// RefractRay computes differentials for a scattered ray.
// For an explanation, see: http://en.wikipedia.org/wiki/Snell's_law#Vector_form
func (d Differentials) RefractRay(in, out *ray.DifferentialRay, ior float) {
	out.FromX = vector.Add(d.Point.Position, d.X)
	out.FromY = vector.Add(d.Point.Position, d.Y)

	incidenceX, incidenceY := vector.Sub(in.Dir(), in.DirX), vector.Sub(in.Dir(), in.DirY)
	normDx, normDy := vector.Dot(incidenceX, d.Point.Normal), vector.Dot(incidenceY, d.Point.Normal)

	muDeriv := ior - (ior*ior*vector.Dot(in.Dir(), d.Point.Normal))/vector.Dot(out.Dir(), d.Point.Normal)
	muDx := muDeriv * normDx
	muDy := muDeriv * normDy

	out.DirX = vector.Add(out.Dir(), vector.ScalarMul(incidenceX, ior), vector.ScalarMul(d.Point.Normal, -muDx))
	out.DirY = vector.Add(out.Dir(), vector.ScalarMul(incidenceY, ior), vector.ScalarMul(d.Point.Normal, -muDy))
}

func (d Differentials) ProjectedPixelArea() float {
	return vector.Cross(d.X, d.Y).Length()
}
