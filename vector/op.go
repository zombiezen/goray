// +build !amd64

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

package vector

// Add computes the sum of two vectors.
func Add(v1, v2 Vector3D) Vector3D {
	return Vector3D{v1[X] + v2[X], v1[Y] + v2[Y], v1[Z] + v2[Z]}
}

// Sub computes the difference of two vectors.
func Sub(v1, v2 Vector3D) Vector3D {
	return Vector3D{v1[X] - v2[X], v1[Y] - v2[Y], v1[Z] - v2[Z]}
}

// Dot computes the dot product of two vectors.
func Dot(v1, v2 Vector3D) float64 {
	return v1[X]*v2[X] + v1[Y]*v2[Y] + v1[Z]*v2[Z]
}
