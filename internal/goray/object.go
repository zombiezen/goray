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

package goray

import (
	"bitbucket.org/zombiezen/math3/vec64"
)

// Object3D is a collection of primitives.
type Object3D interface {
	// Primitives returns all of the primitives for this object.
	Primitives() []Primitive

	// Visible indicates whether the object is shown in the scene.
	Visible() bool
}

// Samplable defines an interface for sampling a surface.
type Samplable interface {
	// EnableSampling tries to enable sampling (may require additional memory and preprocessing time).
	EnableSampling() bool

	// Sample takes a sample of the object's surface.
	Sample(s1, s2 float64) (p, n vec64.Vector)
}

// SamplableObject3D is the set of three-dimensional objects that can sample their surfaces.
type SamplableObject3D interface {
	Object3D
	Samplable
}

// PrimitiveObject is a wrapper type that allows a single primitive to act as an object.
type PrimitiveObject struct {
	Primitive Primitive
}

func (o PrimitiveObject) Primitives() []Primitive {
	return []Primitive{o.Primitive}
}

func (o PrimitiveObject) Visible() bool { return true }
