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

import (
	"fmt"
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
