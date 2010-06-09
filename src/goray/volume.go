//
//  goray/volume.go
//  goray
//
//  Created by Ross Light on 2010-05-28.
//

/* The goray/volume package provides an interface for volumetric effects. */
package volume

import (
	"goray/bound"
	"goray/color"
	"goray/light"
	"goray/ray"
	"goray/vector"
)

/* Region defines a volumetric effect */
type Region interface {
	/* SigmaA returns the amount of light absorption in the volume. */
	SigmaA(p, v vector.Vector3D) color.Color
	/* SigmaS returns the amount of light scattering in the volume. */
	SigmaS(p, v vector.Vector3D) color.Color
	/* Emissions returns the amout of light the volume emits. */
	Emission(p, v vector.Vector3D) color.Color
	SigmaT(p, v vector.Vector3D) color.Color

	/* Attenuation returns how much the volumetric effect dissipates over distance. */
	Attenuation(p vector.Vector3D, l light.Light) float

	P(l, s vector.Vector3D) float

	Tau(r ray.Ray, step, offset float) color.Color

	/* Intersect returns whether a ray intersects the volume. */
	Intersect(r ray.Ray) (t0, t1 float, ok bool)

	/* GetBound returns the bounding box of the volume. */
	GetBound() *bound.Bound
}
