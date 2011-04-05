//
//	goray/std/shaders/texmap/projection.go
//	goray
//
//	Created by Ross Light on 2011-04-02.
//

package texmap

import (
	"math"
	"goray/core/vector"
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

var (
	FlatMap   Projector = ProjectorFunc(flatMap)
	TubeMap   Projector = ProjectorFunc(tubeMap)
	SphereMap Projector = ProjectorFunc(sphereMap)
	CubeMap   Projector = ProjectorFunc(cubeMap)
)
