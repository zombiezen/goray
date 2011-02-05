//
//	goray/std/materials/common.go
//	goray
//
//	Created by Ross Light on 2011-02-04.
//

package common

import (
	"math"
	"goray/core/vector"
)

func Fresnel(i, n vector.Vector3D, ior float64) (kr, kt float64) {
	c := vector.Dot(i, n)
	if c < 0 {
		n = n.Negate()
		c = -c
	}
	g := ior*ior + c*c - 1

	if g <= 0 {
		g = 0
	} else {
		g = math.Sqrt(g)
	}

	aux := c * (g + c)
	kr = (0.5 * (g - c) * (g - c)) / ((g + c) * (g + c)) * (1 + (aux-1)*(aux-1)/((aux+1)*(aux+1)))
	if kr < 1.0 {
		kt = 1 - kr
	} else {
		kt = 0
	}
	return
}
