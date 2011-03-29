//
//	goray/core/background/background.go
//	goray
//
//	Created by Ross Light on 2010-05-28.
//

// The background package provides an interface for a rendering background.
package background

import (
	"goray/core/color"
	"goray/core/light"
	"goray/core/ray"
	"goray/core/render"
)

// A rendering background
type Background interface {
	// Color returns the background color for a given ray.
	Color(r ray.Ray, state *render.State, filtered bool) color.Color
	// Light returns the light source representing background lighting.
	// This may be nil if the background should only be sampled from BSDFs.
	Light() light.Light
}
