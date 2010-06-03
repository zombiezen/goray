//
//  goray/object.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

/* The goray/object package provides an object type--a collection of primitives. */
package object

import (
	"./goray/light"
	"./goray/primitive"
	"./goray/surface"
	"./goray/vector"
)

/* Object3D is a collection of primitives */
type Object3D interface {
	/* GetPrimitives retrieves all of the primitives for this object. */
	GetPrimitives() []primitive.Primitive
	/* EvalVmap evaluates a vertex map (or equivalent for other geometry types if available) */
	EvalVmap(sp surface.Point, id uint, val []float) (dim int)
	/* SetLight associates a light source with this object */
	SetLight(light.Light)
	/* CanSample indicates whether the object's surface can be sampled */
	CanSample() bool
	/* EnableSampling tries to enable sampling (may require additional memory and preprocessing time, if supported) */
	EnableSampling() bool
	/* Sample takes a sample of the object's surface */
	Sample(s1, s2 float) (p, n vector.Vector3D)
	/* IsVisible indicates whether the object is shown in the scene. */
	IsVisible() bool
	/* SetVisible changes whether the object is shown in the scene. */
	SetVisible(bool)
}

type primitiveObject struct {
	prim   primitive.Primitive
	light  light.Light
	hidden bool
}

/* NewPrimitive creates an object that contains only one primitive. */
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
