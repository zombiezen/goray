/*
	Copyright (c) 2011 Ross Light.
	Copyright (c) 2005 Mathias Wein, Alejandro Conty, and Alfredo de Greef.

	This file is part of goray.

	goray is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	goray is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with goray.  If not, see <http://www.gnu.org/licenses/>.
*/

// Package vector provides a three-dimensional vector type and operations.
package vector

import (
	"fmt"
	"math"
)

// Vector3D holds a three-dimensional vector. The default value is a zero vector.
// The old Yafaray engine had a distinct type for a point and a normal, but we just represent everything as vectors.
type Vector3D [3]float64

// Normalize creates a new vector that is of unit length in the same direction as the vector.
func (v Vector3D) Normalize() Vector3D {
	vlen := v.Length()
	if vlen == 0 {
		return v
	}
	return v.Scale(1.0 / vlen)
}

// Length returns the magnitude of the vector.
func (v Vector3D) Length() float64 {
	return math.Sqrt(v.LengthSqr())
}

// LengthSqr returns the magnitude squared of the vector.  This is cheaper to compute than Length.
func (v Vector3D) LengthSqr() float64 {
	return v[X]*v[X] + v[Y]*v[Y] + v[Z]*v[Z]
}

// Abs returns a new vector with all positive components.
func (v Vector3D) Abs() Vector3D {
	return Vector3D{math.Abs(v[X]), math.Abs(v[Y]), math.Abs(v[Z])}
}

// Negate returns a new vector in the opposite direction.
func (v Vector3D) Negate() Vector3D {
	return Vector3D{-v[X], -v[Y], -v[Z]}
}

// Inverse returns a new vector that is the result of 1.0 / v[i] for all i.  Any zero value is left as zero.
func (v Vector3D) Inverse() (r Vector3D) {
	for axis, comp := range v {
		if comp != 0.0 {
			r[axis] = 1.0 / comp
		}
	}
	return
}

// IsZero indicates whether the vector is the zero vector.
func (v Vector3D) IsZero() bool {
	return v[X] == 0 && v[Y] == 0 && v[Z] == 0
}

func (v Vector3D) String() string {
	return fmt.Sprintf("<%.4f, %.4f, %.4f>", v[X], v[Y], v[Z])
}

func (v Vector3D) GoString() string {
	return fmt.Sprintf("vector.Vector3D{%#v, %#v, %#v}", v[X], v[Y], v[Z])
}

// Add computes the sum of two vectors.
func Add(v1, v2 Vector3D) Vector3D {
	return Vector3D{v1[X] + v2[X], v1[Y] + v2[Y], v1[Z] + v2[Z]}
}

// Sum computes the sum of two or more vectors.
func Sum(v1, v2 Vector3D, vn ...Vector3D) Vector3D {
	result := Vector3D{v1[X] + v2[X], v1[Y] + v2[Y], v1[Z] + v2[Z]}
	for _, u := range vn {
		result[X] += u[X]
		result[Y] += u[Y]
		result[Z] += u[Z]
	}
	return result
}

// AddScalar adds a scalar to all of a vector's components.
func (v Vector3D) AddScalar(s float64) Vector3D {
	return Vector3D{v[X] + s, v[Y] + s, v[Z] + s}
}

// Sub computes the difference of two vectors.
func Sub(v1, v2 Vector3D) Vector3D {
	return Vector3D{v1[X] - v2[X], v1[Y] - v2[Y], v1[Z] - v2[Z]}
}

// Scale multiplies all of a vector's components by a scalar.
func (v Vector3D) Scale(s float64) Vector3D {
	return Vector3D{v[X] * s, v[Y] * s, v[Z] * s}
}

// Mul multiplies the components of two vectors together.
func Mul(v1, v2 Vector3D) Vector3D {
	return Vector3D{v1[X] * v2[X], v1[Y] * v2[Y], v1[Z] * v2[Z]}
}

// Dot computes the dot product of two vectors.
func Dot(v1, v2 Vector3D) float64 {
	return v1[X]*v2[X] + v1[Y]*v2[Y] + v1[Z]*v2[Z]
}

// Cross computes the cross product of two vectors.
func Cross(v1, v2 Vector3D) Vector3D {
	return Vector3D{
		v1[Y]*v2[Z] - v1[Z]*v2[Y],
		v1[Z]*v2[X] - v1[X]*v2[Z],
		v1[X]*v2[Y] - v1[Y]*v2[X],
	}
}

// CreateCS finds two normalized vectors orthogonal to the given one that can be used as a coordinate system.
//
// This is particularly useful for UV-mapping and the like.
func CreateCS(normal Vector3D) (u, v Vector3D) {
	if normal[X] == 0 && normal[Y] == 0 {
		if normal[Z] < 0 {
			u = Vector3D{-1.0, 0.0, 0.0}
		} else {
			u = Vector3D{1.0, 0.0, 0.0}
		}
		v = Vector3D{0.0, 1.0, 0.0}
	} else {
		d := 1.0 / math.Sqrt(normal[Y]*normal[Y]+normal[X]*normal[X])
		u = Vector3D{normal[Y] * d, -normal[X] * d, 0.0}
		v = Cross(normal, u)
	}
	return
}

// Reflect calculates a reflection of a vector based on a normal.
func Reflect(v, n Vector3D) Vector3D {
	vn := Dot(v, n)
	if vn < 0 {
		return v.Negate()
	}
	return Sub(n.Scale(2*vn), v)
}