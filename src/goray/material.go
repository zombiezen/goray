//
//  goray/material.go
//  goray
//
//  Created by Ross Light on 2010-05-24.
//

package material

import (
	"./goray/color"
)

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
	// TODO: Render state-dependent methods
	IsTransparent() bool
	GetFlags() BSDF
}
