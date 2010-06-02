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
	TypeDiracDir = 1 << iota
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

/* An entity that emits light */
type Light interface {
	/* SetScene sets up a light for use with a scene. */
	SetScene(scene interface{})
	/* TotalEnergy returns the light's color energy */
	TotalEnergy() color.Color
	EmitPhoton(s1, s2, s3, s4 float) (color.Color, ray.Ray, float)
	EmitSample(wo vector.Vector3D) (color.Color, Sample)
	IllumSample(sp surface.Point, wi ray.Ray) (bool, Sample)
	Illuminate(sp surface.Point, col color.Color, wi ray.Ray) bool
	CanIntersect() bool
	Intersect(r ray.Ray) (ok bool, dist float, col color.Color, ipdf float)
	IllumPdf(sp, spLight surface.Point) float
	EmitPdf(sp surface.Point, wo vector.Vector3D) (areaPdf, dirPdf, cosWo float)
	NumSamples() int
	CanIlluminate(pt vector.Vector3D) bool
	GetFlags() uint
}
