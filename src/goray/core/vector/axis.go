//
//	goray/core/vector/axis.go
//	goray
//
//	Created by Ross Light on 2010-12-30.
//

package vector

import (
	"fmt"
)

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
	return fmt.Sprintf("vector.Axis(%d)", int(a))
}

func (a Axis) GoString() string {
	switch a {
	case X:
		return "vector.X"
	case Y:
		return "vector.Y"
	case Z:
		return "vector.Z"
	}
	return fmt.Sprintf("vector.Axis(%d)", int(a))
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
