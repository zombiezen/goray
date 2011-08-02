/*
	Copyright (c) 2011 Ross Light.
	Copyright (c) 2005 Mathias Wein, Alejandro Conty, and Alfredo de Greef.

	This file is part of goray.

	goray is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	goray is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with goray.  If not, see <http://www.gnu.org/licenses/>.
*/

package goray

// Vmap efficiently stores per-vertex data.
type Vmap struct {
	fmap []float64
	dim  int
}

// NewVmap creates a new vertex map.
func NewVmap(dimension, triCount int) (vm *Vmap) {
	return &Vmap{
		fmap: make([]float64, dimension*triCount*3),
		dim:  dimension,
	}
}

// Dimension returns the number of values for each vertex.
func (vm *Vmap) Dimension() int { return vm.dim }

// Len returns the number of vertices stored in the map.
func (vm *Vmap) Len() int { return len(vm.fmap) / vm.dim }

// GetValue returns all of the values for a triangle in the map.
func (vm *Vmap) GetValue(triangle int) (vals []float64, ok bool) {
	start := vm.dim * 3 * triangle
	end := start + vm.dim*3
	if end > len(vm.fmap) {
		return
	}
	return vm.fmap[start:end], true
}

// SetValue changes the values for a vertex.
func (vm *Vmap) SetValue(triangle, vertex int, vals []float64) (ok bool) {
	base := (triangle*3 + vertex) * vm.dim
	if base+vm.dim > len(vm.fmap) || len(vals) != vm.dim {
		return false
	}
	copy(vm.fmap[base:], vals)
	return true
}

// PushTriValue pushes three sets of vertex values.  It increases the size of the map.
func (vm *Vmap) PushTriValue(vals []float64) (ok bool) {
	if len(vals) != vm.dim*3 {
		return false
	}
	vm.fmap = append(vm.fmap, vals...)
	return true
}
