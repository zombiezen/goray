//
//  goray/primitive.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

type Primitive interface {
    GetBound() Bound
    IntersectsBound(Bound) bool
    HasClippingSupport() bool
    ClipToBound(bound [2][3]float, axis int, clipped Bound)
    Intersect(ray Ray, userdata interface{}) (hit bool, raydepth float)
    GetSurface() (SurfacePoint, Vector3D, userdata interface{})
    GetMaterial() *Material
}

type Sphere struct {
    center Vector3D
    radius float
    material *Material
}

func NewSphere(center Vector3D, radius float, material *Material) Sphere {
    return Sphere{center, radius, material}
}

func (s Sphere) GetBound() Bound {
    r := Vector3D{s.radius * 1.0001, s.radius * 1.0001, s.radius * 1.0001}
    return NewBound(VectorSub(s.center, r), VectorAdd(s.center, r))
}

func (s Sphere) Intersect(ray Ray, userdata interface{}) (bool, float) {
    vf := VectorSub(ray.from, s.center)
    ea := VectorDot
}
