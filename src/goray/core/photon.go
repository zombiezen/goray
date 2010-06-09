//
//  goray/core/photon.go
//  goray
//
//  Created by Ross Light on 2010-06-06
//

package photon

import (
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

type PhotonMap struct {
	photons      []*Photon
	paths        int
	fresh        bool
	searchRadius float
	tree         *kdtree.Tree
}

func NewMap() *PhotonMap {
	return &PhotonMap{photons: make([]*Photon, 0), searchRadius: 1.0}
}

func (pm *PhotonMap) GetNumPaths() int   { return pm.paths }
func (pm *PhotonMap) SetNumPaths(np int) { pm.paths = np }

func (pm *PhotonMap) AddPhoton(p *Photon) {
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

func (pm *PhotonMap) Clear() {
	pm.photons = pm.photons[0:0]
	pm.tree = nil
	pm.fresh = false
}

func (pm *PhotonMap) Ready() bool { return pm.fresh }

func photonGetDim(v kdtree.Value, axis int) (min, max float) {
	photon := v.(*Photon)
	min = photon.position.GetComponent(axis)
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
		pm.tree = kdtree.New(values, photonGetDim, nil)
		pm.fresh = true
	}
}
