//
//  goray/core/object/object.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

/* The object package provides an object type—a collection of primitives. */
package object

import (
	"goray/core/primitive"
	"goray/core/vector"
)

/* Object3D is a collection of primitives. */
type Object3D interface {
	/* GetPrimitives retrieves all of the primitives for this object. */
	GetPrimitives() []primitive.Primitive
	/* IsVisible indicates whether the object is shown in the scene. */
	IsVisible() bool
}

/* Samplable defines an interface for sampling a surface. */
type Samplable interface {
	/* EnableSampling tries to enable sampling (may require additional memory and preprocessing time). */
	EnableSampling() bool
	/* Sample takes a sample of the object's surface. */
	Sample(s1, s2 float) (p, n vector.Vector3D)
}

/* SamplableObject3D is the set of three-dimensional objects that can sample their surfaces. */
type SamplableObject3D interface {
	Object3D
	Samplable
}

/* PrimitiveObject is a wrapper type that allows a single primitive to act as an object. */
type PrimitiveObject struct {
    Primitive primitive.Primitive
}

func (o PrimitiveObject) GetPrimitives() []primitive.Primitive {
	return []primitive.Primitive{o.Primitive}
}

func (o PrimitiveObject) IsVisible() bool { return true }
