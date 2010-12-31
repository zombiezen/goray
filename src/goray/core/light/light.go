//
//	goray/core/light/light.go
//	goray
//
//	Created by Ross Light on 2010-05-23.
//

// The light package provides an interface for an entity that provides light.
package light

import (
	"goray/core/color"
	"goray/core/ray"
	"goray/core/surface"
	"goray/core/vector"
)

const (
	TypeDiracDir = 1 << iota // A light with TypeDiracDir has a Dirac delta distribution
	TypeSingular

	TypeNone = 0
)

// Sample holds data for a light sample.
type Sample struct {
	S1, S2  float64       // 2D sample value for choosing a surface point on the light.
	S3, S4  float64       // 2D sample value for choosing an outgoing direction on the light (EmitSample)
	Pdf     float64       // "Standard" directional PDF from illuminated surface point for MC integration of direct lighting (IllumSample)
	DirPdf  float64       // Probability density for generating this sample direction (EmitSample)
	AreaPdf float64       // Probability density for generating this sample point on light surface (EmitSample)
	Color   color.Color   // Color of the generated sample
	Flags   uint          // Flags of the sampled light source
	Point   surface.Point // Surface point on the light source.  This may only be complete enough to call other light methods with it!
}

// An entity that illuminates a scene.
type Light interface {
	// GetFlags returns the type of light the light is.
	GetFlags() uint
	// SetScene sets up a light for use with a scene.
	SetScene(scene interface{})
	// NumSamples returns the preferred number of samples for direct lighting.
	NumSamples() int
	// TotalEnergy returns the light's color energy emitted during a frame.
	TotalEnergy() color.Color
	// EmitPhoton computes the values necessary for a photon.
	EmitPhoton(s1, s2, s3, s4 float64) (color.Color, ray.Ray, float64)
	// EmitSample creates a light emission sample.  It's more suited to bidirectional methods than EmitPhoton.
	EmitSample(s *Sample) (vector.Vector3D, color.Color)
	// EmitPdf returns the PDFs for sampling with EmitSample.
	EmitPdf(sp surface.Point, wo vector.Vector3D) (areaPdf, dirPdf, cosWo float64)
	// CanIlluminate returns whether the light can illuminate a certain point.
	CanIlluminate(pt vector.Vector3D) bool
	// IlluminateSample samples the illumination at a given point.
	//
	// The Sample passed in will be filled with the proper sample values.
	// The returned ray will be the ray that casted the light.
	IlluminateSample(sp surface.Point, wi ray.Ray, s *Sample) (wo ray.Ray, illuminated bool)
	// IlluminatePdf returns the PDF for sampling with IllumSample.
	IlluminatePdf(sp, spLight surface.Point) float64
}

type Intersecter interface {
	// Intersect intersects the light source with a ray, giving back the distance, the energy, and 1/PDF.
	Intersect(r ray.Ray) (dist float64, col color.Color, ipdf float64, ok bool)
}

type DiracLight interface {
	Light

	// Illuminate computes the amount of light to add to a given surface point.
	Illuminate(sp surface.Point, wi ray.Ray) (col color.Color, wo ray.Ray, ok bool)
}
