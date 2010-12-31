//
//	goray/core/vmap/vmap.go
//	goray
//
//	Created by Ross Light on 2010-05-28.
//

// The vmap package provides a type for efficiently storing per-vertex data.
package vmap

// VMap efficiently stores per-vertex data.
type VMap struct {
	fmap []float64
	dim  int
}

// New creates a new vertex map.
func New(dimension, triCount int) (vm *VMap) {
	return &VMap{
		fmap: make([]float64, dimension*triCount*3),
		dim:  dimension,
	}
}

// GetDimension returns the number of values for each vertex.
func (vm *VMap) GetDimension() int { return vm.dim }

// Len returns the number of vertices stored in the map.
func (vm *VMap) Len() int { return len(vm.fmap) / vm.dim }

// GetValue returns all of the values for a triangle in the map.
func (vm *VMap) GetValue(triangle int) (vals []float64, ok bool) {
	start := vm.dim * 3 * triangle
	end := start + vm.dim*3
	if end > len(vm.fmap) {
		return
	}
	return vm.fmap[start:end], true
}

// SetValue changes the values for a vertex.
func (vm *VMap) SetValue(triangle, vertex int, vals []float64) (ok bool) {
	base := (triangle*3 + vertex) * vm.dim
	if base+vm.dim > len(vm.fmap) || len(vals) != vm.dim {
		return false
	}
	copy(vm.fmap[base:], vals)
	return true
}

// PushTriValue pushes three sets of vertex values.  It increases the size of the map.
func (vm *VMap) PushTriValue(vals []float64) (ok bool) {
	if len(vals) != vm.dim*3 {
		return false
	}
	vm.fmap = append(vm.fmap, vals...)
	return true
}
