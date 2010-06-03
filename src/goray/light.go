//
//  goray/light.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

/* The goray/light package provides an interface for an entity that provides light. */
package light

import (
	"./goray/color"
	"./goray/ray"
	"./goray/surface"
	"./goray/vector"
)

const (
	TypeDiracDir = 1 << iota // A light with TypeDiracDir has a Dirac delta distribution
	TypeSingular

	TypeNone = 0
)

/* A light sample */
type Sample struct {
	S1, S2  float         // 2D sample value for choosing a surface point on the light.
	S3, S4  float         // 2D sample value for choosing an outgoing direction on the light (EmitSample)
	Pdf     float         // "Standard" directional PDF from illuminated surface point for MC integration of direct lighting (IllumSample)
	DirPdf  float         // Probability density for generating this sample direction (EmitSample)
	AreaPdf float         // Probability density for generating this sample point on light surface (EmitSample)
	Col     color.Color   // Color of the generated sample
	Flags   uint          // Flags of the sampled light source
	Sp      surface.Point // Surface point on the light source.  This may only be complete enough to call other light methods with it!
}

/* An entity that illuminates a scene. */
type Light interface {
	/* SetScene sets up a light for use with a scene. */
	SetScene(scene interface{})
	/* TotalEnergy returns the light's color energy emitted during a frame */
	TotalEnergy() color.Color
	/* EmitPhoton computes the values necessary for a photon */
	EmitPhoton(s1, s2, s3, s4 float) (color.Color, ray.Ray, float)
	/* EmitSample creates a light emission sample.  It's more suited to bidirectional methods than EmitPhoton. */
	EmitSample(wo vector.Vector3D) (color.Color, Sample)
	/* Illuminate a given surface point, generating a sample. */
	IllumSample(sp surface.Point, wi *ray.Ray) (s Sample, ok bool)
	/* Illuminate a given surface point.  Only for Dirac lights. */
	Illuminate(sp surface.Point, wi *ray.Ray) (col color.Color, ok bool)
	/* CanIntersect indicates whether the light can intersect with a ray */
	CanIntersect() bool
	/* Intersect intersects the light source with a ray, giving back the distance, the energy, and 1/PDF. */
	Intersect(r ray.Ray) (ok bool, dist float, col color.Color, ipdf float)
	/* IllumPdf returns the PDF for sampling with IllumSample. */
	IllumPdf(sp, spLight surface.Point) float
	/* EmitPdf returns the PDFs for sampling with EmitSample. */
	EmitPdf(sp surface.Point, wo vector.Vector3D) (areaPdf, dirPdf, cosWo float)
	/* NumSamples returns the preferred number of samples for direct lighting. */
	NumSamples() int
	/* CanIlluminate returns whether the light can illuminate a certain point. */
	CanIlluminate(pt vector.Vector3D) bool
	/* DiracLight indicates whether the light has a Dirac delta distribution. */
	IsDirac() bool
	/* GetFlags returns the type of light the light is. */
	GetFlags() uint
}
