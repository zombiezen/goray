//
//	goray/core/shader/shader.go
//	goray
//
//	Created by Ross Light on 2010-07-14.
//

package shader

import (
	"goray/core/color"
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

// Nodes are elements of a node-based shading tree.  A shader associates a color
// or scalar with a surface point.
type Node interface {
	// Eval evalutes the node for a given surface point.
	Eval(params map[string]interface{}, inputs []Result) Result
	// EvalDerivative evaluates the node's partial derivatives for a given
	// surface point (e.g. for bump mapping).
	EvalDerivative(params map[string]interface{}, inputs []Result) Result
	// ViewDependent returns whether the shader value depends on wo and wi.
	ViewDependent() bool
	// Dependencies returns the nodes on which the output of this shader depends.
	Dependencies() []Node
}
