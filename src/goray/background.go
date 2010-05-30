//
//  goray/background.go
//  goray
//
//  Created by Ross Light on 2010-05-28.
//

package background

import (
	"./goray/color"
	"./goray/light"
	"./goray/ray"
	"./goray/render"
)

type Background interface {
	// Get the background color for a given ray
	GetColor(r ray.Ray, state *render.State, filtered bool) color.Color
	Eval(r ray.Ray, filtered bool)
	GetLight() light.Light
}
