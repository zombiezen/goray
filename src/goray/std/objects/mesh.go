//
//  goray/std/objects/mesh.go
//  goray
//
//  Created by Ross Light on 2010-06-04.
//

/* The goray/std/objects/mesh package provides mesh objects constructed from triangles. */
package mesh

import "./fmath"

import (
	"./goray/bound"
	"./goray/light"
	"./goray/material"
	"./goray/matrix"
	"./goray/primitive"
	"./goray/ray"
	"./goray/surface"
	"./goray/vector"
)

/* A Mesh is a collection of triangles. */
type Mesh struct {
	triangles []*Triangle
	vertices  []vector.Vector3D
	normals   []vector.Vector3D
	hasOrco   bool
	light     light.Light
	world2obj *matrix.Matrix
	hidden    bool
}

/* New creates an empty mesh. */
func New(ntris int) (mesh *Mesh) {
	mesh = new(Mesh)
	mesh.triangles = make([]*Triangle, 0, ntris)
	mesh.vertices = make([]vector.Vector3D, 0, ntris*3)
	mesh.normals = make([]vector.Vector3D, 0)
	return
}

func (mesh *Mesh) GetPrimitives() (prims []primitive.Primitive) {
	prims = make([]primitive.Primitive, len(mesh.triangles))
	for i, _ := range prims {
		prims[i] = mesh.triangles[i]
	}
	return
}

func (mesh *Mesh) EvalVmap(sp surface.Point, id uint, val []float) int { return 0 }
func (mesh *Mesh) SetLight(l light.Light)                              { mesh.light = l }

func (mesh *Mesh) EnableSampling() bool {
	// TODO
	return false
}

func (mesh *Mesh) Sample(s1, s2 float) (p, n vector.Vector3D) {
	// TODO
	return
}

func (mesh *Mesh) IsVisible() bool   { return !mesh.hidden }
func (mesh *Mesh) SetVisible(v bool) { mesh.hidden = !v }

func (mesh *Mesh) AddTriangle(t *Triangle) {
	if len(mesh.triangles)+1 > cap(mesh.triangles) {
		newTris := make([]*Triangle, len(mesh.triangles), cap(mesh.triangles)*2)
		copy(newTris, mesh.triangles)
		mesh.triangles = newTris
	}
	mesh.triangles = mesh.triangles[0 : len(mesh.triangles)+1]
	mesh.triangles[len(mesh.triangles)-1] = t
}

/* Triangle stores information for a single triangle. */
type Triangle struct {
	va, vb, vc int // va, vb, and vc are the vertex indices in the mesh's array.
	na, nb, nc int // na, nb, and nc are the normal indices in the mesh's array (if the face is smooth)
	normal     vector.Vector3D
	material   material.Material
	mesh       *Mesh
}

/* NewTriangle creates a new triangle. */
func NewTriangle(a, b, c int, m *Mesh) *Triangle {
	return &Triangle{
		va: a, vb: b, vc: c,
		na: -1, nb: -1, nc: -1,
		mesh: m,
	}
}

func (tri *Triangle) getVertices() (a, b, c vector.Vector3D) {
	return tri.mesh.vertices[tri.va], tri.mesh.vertices[tri.vb], tri.mesh.vertices[tri.vc]
}

func (tri *Triangle) Intersect(r ray.Ray) (coll primitive.Collision) {
	// Tomas Möller and Ben Trumbore ray intersection scheme
	// Ross adds: This is based on an ACM white paper which I don't have access to.
	// I'm just going to smile, nod, and copy the code.  Code monkey very diligent.
    coll.Ray = r
	a, b, c := tri.getVertices()

	edge1, edge2 := vector.Sub(b, a), vector.Sub(c, a)
	pvec := vector.Cross(r.Dir(), edge2)
	det := vector.Dot(edge1, pvec)
	if fmath.Eq(det, 0.0) {
		return
	}
	invDet := 1.0 / det
	tvec := vector.Cross(r.From(), a)
	u := vector.Dot(tvec, pvec) * invDet
	if u < 0.0 || u > 1.0 {
		return
	}
	qvec := vector.Cross(tvec, edge1)
	v := vector.Dot(r.Dir(), qvec) * invDet
	if v < 0.0 || u+v > 1.0 {
		return
	}

    coll.Primitive = tri
	coll.RayDepth = vector.Dot(edge2, qvec) * invDet
    coll.UserData = interface{}([2]float{u, v})
	return
}

func (tri *Triangle) GetSurface(coll primitive.Collision) (sp surface.Point) {
	sp.GeometricNormal = tri.normal
    dat := coll.UserData.([2]float)
    _ = dat
	// TODO: This is gonna be ugly.
	return
}

func (tri *Triangle) GetBound() *bound.Bound {
	a, b, c := tri.getVertices()
	minPt := vector.New(fmath.Min(a.X, b.X, c.X), fmath.Min(a.Y, b.Y, c.Y), fmath.Min(a.Z, b.Z, c.Z))
	maxPt := vector.New(fmath.Max(a.X, b.X, c.X), fmath.Max(a.Y, b.Y, c.Y), fmath.Max(a.Z, b.Z, c.Z))
	return bound.New(minPt, maxPt)
}

func (tri *Triangle) IntersectsBound(*bound.Bound) bool                                       { return true }
func (tri *Triangle) HasClippingSupport() bool                                                { return false }
func (tri *Triangle) ClipToBound(bound [2][3]float, axis int) (clipped *bound.Bound, ok bool) { return }

func (tri *Triangle) GetMaterial() material.Material { return tri.material }

// The rest of these are non-interface triangle-specific methods.

func (tri *Triangle) SetMaterial(mat material.Material) { tri.material = mat }
func (tri *Triangle) SetNormals(a, b, c int)            { tri.na, tri.nb, tri.nc = a, b, c }
func (tri *Triangle) ClearNormals()                     { tri.na, tri.nb, tri.nc = -1, -1, -1 }
func (tri *Triangle) GetNormal() vector.Vector3D        { return tri.normal }

func (tri *Triangle) CalculateNormal() {
	a, b, c := tri.getVertices()
	tri.normal = vector.Cross(vector.Sub(b, a), vector.Sub(c, a)).Normalize()
}

func (tri *Triangle) GetSurfaceArea() float {
	a, b, c := tri.getVertices()
	edge1, edge2 := vector.Sub(b, a), vector.Sub(c, a)
	return vector.Cross(edge1, edge2).Length() * 0.5
}
