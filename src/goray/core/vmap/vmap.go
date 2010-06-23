//
//  goray/core/vmap/vmap.go
//  goray
//
//  Created by Ross Light on 2010-05-28.
//

/* The vmap package provides a type for efficiently storing per-vertex data. */
package vmap

import (
	"container/vector"
)

/* VMap efficiently stores per-vertex data. */
type VMap struct {
	fmap vector.Vector
	dim  int
}

/* New creates a new vertex map. */
func New(dimension, triCount int) *VMap {
	vm := new(VMap)
	vm.dim = dimension
	vecLen := dimension * triCount * 3
	vm.fmap.Resize(vecLen, vecLen)
	for i := 0; i < vecLen; i++ {
		vm.fmap.Set(i, 0.0)
	}
	return vm
}

/* GetDimension returns the number of values for each vertex. */
func (vm *VMap) GetDimension() int { return vm.dim }

/* Len returns the number of vertices stored in the map. */
func (vm *VMap) Len() int { return vm.fmap.Len() / vm.dim }

/* GetValue returns all of the values for a triangle in the map. */
func (vm *VMap) GetValue(triangle int) (vals []float, ok bool) {
	vals = make([]float, 3*vm.dim)
	base := len(vals) * triangle
	if base+len(vals) > len(vm.fmap) {
		vals = nil
		return
	}
	ok = true
	for i, _ := range vals {
		vals[i] = vm.fmap.At(base + i).(float)
	}
	return
}

/* SetValue changes the values for a vertex. */
func (vm *VMap) SetValue(triangle, vertex int, vals []float) (ok bool) {
	base := (triangle*3 + vertex) * vm.dim
	if base+vm.dim > vm.fmap.Len() || len(vals) != vm.dim {
		return false
	}
	for i, v := range vals {
		vm.fmap.Set(base+i, v)
	}
	return true
}

/* PushTriValue pushes three sets of vertex values.  It increases the size of the map. */
func (vm *VMap) PushTriValue(vals []float) (ok bool) {
	if len(vals) != vm.dim*3 {
		return false
	}
	for _, v := range vals {
		vm.fmap.Push(v)
	}
	return true
}
