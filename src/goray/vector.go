//
//  goray/vector.go
//  goray
//
//  Created by Ross Light on 2010-05-22.
//

package vector

import "fmt"
import "./fmath"

type Vector3D struct {
	X, Y, Z float
}

func New(x, y, z float) Vector3D {
	return Vector3D{x, y, z}
}

func (v Vector3D) Normalize() Vector3D {
	vlen := v.Length()
	if vlen == 0 {
		return v
	}
	return ScalarDiv(v, vlen)
}

func (v Vector3D) Length() float {
	return fmath.Sqrt(v.Length())
}

func (v Vector3D) LengthSqr() float {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vector3D) Abs() Vector3D {
	return Vector3D{fmath.Abs(v.X), fmath.Abs(v.Y), fmath.Abs(v.Z)}
}

func (v Vector3D) IsZero() bool {
	return v.X == 0 && v.Y == 0 && v.Z == 0
}

func (v Vector3D) GetAxis(axis int) float {
	switch axis {
	case 0:
		return v.X
	case 1:
		return v.Y
	case 2:
		return v.Z
	}
	return 0.0
}

func (v Vector3D) String() string {
	return fmt.Sprintf("<%.2f, %.2f, %.2f>", v.X, v.Y, v.Z)
}

func Add(v1, v2 Vector3D, vn ...Vector3D) Vector3D {
	result := Vector3D{v1.X + v2.X, v1.Y + v2.Y, v1.Z + v2.Z}
	for i := 0; i < len(vn); i++ {
		result.X += vn[i].X
		result.Y += vn[i].Y
		result.Z += vn[i].Z
	}
	return result
}

// ScalarAdd adds a scalar to all of a vector's components.
func ScalarAdd(v Vector3D, s float) Vector3D {
	return Vector3D{v.X + s, v.Y + s, v.Z + s}
}

func Sub(v1, v2 Vector3D, vn ...Vector3D) Vector3D {
	result := Vector3D{v1.X - v2.X, v1.Y - v2.Y, v1.Z - v2.Z}
	for i := 0; i < len(vn); i++ {
		result.X -= vn[i].X
		result.Y -= vn[i].Y
		result.Z -= vn[i].Z
	}
	return result
}

// ScalarSub subtracts a scalar from all of a vector's components.
func ScalarSub(v Vector3D, s float) Vector3D {
	return Vector3D{v.X - s, v.Y - s, v.Z - s}
}

func ScalarMul(v Vector3D, s float) Vector3D {
	return Vector3D{v.X * s, v.Y * s, v.Z * s}
}

func ScalarDiv(v Vector3D, s float) Vector3D {
	return Vector3D{v.X / s, v.Y / s, v.Z / s}
}

func Dot(v1, v2 Vector3D) float {
	return v1.X*v2.X + v1.Y*v2.Y + v1.Z*v2.Z
}

func Cross(v1, v2 Vector3D) Vector3D {
	return Vector3D{
		v1.Y*v2.Z - v1.Z*v2.Y,
		v1.Z*v2.X - v1.X*v2.Z,
		v1.X*v2.Z - v1.Y*v2.X,
	}
}

// CompMul multiplies the components of two vectors together.
func CompMul(v1, v2 Vector3D) Vector3D {
	return Vector3D{v1.X * v2.X, v1.Y * v2.Y, v1.Z * v2.Z}
}

// CompDiv divides the components of two vectors.
func CompDiv(v1, v2 Vector3D) Vector3D {
	return Vector3D{v1.X / v2.X, v1.Y / v2.Y, v1.Z / v2.Z}
}
