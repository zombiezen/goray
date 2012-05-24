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

// Package shader is a node-based shader system.
package shader

import (
	"bitbucket.org/zombiezen/goray/color"
)

type Result [4]float64

func (r Result) Scalar() float64 { return r[0] }

func (r Result) Color() color.AlphaColor {
	return color.RGBA{r[0], r[1], r[2], r[3]}
}

// Derivative calculates the (approximate) partial derivatives of df/dNU and
// df/dNV where f is the shader function, and NU/NV/N build the shading coordinate
// system.
func (r Result) Derivative() (du, dv float64) { return r[0], r[1] }

// Params convey non-result values to shaders.
type Params map[string]interface{}

// Nodes are elements of a node-based shading tree.  A shader associates a color
// or scalar with a surface point.
type Node interface {
	// Eval evalutes the node for a given surface point.
	Eval(inputs []Result, params Params) Result

	// EvalDerivative evaluates the node's partial derivatives for a given
	// surface point (e.g. for bump mapping).
	EvalDerivative(inputs []Result, params Params) Result

	// ViewDependent returns whether the shader value depends on wo and wi.
	ViewDependent() bool

	// Dependencies returns the nodes on which the output of this shader depends.
	Dependencies() []Node
}
