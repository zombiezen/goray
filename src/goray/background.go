//
//  goray/background.go
//  goray
//
//  Created by Ross Light on 2010-05-28.
//

/*
   The goray/background package provides an interface for a rendering
   background.
*/
package background

import (
	"goray/color"
	"goray/light"
	"goray/ray"
	"goray/render"
)

/* A rendering background */
type Background interface {
	/* GetColor returns the background color for a given ray */
	GetColor(r ray.Ray, state *render.State, filtered bool) color.Color
	/* GetLight returns the light source representing background lighting.
	   This may be nil if the background should only be sampled from BSDFs */
	GetLight() light.Light
}
