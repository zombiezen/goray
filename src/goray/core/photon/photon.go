//
//	goray/core/photon/photon.go
//	goray
//
//	Created by Ross Light on 2010-06-06
//

package photon

import (
	"container/heap"
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
	searchRadius float64
	tree         *kdtree.Tree
}

func NewMap() *Map {
	return &Map{photons: make([]*Photon, 0), searchRadius: 1.0}
}

func (pm *Map) GetNumPaths() int   { return pm.paths }
func (pm *Map) SetNumPaths(np int) { pm.paths = np }

func (pm *Map) AddPhoton(p *Photon) {
	pm.photons = append(pm.photons, p)
	pm.fresh = false
}

func (pm *Map) Clear() {
	for i, _ := range pm.photons {
		pm.photons[i] = nil
	}
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
	(*h)[len(*h)-1].Photon = nil
	*h = (*h)[:len(*h)-1]
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
	st := []kdtree.Node{root}
	maxDistSqr := <-distCh

	next := func() (n kdtree.Node, empty bool) {
		empty = len(st) == 0
		if empty {
			return
		}
		n = st[len(st)-1]
		st = st[:len(st)-1]
		return
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
			st = append(st, altChild)
		}
		st = append(st, primaryChild)
	}
}
