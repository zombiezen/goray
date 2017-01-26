package vecutil

import (
	"fmt"
	"math"

	"bitbucket.org/zombiezen/math3/vec64"
)

// An Axis is one of the coordinate axes.
type Axis int

// Axes
const (
	X Axis = iota
	Y
	Z
)

const NumAxes = 3

func (a Axis) String() string {
	switch a {
	case X:
		return "X"
	case Y:
		return "Y"
	case Z:
		return "Z"
	}
	return fmt.Sprintf("vecutil.Axis(%d)", int(a))
}

func (a Axis) GoString() string {
	switch a {
	case X:
		return "vecutil.X"
	case Y:
		return "vecutil.Y"
	case Z:
		return "vecutil.Z"
	}
	return fmt.Sprintf("vecutil.Axis(%d)", int(a))
}

// Next returns the next axis (and wraps after Z).
func (a Axis) Next() Axis {
	return (a + 1) % 3
}

// Prev returns the previous axis (and wraps before X).
func (a Axis) Prev() Axis {
	return (a + 2) % 3
}

// LargestAxis returns which parameter is the largest.
func LargestAxis(x, y, z float64) Axis {
	switch {
	case z > x && z > y:
		return Z
	case y > x && y > z:
		return Y
	}
	return X
}

// CreateCS finds two normalized vectors orthogonal to the given one that can be used as a coordinate system.
//
// This is particularly useful for UV-mapping and the like.
func CreateCS(normal vec64.Vector) (u, v vec64.Vector) {
	if normal[X] == 0 && normal[Y] == 0 {
		if normal[Z] < 0 {
			u = vec64.Vector{-1.0, 0.0, 0.0}
		} else {
			u = vec64.Vector{1.0, 0.0, 0.0}
		}
		v = vec64.Vector{0.0, 1.0, 0.0}
	} else {
		d := 1.0 / math.Sqrt(normal[Y]*normal[Y]+normal[X]*normal[X])
		u = vec64.Vector{normal[Y] * d, -normal[X] * d, 0.0}
		v = vec64.Cross(normal, u)
	}
	return
}

// Float3 converts a vector into a float64 triple.
func Float3(v vec64.Vector) [3]float64 {
	return [3]float64{v[0], v[1], v[2]}
}
