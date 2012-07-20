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
	"bitbucket.org/zombiezen/goray/color"
	"bitbucket.org/zombiezen/math3/vec64"
)

// Light types
const (
	LightTypeDiracDir = 1 << iota // A light with TypeDiracDir has a Dirac delta distribution
	LightTypeSingular

	LightTypeNone = 0
)

// LightSample holds data for a light sample.
type LightSample struct {
	S1, S2  float64      // 2D sample value for choosing a surface point on the light.
	S3, S4  float64      // 2D sample value for choosing an outgoing direction on the light (EmitSample)
	Pdf     float64      // "Standard" directional PDF from illuminated surface point for MC integration of direct lighting (IllumSample)
	DirPdf  float64      // Probability density for generating this sample direction (EmitSample)
	AreaPdf float64      // Probability density for generating this sample point on light surface (EmitSample)
	Color   color.Color  // Color of the generated sample
	Flags   uint         // Flags of the sampled light source
	Point   SurfacePoint // Surface point on the light source.  This may only be complete enough to call other light methods with it!
}

// Light is an entity that illuminates a scene.
type Light interface {
	// LightFlags returns the type of light the light is.
	LightFlags() uint

	// SetScene sets up a light for use with a scene.
	SetScene(scene *Scene)

	// NumSamples returns the preferred number of samples for direct lighting.
	NumSamples() int

	// TotalEnergy returns the light's color energy emitted during a frame.
	TotalEnergy() color.Color

	// EmitPhoton computes the values necessary for a photon.
	EmitPhoton(s1, s2, s3, s4 float64) (color.Color, Ray, float64)

	// EmitSample creates a light emission sample.  It's more suited to bidirectional methods than EmitPhoton.
	EmitSample(s *LightSample) (vec64.Vector, color.Color)

	// EmitPdf returns the PDFs for sampling with EmitSample.
	EmitPdf(sp SurfacePoint, wo vec64.Vector) (areaPdf, dirPdf, cosWo float64)

	// CanIlluminate returns whether the light can illuminate a certain point.
	CanIlluminate(pt vec64.Vector) bool

	// IlluminateSample samples the illumination at a given point.
	//
	// The Sample passed in will be filled with the proper sample values.
	// The returned ray will be the ray that casted the light.
	IlluminateSample(sp SurfacePoint, wi *Ray, s *LightSample) (illuminated bool)

	// IlluminatePdf returns the PDF for sampling with IllumSample.
	IlluminatePdf(sp, spLight SurfacePoint) float64
}

type LightIntersecter interface {
	// Intersect intersects the light source with a ray, giving back the distance, the energy, and 1/PDF.
	Intersect(r Ray) (dist float64, col color.Color, ipdf float64, ok bool)
}

type DiracLight interface {
	Light

	// Illuminate computes the amount of light to add to a given surface point.
	Illuminate(sp SurfacePoint, wi *Ray) (col color.Color, ok bool)
}
