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

import (
	"container/heap"

	"bitbucket.org/zombiezen/goray/color"
	"bitbucket.org/zombiezen/goray/kdtree"
	"bitbucket.org/zombiezen/goray/vector"
)

type Photon struct {
	Position  vector.Vector3D
	Direction vector.Vector3D
	Color     color.Color
}

type PhotonMap struct {
	photons      []Photon
	paths        int
	fresh        bool
	searchRadius float64
	tree         *kdtree.Tree
}

func NewMap() *PhotonMap {
	return &PhotonMap{searchRadius: 1.0}
}

func (pm *PhotonMap) NumPaths() int      { return pm.paths }
func (pm *PhotonMap) SetNumPaths(np int) { pm.paths = np }

func (pm *PhotonMap) AddPhoton(p Photon) {
	pm.photons = append(pm.photons, p)
	pm.fresh = false
}

func (pm *PhotonMap) Clear() {
	pm.photons = pm.photons[:0]
	pm.tree = nil
	pm.fresh = false
}

func (pm *PhotonMap) Ready() bool { return pm.fresh }

func photonGetDim(v kdtree.Value, axis vector.Axis) (min, max float64) {
	photon := v.(Photon)
	min = photon.Position[axis]
	max = min
	return
}

func (pm *PhotonMap) Update() {
	pm.tree = nil
	if len(pm.photons) > 0 {
		values := make([]kdtree.Value, len(pm.photons))
		for i, _ := range values {
			values[i] = pm.photons[i]
		}
		opts := kdtree.MakeOptions(photonGetDim, nil)
		opts.LeafSize = 1
		pm.tree = kdtree.New(values, opts)
		pm.fresh = true
	}
}

type GatherResult struct {
	Photon   Photon
	Distance float64
}

type gatherHeap []GatherResult

func (h gatherHeap) Len() int { return len(h) }
func (h gatherHeap) Cap() int { return cap(h) }

func (h gatherHeap) Less(i, j int) bool {
	// This is a max heap.
	return h[i].Distance >= h[j].Distance
}

func (h gatherHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Add inserts a result onto the heap. If the heap is full, then the smallest distance photon is removed to make room.
func (h *gatherHeap) Add(val GatherResult) {
	if len(*h) < cap(*h) {
		h.Push(val)
		if len(*h) == cap(*h) {
			heap.Init(h)
		}
	} else {
		heap.Pop(h)
		heap.Push(h, val)
	}
}

func (h *gatherHeap) Push(val interface{}) {
	*h = append(*h, val.(GatherResult))
}

func (h *gatherHeap) Pop() (val interface{}) {
	val = (*h)[len(*h)-1]
	*h = (*h)[:len(*h)-1]
	return
}

func (pm *PhotonMap) Gather(p vector.Vector3D, nLookup int, maxDist float64) []GatherResult {
	resultHeap := make(gatherHeap, 0, nLookup)

	ch, distCh := make(chan GatherResult), make(chan float64)
	go lookup(p, ch, distCh, pm.tree.Root())
	distCh <- maxDist

	for gresult := range ch {
		resultHeap.Add(gresult)
	}
	return resultHeap
}

func (pm *PhotonMap) FindNearest(p, n vector.Vector3D, dist float64) (nearest Photon) {
	ch, distCh := make(chan GatherResult), make(chan float64)
	go lookup(p, ch, distCh, pm.tree.Root())
	distCh <- dist

	for gresult := range ch {
		if vector.Dot(gresult.Photon.Direction, n) > 0 {
			nearest, dist = gresult.Photon, gresult.Distance
		}
		distCh <- dist
	}
	return
}

func lookup(p vector.Vector3D, ch chan<- GatherResult, distCh <-chan float64, root *kdtree.Node) {
	defer close(ch)
	st := []*kdtree.Node{root}
	maxDistSqr := <-distCh

	next := func() (n *kdtree.Node, empty bool) {
		empty = len(st) == 0
		if empty {
			return
		}
		n = st[len(st)-1]
		st = st[:len(st)-1]
		return
	}

	for currNode, empty := next(); !empty; currNode, empty = next() {
		if currNode.Leaf() {
			phot := currNode.Values()[0].(Photon)
			v := vector.Sub(phot.Position, p)
			distSqr := v.LengthSqr()
			if distSqr < maxDistSqr {
				ch <- GatherResult{phot, distSqr}
				maxDistSqr = <-distCh
			}
			continue
		}

		axis := currNode.Axis()
		dist2 := p[axis] - currNode.Pivot()
		dist2 *= dist2

		primaryChild, altChild := currNode.Left(), currNode.Right()
		if p[axis] > currNode.Pivot() {
			primaryChild, altChild = altChild, primaryChild
		}

		if dist2 < maxDistSqr {
			st = append(st, altChild)
		}
		st = append(st, primaryChild)
	}
}
