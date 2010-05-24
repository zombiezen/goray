//
//  goray/object.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

package goray

type Object3D interface {
    NumPrimitives() uint
    GetPrimitives() []*Primitive
    IsVisible() bool
    SetVisibility(bool)
}

type Object3D struct {
    light *Light
    hidden bool
}


