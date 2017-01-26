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
	"zombiezen.com/go/goray/internal/color"
)

// VolumeHandler defines a type that handles light scattering.
type VolumeHandler interface {
	Transmittance(state *RenderState, r Ray) (color.Color, bool)
	Scatter(state *RenderState, r Ray) (Ray, PhotonSample, bool)
}

// BSDF holds bidirectional scattering distribution function flags.
type BSDF uint

// These constants specify material attributes. For more information, see
// http://en.wikipedia.org/wiki/Bidirectional_scattering_distribution_function
const (
	BSDFSpecular BSDF = 1 << iota
	BSDFGlossy
	BSDFDiffuse
	BSDFDispersive
	BSDFReflect
	BSDFTransmit
	BSDFFilter
	BSDFEmit
	BSDFVolumetric

	BSDFNone        = 0
	BSDFAllSpecular = BSDFSpecular | BSDFReflect | BSDFTransmit
	BSDFAll         = BSDFSpecular | BSDFGlossy | BSDFDiffuse | BSDFDispersive | BSDFReflect | BSDFTransmit | BSDFFilter
)

// MaterialSample is a sample of a material on a surface.
type MaterialSample struct {
	S1, S2              float64
	Pdf                 float64
	Flags, SampledFlags BSDF
	Reverse             bool
	PdfBack             float64
	ColorBack           color.Color
}

func NewMaterialSample(s1, s2 float64) (s MaterialSample) {
	return MaterialSample{
		S1:           s1,
		S2:           s2,
		Flags:        BSDFAll,
		SampledFlags: BSDFNone,
		Reverse:      false,
		ColorBack:    color.Black,
	}
}

// PhotonSample is a sample of a material on a surface, along with photon information.
type PhotonSample struct {
	MaterialSample
	S3        float64
	LastColor color.Color // LastColor is the photon color from the last scattering.
	Alpha     color.Color // Alpha is the filter color between the last scattering and this collision.
	Color     color.Color // Color is the new color after scattering.  LastColor will use this value for the next scatter.
}

func NewPhotonSample(s1, s2, s3 float64, flags BSDF, lCol color.Color) (s PhotonSample) {
	s = PhotonSample{
		MaterialSample: NewMaterialSample(s1, s2),
		S3:             s3,
		LastColor:      lCol,
		Alpha:          color.White,
		Color:          color.Black,
	}
	s.MaterialSample.Flags = flags
	return
}

// Material defines the behavior of the surface properties of an object.
type Material interface {
	// InitBSDF initializes the BSDF of a material.  You must call this with
	// the current surface point first before any other methods (except
	// Transparency).
	InitBSDF(state *RenderState, sp SurfacePoint) BSDF

	// MaterialFlags returns the attributes of a material.
	MaterialFlags() BSDF

	// Eval evaluates the BSDF for the given components.
	Eval(state *RenderState, sp SurfacePoint, wo, wl vec64.Vector, types BSDF) color.Color

	// Sample takes a sample from the BSDF.  The sample pointer will be filled in with the computed values.
	Sample(state *RenderState, sp SurfacePoint, wo vec64.Vector, s *MaterialSample) (color.Color, vec64.Vector)

	// Pdf returns the PDF for sampling the BSDF.
	Pdf(state *RenderState, sp SurfacePoint, wo, wi vec64.Vector, bsdfs BSDF) float64

	// Specular evaluates the specular components of a material for a given direction.
	Specular(state *RenderState, sp SurfacePoint, wo vec64.Vector) (reflect, refract bool, dir [2]vec64.Vector, col [2]color.Color)

	// Reflectivity returns the overal reflectivity of a material.
	Reflectivity(state *RenderState, sp SurfacePoint, flags BSDF) color.Color

	// Alpha returns the alpha value of a material.
	Alpha(state *RenderState, sp SurfacePoint, wo vec64.Vector) float64

	// ScatterPhoton performs photon mapping.  The sample pointer will be filled in with the computed values.
	ScatterPhoton(state *RenderState, sp SurfacePoint, wi vec64.Vector, s *PhotonSample) (wo vec64.Vector, scattered bool)
}

// TransparentMaterial defines a material that can allow light to pass through it.
type TransparentMaterial interface {
	Material
	// Transparency returns the color that the light is multiplied by when
	// passing through it.  If the color is black, then the material is opaque.
	Transparency(state *RenderState, sp SurfacePoint, wo vec64.Vector) color.Color
}

// EmitMaterial defines a material that contributes light to the scene.
type EmitMaterial interface {
	Material
	// Emit returns the amount of light to contribute.
	Emit(state *RenderState, sp SurfacePoint, wo vec64.Vector) color.Color
}

// VolumetricMaterial defines a material that is aware of volumetric effects.
type VolumetricMaterial interface {
	Material
	// VolumeTransmittance allows attenuation due to intra-object volumetric effects.
	VolumeTransmittance(state *RenderState, sp SurfacePoint, r Ray) (color.Color, bool)

	// VolumeHandler returns the volumetric handler for the space on a given side.
	VolumeHandler(inside bool) VolumeHandler
}
