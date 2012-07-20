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
	"bitbucket.org/zombiezen/goray/vecutil"
	"bitbucket.org/zombiezen/math3/vec64"
	"math"
)

// A Projector computes the projection of 3D space into a 2D texture coordinate.
type Projector interface {
	Project(p, n vec64.Vector) vec64.Vector
}

// A ProjectorFunc is a simple function-based projector.
type ProjectorFunc func(p, n vec64.Vector) vec64.Vector

func (f ProjectorFunc) Project(p, n vec64.Vector) vec64.Vector {
	return f(p, n)
}

func flatMap(p, n vec64.Vector) vec64.Vector {
	return p
}

func tubeMap(p, n vec64.Vector) (res vec64.Vector) {
	res[vecutil.Y] = p[vecutil.Z]
	d := p[vecutil.X]*p[vecutil.X] + p[vecutil.Y]*p[vecutil.Y]
	if d > 0 {
		res[vecutil.Z] = 1 / math.Sqrt(d)
		res[vecutil.X] = -math.Atan2(p[vecutil.X], p[vecutil.Y]) / math.Pi
	}
	return
}

func sphereMap(p, n vec64.Vector) (res vec64.Vector) {
	d := p[vecutil.X]*p[vecutil.X] + p[vecutil.Y]*p[vecutil.Y] + p[vecutil.Z]*p[vecutil.Z]
	if d > 0 {
		res[vecutil.Z] = math.Sqrt(d)
		if p[vecutil.X] != 0 && p[vecutil.Y] != 0 {
			res[vecutil.X] = -math.Atan2(p[vecutil.X], p[vecutil.Y]) / math.Pi
		}
		res[vecutil.Y] = 1.0 - 2.0*math.Acos(p[vecutil.Z]/res[vecutil.Z])/math.Pi
	}
	return
}

func cubeMap(p, n vec64.Vector) (res vec64.Vector) {
	axis := vecutil.LargestAxis(math.Abs(n[vecutil.X]), math.Abs(n[vecutil.Y]), math.Abs(n[vecutil.Z]))
	return vec64.Vector{p[axis.Next()], p[axis.Prev()], p[axis]}
}

// Built-in projection schemes
var (
	FlatMap   Projector = ProjectorFunc(flatMap)
	TubeMap   Projector = ProjectorFunc(tubeMap)
	SphereMap Projector = ProjectorFunc(sphereMap)
	CubeMap   Projector = ProjectorFunc(cubeMap)
)
