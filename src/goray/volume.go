//
//  goray/volume.go
//  goray
//
//  Created by Ross Light on 2010-05-28.
//

package volume

import (
	"./goray/bound"
	"./goray/color"
	"./goray/light"
	"./goray/ray"
	"./goray/vector"
)

type Region interface {
	SigmaA(p, v vector.Vector3D) color.Color
	SigmaS(p, v vector.Vector3D) color.Color
	Emission(p, v vector.Vector3D) color.Color
	SigmaT(p, v vector.Vector3D) color.Color

	Attenuation(p vector.Vector3D, l *light.Light) float

	P(l, s vector.Vector3D) float

	Tau(r ray.Ray, step, offset float) color.Color

	Intersect(r ray.Ray) (ok bool, t0, t1 float)

	GetBoundingBox() *bound.Bound
}
