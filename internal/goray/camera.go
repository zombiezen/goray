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

// Camera is a viewpoint of a scene.
type Camera interface {
	// ShootRay calculates the initial ray used for computing a fragment of the
	// output.  U and V are sample coordinates that are only calculated if
	// SampleLens returns true.
	ShootRay(x, y, u, v float64) (Ray, float64)

	// ResolutionX returns the number of fragments wide that the camera is.
	ResolutionX() int

	// ResolutionY returns the number of fragments high that the camera is.
	ResolutionY() int

	// Project calculates the projection of a ray onto the fragment plane.
	Project(wo Ray, lu, lv *float64) (pdf float64, changed bool)

	// SampleLens returns whether the lens needs to be sampled using the u and v
	// parameters of ShootRay.  This is useful for DOF-like effects.  When this
	// returns false, no lens samples need to be computed.
	SampleLens() bool
}
