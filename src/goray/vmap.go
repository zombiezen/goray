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
    if base + len(vals) > len(vm.fmap) {
        vals = nil
        return
    }
    ok = true
    for i, _ := range vals {
        vals[i] = vm.fmap.At(base + i).(float)
    }
    return
}

func (vm *VMap) SetValue(triangle, vertex int, vals []float) (ok bool) {
    base := (triangle * 3 + vertex) * vm.dim
    if base + vm.dim > vm.fmap.Len() || len(vals) != vm.dim {
        return false
    }
    for i, v := range vals {
        vm.fmap.Set(base + i, v)
    }
    return true
}

func (vm *VMap) PushTriValue(vals []float) (ok bool) {
    if len(vals) != vm.dim * 3 {
        return false
    }
    for _, v := range vals {
        vm.fmap.Push(v)
    }
    return true
}
