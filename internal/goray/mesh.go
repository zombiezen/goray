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
	"bitbucket.org/zombiezen/math3/vec64"
)

// UV holds a set of texture coordinates.
type UV [2]float64

// A Mesh is a collection of triangles.
// The basic workflow for making a working mesh is: create the mesh, set the mesh's data, then add the triangles.
type Mesh struct {
	triangles []*Triangle
	vertices  []vec64.Vector
	normals   []vec64.Vector
	uvs       []UV
	hasOrco   bool
	light     Light
	hidden    bool
}

var _ Object3D = &Mesh{}

// NewMesh creates an empty mesh.
func NewMesh(ntris int, hasOrco bool) (mesh *Mesh) {
	mesh = new(Mesh)
	mesh.triangles = make([]*Triangle, 0, ntris)
	mesh.vertices = nil
	mesh.normals = nil
	mesh.uvs = nil
	mesh.hasOrco = hasOrco
	return
}

func (mesh *Mesh) Primitives() (prims []Primitive) {
	prims = make([]Primitive, len(mesh.triangles))
	for i, _ := range prims {
		prims[i] = mesh.triangles[i]
	}
	return
}

func (mesh *Mesh) Visible() bool     { return !mesh.hidden }
func (mesh *Mesh) SetVisible(v bool) { mesh.hidden = !v }

//func (mesh *Mesh) EvalVmap(sp surface.Point, id uint, val []float) int { return 0 }
func (mesh *Mesh) SetLight(l Light) { mesh.light = l }

//func (mesh *Mesh) EnableSampling() bool {}
//func (mesh *Mesh) Sample(s1, s2 float) (p, n vec64.Vector) {}

// SetData changes the mesh's data.
//
// For memory efficiency, the actual data for a mesh isn't stored in the
// triangles; the data is stored in the mesh.  The triangles simply contain
// indices that point to parts of the various arrays kept by the mesh.  Because
// most meshes have connected faces, this means that each vertex is stored once,
// instead of three times (much better!).
//
// Both normals and uvs are optional.  If you don't want to enable per-vertex
// normals or UV coordinates, then pass nil for the corresponding parameter.
// Any triangles that don't have per-vertex normals set will use the computed
// normal.
func (mesh *Mesh) SetData(vertices, normals []vec64.Vector, uvs []UV) {
	mesh.vertices, mesh.normals, mesh.uvs = vertices, normals, uvs
}

// AddTriangle adds a face to the mesh.
func (mesh *Mesh) AddTriangle(t *Triangle) {
	t.index = len(mesh.triangles)
	mesh.triangles = append(mesh.triangles, t)
}
