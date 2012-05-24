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
	"bitbucket.org/zombiezen/goray/bound"
	"bitbucket.org/zombiezen/goray/color"
	"bitbucket.org/zombiezen/goray/vector"
)

// VolumeRegion defines a volumetric effect.
type VolumeRegion interface {
	// SigmaA returns the amount of light absorption in the volume.
	SigmaA(p, v vector.Vector3D) color.Color

	// SigmaS returns the amount of light scattering in the volume.
	SigmaS(p, v vector.Vector3D) color.Color

	// Emission returns the amout of light the volume emits.
	Emission(p, v vector.Vector3D) color.Color

	SigmaT(p, v vector.Vector3D) color.Color

	// Attenuation returns how much the volumetric effect dissipates over distance.
	Attenuation(p vector.Vector3D, l Light) float64

	P(l, s vector.Vector3D) float64

	Tau(r Ray, step, offset float64) color.Color

	// Intersect returns whether a ray intersects the volume.
	Intersect(r Ray) (t0, t1 float64, ok bool)

	// Bound returns the bounding box of the volume.
	Bound() bound.Bound
}
