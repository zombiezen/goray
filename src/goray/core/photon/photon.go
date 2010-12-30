//
//	goray/core/photon/photon.go
//	goray
//
//	Created by Ross Light on 2010-06-06
//

package photon

import (
	"container/heap"
	"goray/stack"
	"goray/core/color"
	"goray/core/kdtree"
	"goray/core/vector"
)

type Photon struct {
	position  vector.Vector3D
	direction vector.Vector3D
	color     color.Color
}

func New(position, direction vector.Vector3D, col color.Color) *Photon {
	return &Photon{position, direction, col}
}

func (p *Photon) GetPosition() vector.Vector3D  { return p.position }
func (p *Photon) GetDirection() vector.Vector3D { return p.direction }
func (p *Photon) GetColor() color.Color         { return p.color }

func (p *Photon) SetPosition(v vector.Vector3D) { p.position = v }

func (p *Photon) SetDirection(v vector.Vector3D) { p.direction = v }

func (p *Photon) SetColor(c color.Color) { p.color = c }

type Map struct {
	photons      []*Photon
	paths        int
	fresh        bool
	searchRadius float
	tree         *kdtree.Tree
}

func NewMap() *Map {
	return &Map{photons: make([]*Photon, 0), searchRadius: 1.0}
}

func (pm *Map) GetNumPaths() int   { return pm.paths }
func (pm *Map) SetNumPaths(np int) { pm.paths = np }

func (pm *Map) AddPhoton(p *Photon) {
	sliceLen := len(pm.photons)
	if cap(pm.photons) < sliceLen+1 {
		newPhotons := make([]*Photon, sliceLen, (sliceLen+1)*2)
		copy(newPhotons, pm.photons)
		pm.photons = newPhotons
	}
	pm.photons = pm.photons[0 : sliceLen+1]
	pm.photons[sliceLen] = p
	pm.fresh = false
}

func (pm *Map) Clear() {
	pm.photons = pm.photons[0:0]
	pm.tree = nil
	pm.fresh = false
}

func (pm *Map) Ready() bool { return pm.fresh }

func photonGetDim(v kdtree.Value, axis vector.Axis) (min, max float64) {
	photon := v.(*Photon)
	min = photon.position[axis]
	max = min
	return
}

func (pm *Map) Update() {
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
	Photon   *Photon
	Distance float64
}

type gatherHeap []GatherResult

func (h gatherHeap) Len() int { return len(h) }
func (h gatherHeap) Cap() int { return cap(h) }

func (h gatherHeap) Less(i, j int) bool {
	// This is a max heap.
	data := []GatherResult(h)
	return data[i].Distance >= data[j].Distance
}

func (h *gatherHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

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
	data := []GatherResult(*h)
	oldSize := len(data)
	if oldSize+1 > cap(data) {
		temp := make([]GatherResult, oldSize, (oldSize+1)*2)
		copy(temp, data)
		data = temp
	}
	data = data[0 : oldSize+1]
	data[oldSize] = val.(GatherResult)
	*h = data
}

func (h *gatherHeap) Pop() (val interface{}) {
	data := []GatherResult(*h)
	val = data[len(data)-1]
	*h = data[0 : len(data)-1]
	return
}

func (pm *Map) Gather(p vector.Vector3D, nLookup int, maxDist float64) []GatherResult {
	resultHeap := make(gatherHeap, 0, nLookup)

	ch, distCh := make(chan GatherResult), make(chan float64)
	go lookup(p, ch, distCh, pm.tree.GetRoot())
	distCh <- maxDist

	for gresult := range ch {
		resultHeap.Add(gresult)
	}
	return resultHeap
}

func (pm *Map) FindNearest(p, n vector.Vector3D, dist float64) (nearest *Photon) {
	ch, distCh := make(chan GatherResult), make(chan float64)
	go lookup(p, ch, distCh, pm.tree.GetRoot())
	distCh <- dist

	for gresult := range ch {
		if vector.Dot(gresult.Photon.GetDirection(), n) > 0 {
			nearest, dist = gresult.Photon, gresult.Distance
		}
		distCh <- dist
	}
	return
}

func lookup(p vector.Vector3D, ch chan<- GatherResult, distCh <-chan float64, root kdtree.Node) {
	defer close(ch)
	st := stack.New()
	st.Push(root)
	maxDistSqr := <-distCh

	next := func() (kdtree.Node, bool) {
		empty := st.Empty()
		top := st.Pop()
		return top.(kdtree.Node), empty
	}

	for currNode, empty := next(); !empty; currNode, empty = next() {
		if currNode.IsLeaf() {
			leaf := currNode.(*kdtree.Leaf)
			phot := leaf.GetValues()[0].(*Photon)
			v := vector.Sub(phot.position, p)
			distSqr := v.LengthSqr()
			if distSqr < maxDistSqr {
				ch <- GatherResult{phot, distSqr}
				maxDistSqr = <-distCh
			}
			continue
		}

		currInt := currNode.(*kdtree.Interior)
		axis := currInt.GetAxis()
		dist2 := p[axis] - currInt.GetPivot()
		dist2 *= dist2

		primaryChild, altChild := currInt.GetLeft(), currInt.GetRight()
		if p[axis] > currInt.GetPivot() {
			primaryChild, altChild = altChild, primaryChild
		}

		if dist2 < maxDistSqr {
			st.Push(altChild)
		}
		st.Push(primaryChild)
	}
}
