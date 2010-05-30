//
//  goray/material.go
//  goray
//
//  Created by Ross Light on 2010-05-24.
//

package material

import (
	"./goray/color"
	"./goray/render"
	"./goray/ray"
	"./goray/surface"
	"./goray/vector"
)

type VolumeHandler interface {
	Transmittance(state render.State, r ray.Ray) (bool, color.Color)
	Scatter(state render.State, r ray.Ray) (bool, ray.Ray, PhotonSample)
}

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

type PhotonSample struct {
	Sample
	S3                      float
	LastColor, Alpha, Color color.Color
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
	// Initialize the BSDF of a material.  You must call this with the current surface point
	// first before any other methods (except isTransparent/getTransparency)! The renderstate
	// holds a pointer to preallocated userdata to save data that only depends on the current sp,
	// like texture lookups etc.
	InitBSDF(state *render.State, sp surface.Point) BSDF
	Eval(state *render.State, sp surface.Point, wo, wl vector.Vector3D, types BSDF) color.Color
	Sample(state *render.State, sp surface.Point, wo vector.Vector3D) (color.Color, vector.Vector3D, Sample)
	Pdf(state *render.State, sp surface.Point, wo, wi vector.Vector3D, bsdfs BSDF) float
	IsTransparent() bool
	GetTransparency(state *render.State, sp surface.Point, wo vector.Vector3D) color.Color
	GetSpecular(state *render.State, sp surface.Point, wo vector.Vector3D) (reflect, refract bool, dir [2]vector.Vector3D, col [2]color.Color)
	GetReflectivity(state *render.State, sp surface.Point, flags BSDF) color.Color
	Emit(state *render.State, sp surface.Point, wo vector.Vector3D) color.Color
	VolumeTransmittance(state *render.State, sp surface.Point, r ray.Ray, col color.Color) bool
	GetVolumeHandler(inside bool) VolumeHandler
	GetAlpha(state *render.State, sp surface.Point, wo vector.Vector3D) float
	ScatterPhoton(state *render.State, sp surface.Point, wi vector.Vector3D) (bool, vector.Vector3D, PhotonSample)
	GetFlags() BSDF
}
