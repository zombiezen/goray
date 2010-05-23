//
//  goray/vector.go
//  goray
//
//  Created by Ross Light on 2010-05-22.
//

package goray

import "math"

type Vector3D struct {
    X, Y, Z float
}

func (v Vector3D) Normalize() Vector3D {
    vlen := v.Length()
    if vlen == 0 {
        return v
    }
    return VectorScalarMul(v, 1.0 / vlen)
}

func (v Vector3D) Length() float {
    return float(math.Sqrt(float64(v.Length())))
}

func (v Vector3D) LengthSqr() float {
    return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vector3D) Abs() Vector3D {
    return Vector3D{float(math.Fabs(float64(v.X))), float(math.Fabs(float64(v.Y))), float(math.Fabs(float64(v.Z)))}
}

func (v Vector3D) IsZero() bool {
    return v.X == 0 && v.Y == 0 && v.Z == 0
}

func VectorAdd(v1, v2 Vector3D) Vector3D {
    return Vector3D{v1.X+v2.X, v1.Y+v2.Y, v1.Z+v2.Z}
}

func VectorSub(v1, v2 Vector3D) Vector3D {
    return Vector3D{v1.X-v2.X, v1.Y-v2.Y, v1.Z-v2.Z}
}

func VectorScalarMul(v Vector3D, s float) Vector3D {
    return Vector3D{v.X*s, v.Y*s, v.Z*s}
}

func VectorScalarDiv(v Vector3D, s float) Vector3D {
    return Vector3D{v.X/s, v.Y/s, v.Z/s}
}

func VectorDot(v1, v2 Vector3D) float {
    return v1.X*v2.X + v1.Y*v2.Y + v1.Z*v2.Z
}

type Point3D struct {
    X, Y, Z float
}
