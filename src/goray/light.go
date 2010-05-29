//
//  goray/light.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

package light

import (
	"./goray/color"
	"./goray/material"
	"./goray/ray"
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
	Col color.Color
	// Flags of the sampled light source
	Flags uint
	// Surface point on the light source, may only be complete enough to call other light methods with it!
	Sp *material.SurfacePoint
}

type Light interface {
	//Init(*Scene)
	TotalEnergy() color.Color
	EmitPhoton(s1, s2, s3, s4 float) (color.Color, ray.Ray, float)
	EmitSample(wo vector.Vector3D) (color.Color, Sample)
	IllumSample(sp material.SurfacePoint, wi ray.Ray) (bool, Sample)
	Illuminate(sp material.SurfacePoint, col color.Color, wi ray.Ray) bool
	CanIntersect() bool
	Intersect(r ray.Ray) (ok bool, dist float, col color.Color, ipdf float)
	IllumPdf(sp, spLight material.SurfacePoint) float
	EmitPdf(sp material.SurfacePoint, wo vector.Vector3D) (areaPdf, dirPdf, cosWo float)
	NumSamples() int
	CanIlluminate(pt vector.Vector3D) bool
	GetFlags() uint
}
