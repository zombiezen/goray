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
	"fmt"
)

// Ray defines the path of light.
type Ray struct {
	From       vec64.Vector
	Dir        vec64.Vector
	TMin, TMax float64
	Time       float64
}

func (r Ray) String() string {
	return fmt.Sprintf("Ray{From: %v, Dir: %v, TMin: %.4f, TMax: %.4f, Time: %.4f}", r.From, r.Dir, r.TMin, r.TMax, r.Time)
}

// DifferentialRay stores additional information about a ray for use in surface intersections.
// For an explanation, see http://www.opticalres.com/white%20papers/DifferentialRayTracing.pdf
type DifferentialRay struct {
	Ray
	FromX, FromY vec64.Vector
	DirX, DirY   vec64.Vector
}

func (r DifferentialRay) String() string {
	return fmt.Sprintf("DifferentialRay{Ray: %v, FromX: %v, FromY: %v, DirX: %v, DirY: %v}", r.Ray, r.FromX, r.FromY, r.DirX, r.DirY)
}
