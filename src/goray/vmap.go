//
//  goray/vmap.go
//  goray
//
//  Created by Ross Light on 2010-05-28.
//

package vmap

import (
    "container/vector"
)

// VMap is a vertex map class for triangle meshes
type VMap struct {
    fmap vector.Vector
    dim int
}

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

func (vm *VMap) GetDimension() int { return vm.dim }
func (vm *VMap) Len() int { return vm.fmap.Len() / vm.dim }

func (vm *VMap) GetValue(triangle int) (vals []float, ok bool) {
    vals = make([]float, 3 * vm.dim)
    base := len(vals) * triangle
    end := base + len(vals)
    if end > len(vm.fmap) {
        vals = nil
        return
    }
    ok = true
    for i := 0; i < len(vals); i++ {
        vals[i] = vm.fmap.At(i).(float)
    }
    return
}

func (vm *VMap) SetValue(triangle, vertex int, vals []float) (ok bool) {
    base := (triangle * 3 + vertex) * vm.dim
    if base + vm.dim > vm.fmap.Len() || len(vals) != vm.dim {
        return false
    }
    for i := 0; i < vm.dim; i++ {
        vm.fmap.Set(base + i, vals[i])
    }
    return true
}

func (vm *VMap) PushTriValue(vals []float) (ok bool) {
    n := 3 * vm.dim
    if len(vals) != n {
        return false
    }
    for i := 0; i < n; i++ {
        vm.fmap.Push(vals[i])
    }
    return true
}
