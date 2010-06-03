//
//  goray/material.go
//  goray
//
//  Created by Ross Light on 2010-05-24.
//

/* The goray/material package provides a set of interfaces for dealing with object materials. */
package material

import (
	"./goray/color"
	"./goray/render"
	"./goray/ray"
	"./goray/surface"
	"./goray/vector"
)

/* VolumeHandler defines a type that handles light scattering. */
type VolumeHandler interface {
	Transmittance(state *render.State, r ray.Ray) (color.Color, bool)
	Scatter(state *render.State, r ray.Ray) (ray.Ray, PhotonSample, bool)
}

/* These constants specify material attributes.
   For more information, see http://en.wikipedia.org/wiki/Bidirectional_scattering_distribution_function */
const (
	BSDFSpecular = 1 << iota
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

type BSDF uint

/* Sample is a sample of a material on a surface. */
type Sample struct {
	S1, S2              float
	Pdf                 float
	Flags, SampledFlags BSDF
	Reverse             bool
	PdfBack             float
	ColorBack           color.Color
}

func NewSample(s1, s2 float) Sample {
	s := Sample{S1: s1, S2: s2}
	s.Flags = BSDFAll
	s.SampledFlags = BSDFNone
	s.Reverse = false
	s.Pdf = 0.0
	s.ColorBack = color.NewRGB(0.0, 0.0, 0.0)
	return s
}

/* PhotonSample is a sample of a material on a surface, along with photon information. */
type PhotonSample struct {
	Sample
	S3        float
	LastColor color.Color // LastColor is the photon color from the last scattering.
	Alpha     color.Color // Alpha is the filter color between the last scattering and this collision.
	Color     color.Color // Color is the new color after scattering.  LastColor will use this value for the next scatter.
}

func NewPhotonSample(s1, s2, s3 float, flags BSDF, lCol color.Color) PhotonSample {
	s := PhotonSample{S3: s3}
	s.Sample = NewSample(s1, s2)
	s.Sample.Flags = flags
	s.LastColor = lCol
	s.Alpha = color.NewRGB(1.0, 1.0, 1.0)
	s.Color = color.NewRGB(0.0, 0.0, 0.0)
	return s
}

type Material interface {
	/* InitBSDF initializes the BSDF of a material.  You must call this with the current surface point
	   first before any other methods (except IsTransparent/GetTransparency). */
	InitBSDF(state *render.State, sp surface.Point) BSDF
	/* Eval evaluates the BSDF for the given components. */
	Eval(state *render.State, sp surface.Point, wo, wl vector.Vector3D, types BSDF) color.Color
	/* Sample takes a sample from the BSDF.  The sample pointer will be filled in with the computed values. */
	Sample(state *render.State, sp surface.Point, wo vector.Vector3D, s *Sample) (color.Color, vector.Vector3D)
	/* Pdf returns the PDF for sampling the BSDF. */
	Pdf(state *render.State, sp surface.Point, wo, wi vector.Vector3D, bsdfs BSDF) float
	/* IsTransparent indicates whether light can (at least partially) pass through the material without getting refracted. */
	IsTransparent() bool
	/* GetTransparency is used when computing transparent shadows. */
	GetTransparency(state *render.State, sp surface.Point, wo vector.Vector3D) color.Color
	/* GetSpecular evaluates the specular components of a material for a given direction. */
	GetSpecular(state *render.State, sp surface.Point, wo vector.Vector3D) (reflect, refract bool, dir [2]vector.Vector3D, col [2]color.Color)
	/* GetReflectivity returns the overal reflectivity of a material. */
	GetReflectivity(state *render.State, sp surface.Point, flags BSDF) color.Color
	/* Emit allows light-emitting materials. */
	Emit(state *render.State, sp surface.Point, wo vector.Vector3D) color.Color
	/* VolumeTransmittance allows attenuation due to intra-object volumetric effects. */
	VolumeTransmittance(state *render.State, sp surface.Point, r ray.Ray) (color.Color, bool)
	/* GetVolumeHandler returns the volumetric handler for the space on a given side. */
	GetVolumeHandler(inside bool) VolumeHandler
	/* GetAlpha returns the alpha value of a material. */
	GetAlpha(state *render.State, sp surface.Point, wo vector.Vector3D) float
	/* ScatterPhoton performs photon mapping.  The sample pointer will be filled in with the computed values. */
	ScatterPhoton(state *render.State, sp surface.Point, wi vector.Vector3D, s *PhotonSample) (wo vector.Vector3D, scattered bool)
	/* GetFlags returns the attributes of a material. */
	GetFlags() BSDF
}
