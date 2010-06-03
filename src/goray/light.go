//
//  goray/light.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

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

type Sample struct {
	// 2D sample value for choosing a surface point on the light.
	S1, S2 float
	// 2D sample value for choosing an outgoing direction on the light (EmitSample)
	S3, S4 float
	// "Standard" directional PDF from illuminated surface point for MC integration of direct lighting (illumSample)
	Pdf float
	// Probability density for generating this sample direction (emitSample)
	DirPdf float
	// Probability density for generating this sample point on light surface (emitSample)
	AreaPdf float
	// Color of the generated sample
	Color color.Color
	// Flags of the sampled light source
	Flags uint
	// Surface point on the light source, may only be complete enough to call other light methods with it!
	Sp surface.Point
}

type Light interface {
	Init(scene interface{})
	TotalEnergy() color.Color
	EmitPhoton(s1, s2, s3, s4 float) (color.Color, ray.Ray, float)
	EmitSample(wo vector.Vector3D) (color.Color, Sample)
	IllumSample(sp surface.Point, wi *ray.Ray) (s Sample, ok bool)
	Illuminate(sp surface.Point, wi *ray.Ray) (col color.Color, ok bool)
	CanIntersect() bool
	Intersect(r ray.Ray) (ok bool, dist float, col color.Color, ipdf float)
	IllumPdf(sp, spLight surface.Point) float
	EmitPdf(sp surface.Point, wo vector.Vector3D) (areaPdf, dirPdf, cosWo float)
	NumSamples() int
	CanIlluminate(pt vector.Vector3D) bool
	GetFlags() uint
}
