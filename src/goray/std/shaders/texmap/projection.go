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

package texmap

import (
	"math"
	"goray/vector"
)

// A Projector computes the projection of 3D space into a 2D texture coordinate.
type Projector interface {
	Project(p, n vector.Vector3D) vector.Vector3D
}

// A ProjectorFunc is a simple function-based projector.
type ProjectorFunc func(p, n vector.Vector3D) vector.Vector3D

func (f ProjectorFunc) Project(p, n vector.Vector3D) vector.Vector3D { return f(p, n) }

func flatMap(p, n vector.Vector3D) vector.Vector3D {
	return p
}

func tubeMap(p, n vector.Vector3D) (res vector.Vector3D) {
	res[vector.Y] = p[vector.Z]
	d := p[vector.X]*p[vector.X] + p[vector.Y]*p[vector.Y]
	if d > 0 {
		res[vector.Z] = 1 / math.Sqrt(d)
		res[vector.X] = -math.Atan2(p[vector.X], p[vector.Y]) / math.Pi
	}
	return
}

func sphereMap(p, n vector.Vector3D) (res vector.Vector3D) {
	d := p[vector.X]*p[vector.X] + p[vector.Y]*p[vector.Y] + p[vector.Z]*p[vector.Z]
	if d > 0 {
		res[vector.Z] = math.Sqrt(d)
		if p[vector.X] != 0 && p[vector.Y] != 0 {
			res[vector.X] = -math.Atan2(p[vector.X], p[vector.Y]) / math.Pi
		}
		res[vector.Y] = 1.0 - 2.0*math.Acos(p[vector.Z]/res[vector.Z])/math.Pi
	}
	return
}

func cubeMap(p, n vector.Vector3D) (res vector.Vector3D) {
	axis := vector.LargestAxis(math.Fabs(n[vector.X]), math.Fabs(n[vector.Y]), math.Fabs(n[vector.Z]))
	return vector.Vector3D{p[axis.Next()], p[axis.Prev()], p[axis]}
}

// Built-in projection schemes
var (
	FlatMap   Projector = ProjectorFunc(flatMap)
	TubeMap   Projector = ProjectorFunc(tubeMap)
	SphereMap Projector = ProjectorFunc(sphereMap)
	CubeMap   Projector = ProjectorFunc(cubeMap)
)
