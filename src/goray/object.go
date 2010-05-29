//
//  goray/object.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

package object

import (
	"./goray/light"
	"./goray/primitive"
	"./goray/surface"
	"./goray/vector"
)

type Object3D interface {
	GetPrimitives() []primitive.Primitive
	// Evaluate vertex map (or equivalent for other geometry types if available)
	EvalVmap(sp surface.Point, id uint, val []float) (dim int)
	// Set a light source to be associated with this object
	SetLight(light.Light)
	// Query whether object surface can be sampled right now
	CanSample() bool
	// Try to enable sampling (may require additional memory and preprocessing time, if supported)
	EnableSampling() bool
	// Sample object surface
	Sample(s1, s2 float) (p, n vector.Vector3D)
	IsVisible() bool
	SetVisible(bool)
}

type primitiveObject struct {
	prim   primitive.Primitive
	light  light.Light
	hidden bool
}

func NewPrimitive(p primitive.Primitive) Object3D {
	return &primitiveObject{p, nil, false}
}

func (o *primitiveObject) GetPrimitives() []primitive.Primitive {
	return []primitive.Primitive{o.prim}
}

func (o *primitiveObject) SetLight(l light.Light) { o.light = l }
func (o *primitiveObject) CanSample() bool        { return false }
func (o *primitiveObject) EnableSampling() bool   { return false }
func (o *primitiveObject) IsVisible() bool        { return !o.hidden }
func (o *primitiveObject) SetVisible(v bool)      { o.hidden = !v }

func (o *primitiveObject) EvalVmap(sp surface.Point, id uint, val []float) int {
	return 0
}

func (o *primitiveObject) Sample(s1, s2 float) (p, n vector.Vector3D) {
	return
}
