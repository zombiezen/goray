//
//	goray/std/objects/mesh/mesh.go
//	goray
//
//	Created by Ross Light on 2010-06-04.
//

/*
	The mesh package provides mesh objects constructed from triangles.

	The basic workflow for making a working mesh is: create the mesh, set the mesh's data, then add the triangles.
*/
package mesh

import (
	"os"
	"goray/core/light"
	"goray/core/material"
	"goray/core/matrix"
	"goray/core/primitive"
	"goray/core/vector"
	yamldata "yaml/data"
)

// UV holds a set of texture coordinates.
type UV struct {
	U, V float
}

// A Mesh is a collection of triangles.
type Mesh struct {
	triangles []*Triangle
	vertices  []vector.Vector3D
	normals   []vector.Vector3D
	uvs       []UV
	hasOrco   bool
	light     light.Light
	world2obj *matrix.Matrix
	hidden    bool
}

// New creates an empty mesh.
func New(ntris int, hasOrco bool) (mesh *Mesh) {
	mesh = new(Mesh)
	mesh.triangles = make([]*Triangle, 0, ntris)
	mesh.vertices = nil
	mesh.normals = nil
	mesh.uvs = nil
	mesh.hasOrco = hasOrco
	return
}

func (mesh *Mesh) GetPrimitives() (prims []primitive.Primitive) {
	prims = make([]primitive.Primitive, len(mesh.triangles))
	for i, _ := range prims {
		prims[i] = mesh.triangles[i]
	}
	return
}

func (mesh *Mesh) IsVisible() bool   { return !mesh.hidden }
func (mesh *Mesh) SetVisible(v bool) { mesh.hidden = !v }

//func (mesh *Mesh) EvalVmap(sp surface.Point, id uint, val []float) int { return 0 }
func (mesh *Mesh) SetLight(l light.Light) { mesh.light = l }

//func (mesh *Mesh) EnableSampling() bool {}
//func (mesh *Mesh) Sample(s1, s2 float) (p, n vector.Vector3D) {}

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
func (mesh *Mesh) SetData(vertices, normals []vector.Vector3D, uvs []UV) {
	mesh.vertices, mesh.normals, mesh.uvs = vertices, normals, uvs
}

// AddTriangle adds a face to the mesh.
func (mesh *Mesh) AddTriangle(t *Triangle) {
	if len(mesh.triangles)+1 > cap(mesh.triangles) {
		newTris := make([]*Triangle, len(mesh.triangles), cap(mesh.triangles)*2)
		copy(newTris, mesh.triangles)
		mesh.triangles = newTris
	}
	t.index = len(mesh.triangles)
	mesh.triangles = mesh.triangles[0 : len(mesh.triangles)+1]
	mesh.triangles[t.index] = t
}

func Construct(m yamldata.Map) (data interface{}, err os.Error) {
	m = m.Copy()
	m.SetDefault("vertices", make(yamldata.Sequence, 0))
	m.SetDefault("faces", make(yamldata.Sequence, 0))

	vertices, ok := yamldata.AsSequence(m["vertices"])
	if !ok {
		err = os.NewError("Vertices must be a sequence")
		return
	}

	faces, ok := yamldata.AsSequence(m["faces"])
	if !ok {
		err = os.NewError("Faces must be a sequence")
		return
	}

	mesh := New(len(faces), false)

	// Parse vertices
	// TODO: Error handling
	vertexData := make([]vector.Vector3D, len(vertices))
	for i, _ := range vertices {
		vseq, _ := yamldata.AsSequence(vertices[i])
		x, _ := yamldata.AsFloat(vseq[0])
		y, _ := yamldata.AsFloat(vseq[1])
		z, _ := yamldata.AsFloat(vseq[2])
		vertexData[i] = vector.New(float(x), float(y), float(z))
	}
	mesh.SetData(vertexData, nil, nil)

	// Parse faces
	// TODO: Error handling
	for i, _ := range faces {
		fmap, _ := yamldata.AsMap(faces[i])
		vindices, _ := yamldata.AsSequence(fmap["vertices"])
		a, _ := yamldata.AsInt(vindices[0])
		b, _ := yamldata.AsInt(vindices[1])
		c, _ := yamldata.AsInt(vindices[2])
		tri := NewTriangle(int(a), int(b), int(c), mesh)
		tri.SetMaterial(fmap["material"].(material.Material))
		mesh.AddTriangle(tri)
	}

	return mesh, nil
}