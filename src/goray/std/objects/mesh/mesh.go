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
	"goray/core/object"
	"goray/core/primitive"
	"goray/core/vector"
	yamldata "goyaml.googlecode.com/hg/data"
)

// UV holds a set of texture coordinates.
type UV [2]float64

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

var _ object.Object3D = &Mesh{}

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

func (mesh *Mesh) Primitives() (prims []primitive.Primitive) {
	prims = make([]primitive.Primitive, len(mesh.triangles))
	for i, _ := range prims {
		prims[i] = mesh.triangles[i]
	}
	return
}

func (mesh *Mesh) Visible() bool     { return !mesh.hidden }
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
	t.index = len(mesh.triangles)
	mesh.triangles = append(mesh.triangles, t)
}

func Construct(m yamldata.Map) (data interface{}, err os.Error) {
	m = m.Copy()
	m.SetDefault("vertices", []interface{}{})
	m.SetDefault("uvs", []interface{}{})
	m.SetDefault("faces", []interface{}{})

	vertices, ok := yamldata.AsSequence(m["vertices"])
	if !ok {
		err = os.NewError("Vertices must be a sequence")
		return
	}

	uvs, ok := yamldata.AsSequence(m["uvs"])
	if !ok {
		err = os.NewError("UVs must be a sequence")
		return
	}

	faces, ok := yamldata.AsSequence(m["faces"])
	if !ok {
		err = os.NewError("Faces must be a sequence")
		return
	}

	mesh := New(len(faces), false)

	var vertexData []vector.Vector3D
	var uvData []UV
	// Parse vertices
	// TODO: Error handling
	vertexData = make([]vector.Vector3D, len(vertices))
	for i, _ := range vertices {
		vseq, _ := yamldata.AsSequence(vertices[i])
		x, _ := yamldata.AsFloat(vseq[0])
		y, _ := yamldata.AsFloat(vseq[1])
		z, _ := yamldata.AsFloat(vseq[2])
		vertexData[i] = vector.Vector3D{x, y, z}
	}
	// Parse UVs
	// TODO: Error handling
	if len(uvs) > 0 {
		uvData = make([]UV, len(uvs))
		for i, _ := range uvs {
			uvseq, _ := yamldata.AsSequence(uvs[i])
			u, _ := yamldata.AsFloat(uvseq[0])
			v, _ := yamldata.AsFloat(uvseq[1])
			uvData[i] = UV{u, v}
		}
	}
	mesh.SetData(vertexData, nil, uvData)

	// Parse faces
	// TODO: Error handling
	for i, _ := range faces {
		fmap, _ := yamldata.AsMap(faces[i])
		// Vertices
		vindices, _ := yamldata.AsSequence(fmap["vertices"])
		va, _ := yamldata.AsInt(vindices[0])
		vb, _ := yamldata.AsInt(vindices[1])
		vc, _ := yamldata.AsInt(vindices[2])
		// UVs
		var uva, uvb, uvc int64 = -1, -1, -1
		if _, hasUVs := fmap["uvs"]; len(uvs) > 0 && hasUVs {
			uvindices, _ := yamldata.AsSequence(fmap["uvs"])
			uva, _ = yamldata.AsInt(uvindices[0])
			uvb, _ = yamldata.AsInt(uvindices[1])
			uvc, _ = yamldata.AsInt(uvindices[2])
		}
		// Create triangle
		tri := NewTriangle(int(va), int(vb), int(vc), mesh)
		tri.SetUVs(int(uva), int(uvb), int(uvc))
		tri.SetMaterial(fmap["material"].(material.Material))
		mesh.AddTriangle(tri)
	}

	return mesh, nil
}
