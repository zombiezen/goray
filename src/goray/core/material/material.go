//
//	goray/core/material/material.go
//	goray
//
//	Created by Ross Light on 2010-05-24.
//

// The material package provides a set of interfaces for dealing with object materials.
package material

import (
	"goray/core/color"
	"goray/core/render"
	"goray/core/ray"
	"goray/core/surface"
	"goray/core/vector"
)

// VolumeHandler defines a type that handles light scattering.
type VolumeHandler interface {
	Transmittance(state *render.State, r ray.Ray) (color.Color, bool)
	Scatter(state *render.State, r ray.Ray) (ray.Ray, PhotonSample, bool)
}

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

// Sample is a sample of a material on a surface.
type Sample struct {
	S1, S2              float64
	Pdf                 float64
	Flags, SampledFlags BSDF
	Reverse             bool
	PdfBack             float64
	ColorBack           color.Color
}

func NewSample(s1, s2 float64) (s Sample) {
	return Sample{
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
	Sample
	S3        float64
	LastColor color.Color // LastColor is the photon color from the last scattering.
	Alpha     color.Color // Alpha is the filter color between the last scattering and this collision.
	Color     color.Color // Color is the new color after scattering.  LastColor will use this value for the next scatter.
}

func NewPhotonSample(s1, s2, s3 float64, flags BSDF, lCol color.Color) (s PhotonSample) {
	s = PhotonSample{
		Sample:    NewSample(s1, s2),
		S3:        s3,
		LastColor: lCol,
		Alpha:     color.White,
		Color:     color.Black,
	}
	s.Sample.Flags = flags
	return
}

// Material defines the behavior of the surface properties of an object.
type Material interface {
	// InitBSDF initializes the BSDF of a material.  You must call this with
	// the current surface point first before any other methods (except
	// Transparency).
	InitBSDF(state *render.State, sp surface.Point) BSDF
	// MaterialFlags returns the attributes of a material.
	MaterialFlags() BSDF
	// Eval evaluates the BSDF for the given components.
	Eval(state *render.State, sp surface.Point, wo, wl vector.Vector3D, types BSDF) color.Color
	// Sample takes a sample from the BSDF.  The sample pointer will be filled in with the computed values.
	Sample(state *render.State, sp surface.Point, wo vector.Vector3D, s *Sample) (color.Color, vector.Vector3D)
	// Pdf returns the PDF for sampling the BSDF.
	Pdf(state *render.State, sp surface.Point, wo, wi vector.Vector3D, bsdfs BSDF) float64
	// Specular evaluates the specular components of a material for a given direction.
	Specular(state *render.State, sp surface.Point, wo vector.Vector3D) (reflect, refract bool, dir [2]vector.Vector3D, col [2]color.Color)
	// Reflectivity returns the overal reflectivity of a material.
	Reflectivity(state *render.State, sp surface.Point, flags BSDF) color.Color
	// Alpha returns the alpha value of a material.
	Alpha(state *render.State, sp surface.Point, wo vector.Vector3D) float64
	// ScatterPhoton performs photon mapping.  The sample pointer will be filled in with the computed values.
	ScatterPhoton(state *render.State, sp surface.Point, wi vector.Vector3D, s *PhotonSample) (wo vector.Vector3D, scattered bool)
}

// TransparentMaterial defines a material that can allow light to pass through it.
type TransparentMaterial interface {
	Material
	// Transparency is used when computing transparent shadows.
	Transparency(state *render.State, sp surface.Point, wo vector.Vector3D) color.Color
}

// EmitMaterial defines a material that contributes light to the scene.
type EmitMaterial interface {
	Material
	// Emit returns the amount of light to contribute.
	Emit(state *render.State, sp surface.Point, wo vector.Vector3D) color.Color
}

// VolumetricMaterial defines a material that is aware of volumetric effects.
type VolumetricMaterial interface {
	Material
	// VolumeTransmittance allows attenuation due to intra-object volumetric effects.
	VolumeTransmittance(state *render.State, sp surface.Point, r ray.Ray) (color.Color, bool)
	// VolumeHandler returns the volumetric handler for the space on a given side.
	VolumeHandler(inside bool) VolumeHandler
}
