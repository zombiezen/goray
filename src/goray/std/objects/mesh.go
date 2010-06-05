//
//  goray/std/objects/mesh.go
//  goray
//
//  Created by Ross Light on 2010-06-04.
//

/* The goray/std/objects/mesh package provides mesh objects constructed from triangles. */
package mesh

import "fmt"
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
	mesh.vertices = nil
	mesh.normals = nil
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

func (mesh *Mesh) CanSample() bool { return false }

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

func (mesh *Mesh) SetData(vertices, normals []vector.Vector3D) {
	mesh.vertices, mesh.normals = vertices, normals
}

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

/* Triangle stores information for a single triangle. */
type Triangle struct {
	va, vb, vc int // va, vb, and vc are the vertex indices in the mesh's array.
	na, nb, nc int // na, nb, and nc are the normal indices in the mesh's array (if the face is smooth)
	index      int
	normal     vector.Vector3D
	material   material.Material
	mesh       *Mesh
}

/* NewTriangle creates a new triangle. */
func NewTriangle(a, b, c int, m *Mesh) (tri *Triangle) {
	tri = &Triangle{
		va: a, vb: b, vc: c,
		na: -1, nb: -1, nc: -1,
		index: -1,
		mesh:  m,
	}
	tri.CalculateNormal()
	return tri
}

func (tri *Triangle) String() string {
	a, b, c := tri.getVertices()
	return fmt.Sprintf("Triangle{%v %v %v N:%v}", a, b, c, tri.normal)
}

func (tri *Triangle) getVertices() (a, b, c vector.Vector3D) {
	return tri.mesh.vertices[tri.va], tri.mesh.vertices[tri.vb], tri.mesh.vertices[tri.vc]
}

func (tri *Triangle) getNormals() (a, b, c vector.Vector3D) {
	f := func(i int) vector.Vector3D {
		if i >= 0 && tri.mesh.normals != nil {
			return tri.mesh.normals[i]
		}
		return tri.normal
	}
	return f(tri.na), f(tri.nb), f(tri.nc)
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
	tvec := vector.Sub(r.From(), a)
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
	// The u and v in intersection code are actually v and w
	v, w := dat[0], dat[1]
	u := 1.0 - v - w

	if tri.mesh.normals != nil {
		na, nb, nc := tri.getNormals()
		sp.Normal = vector.Add(vector.ScalarMul(na, u), vector.ScalarMul(nb, v), vector.ScalarMul(nc, w)).Normalize()
	} else {
		sp.Normal = tri.normal
	}

	if tri.mesh.hasOrco {
		// TODO
	} else {
		sp.OrcoPosition = coll.GetPoint()
		sp.HasOrco = false
		sp.OrcoNormal = sp.GeometricNormal
	}
	// TODO if mesh.hasUV
	// else...
	{
		a, b, c := tri.getVertices()
		sp.U, sp.V = u, v
		sp.WorldU, sp.WorldV = vector.Sub(b, a), vector.Sub(c, a)
	}

	sp.Object = tri.mesh
	sp.Primitive = tri
	sp.Light = tri.mesh.light
	sp.Material = tri.material

	sp.SurfaceU, sp.SurfaceV = u, v
	sp.PrimitiveNumber = tri.index
	sp.Position = coll.GetPoint()

	sp.NormalU, sp.NormalV = vector.CreateCS(sp.Normal)
	sp.ShadingU.X = vector.Dot(sp.NormalU, sp.WorldU)
	sp.ShadingU.Y = vector.Dot(sp.NormalV, sp.WorldU)
	sp.ShadingU.Z = vector.Dot(sp.Normal, sp.WorldU)
	sp.ShadingV.X = vector.Dot(sp.NormalU, sp.WorldV)
	sp.ShadingV.Y = vector.Dot(sp.NormalV, sp.WorldV)
	sp.ShadingV.Z = vector.Dot(sp.Normal, sp.WorldV)

	return
}

func (tri *Triangle) GetBound() *bound.Bound {
	a, b, c := tri.getVertices()
	minPt := vector.New(fmath.Min(a.X, b.X, c.X), fmath.Min(a.Y, b.Y, c.Y), fmath.Min(a.Z, b.Z, c.Z))
	maxPt := vector.New(fmath.Max(a.X, b.X, c.X), fmath.Max(a.Y, b.Y, c.Y), fmath.Max(a.Z, b.Z, c.Z))
	return bound.New(minPt, maxPt)
}

func (tri *Triangle) IntersectsBound(bd *bound.Bound) bool {
	var points [3][3]float
	a, b, c := tri.getVertices()

	for i := 0; i < 3; i++ {
		points[0][i] = a.GetComponent(i)
		points[1][i] = b.GetComponent(i)
		points[2][i] = c.GetComponent(i)
	}
	ctr := bd.GetCenter()
	return triBoxOverlap([3]float{ctr.X, ctr.Y, ctr.Z}, bd.GetHalfSize(), points)
}

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

// Triangle bound intersection methods
// Note that a lot of functionality gets inlined for efficiency reasons.

func planeBoxOverlap(normal, vert, maxbox [3]float) bool {
	var vmin, vmax [3]float

	for q := 0; q < 3; q++ {
		v := vert[q]
		if normal[q] > 0.0 {
			vmin[q] = -maxbox[q] - v
			vmax[q] = maxbox[q] - v
		} else {
			vmin[q] = maxbox[q] - v
			vmax[q] = -maxbox[q] - v
		}
	}

	if normal[0]*vmin[0]+normal[1]*vmin[1]+normal[2]*vmin[2] > 0 {
		return false
	}
	if normal[0]*vmax[0]+normal[1]*vmax[1]+normal[2]*vmax[2] >= 0 {
		return true
	}

	return false
}

func triBoxOverlap(boxcenter, boxhalfsize [3]float, verts [3][3]float) bool {
	const X = 0
	const Y = 1
	const Z = 2
	var normal, f [3]float
	var v, e [3][3]float
	var min, max float

	axisTest := func(axis, i, j, edgeNum int) bool {
		var axis1, axis2 int
		var a, b, fa, fb float
		var p1, p2, rad, min, max float
		var v1, v2 []float

		switch axis {
		case X:
			axis1, axis2 = Y, Z
		case Y:
			axis1, axis2 = X, Z
		case Z:
			axis1, axis2 = X, Y
		}

		v1, v2 = &v[i], &v[j]
		a, b = e[edgeNum][axis2], e[edgeNum][axis1]
		fa, fb = f[axis2], f[axis1]

		p1 = -a*v1[axis1] + b*v1[axis2]
		p2 = -a*v2[axis1] + b*v2[axis2]
		if p1 < p2 {
			min, max = p1, p2
		} else {
			min, max = p2, p1
		}
		rad = fa*boxhalfsize[axis1] + fb*boxhalfsize[axis2]
		return min <= rad && max >= -rad
	}

	// Move everything so that the boxcenter is in (0, 0, 0)
	v[0][X] = verts[0][X] - boxcenter[X]
	v[0][Y] = verts[0][Y] - boxcenter[Y]
	v[0][Z] = verts[0][Z] - boxcenter[Z]

	v[1][X] = verts[1][X] - boxcenter[X]
	v[1][Y] = verts[1][Y] - boxcenter[Y]
	v[1][Z] = verts[1][Z] - boxcenter[Z]

	v[2][X] = verts[2][X] - boxcenter[X]
	v[2][Y] = verts[2][Y] - boxcenter[Y]
	v[2][Z] = verts[2][Z] - boxcenter[Z]

	// Compute triangle edges
	e[0][X] = v[1][X] - v[0][X]
	e[0][Y] = v[1][Y] - v[0][Y]
	e[0][Z] = v[1][Z] - v[0][Z]

	e[1][X] = v[2][X] - v[1][X]
	e[1][Y] = v[2][Y] - v[1][Y]
	e[1][Z] = v[2][Z] - v[1][Z]

	e[2][X] = v[0][X] - v[2][X]
	e[2][Y] = v[0][Y] - v[2][Y]
	e[2][Z] = v[0][Z] - v[2][Z]

	// Run the nine tests
	f = [3]float{fmath.Abs(e[0][X]), fmath.Abs(e[0][Y]), fmath.Abs(e[0][Z])}
	if !axisTest(X, 0, 1, 0) {
		return false
	}
	if !axisTest(Y, 0, 2, 0) {
		return false
	}
	if !axisTest(Z, 1, 2, 0) {
		return false
	}

	f = [3]float{fmath.Abs(e[1][X]), fmath.Abs(e[1][Y]), fmath.Abs(e[1][Z])}
	if !axisTest(X, 0, 1, 1) {
		return false
	}
	if !axisTest(Y, 0, 2, 1) {
		return false
	}
	if !axisTest(Z, 0, 1, 1) {
		return false
	}

	f = [3]float{fmath.Abs(e[2][X]), fmath.Abs(e[2][Y]), fmath.Abs(e[2][Z])}
	if !axisTest(X, 0, 2, 2) {
		return false
	}
	if !axisTest(Y, 0, 1, 2) {
		return false
	}
	if !axisTest(Z, 1, 2, 2) {
		return false
	}

	// First test overlap in the x,y,z directions
	// This is equivalent to testing a minimal AABB
	min, max = v[0][X], v[0][X]
	if v[1][X] < min {
		min = v[1][X]
	}
	if v[1][X] > max {
		max = v[1][X]
	}
	if v[2][X] < min {
		min = v[2][X]
	}
	if v[2][X] > max {
		max = v[2][X]
	}
	if min > boxhalfsize[X] || max < -boxhalfsize[X] {
		return false
	}

	min, max = v[0][Y], v[0][Y]
	if v[1][Y] < min {
		min = v[1][Y]
	}
	if v[1][Y] > max {
		max = v[1][Y]
	}
	if v[2][Y] < min {
		min = v[2][Y]
	}
	if v[2][Y] > max {
		max = v[2][Y]
	}
	if min > boxhalfsize[Y] || max < -boxhalfsize[Y] {
		return false
	}

	min, max = v[0][Z], v[0][Z]
	if v[1][Z] < min {
		min = v[1][Z]
	}
	if v[1][Z] > max {
		max = v[1][Z]
	}
	if v[2][Z] < min {
		min = v[2][Z]
	}
	if v[2][Z] > max {
		max = v[2][Z]
	}
	if min > boxhalfsize[Z] || max < -boxhalfsize[Z] {
		return false
	}

	// Test if the box intersects the plane of the triangle
	// Plane equation of triangle: normal * x + d = 0
	normal[0] = e[0][1]*e[1][2] - e[0][2]*e[1][1]
	normal[1] = e[0][2]*e[1][0] - e[0][0]*e[1][2]
	normal[2] = e[0][0]*e[1][1] - e[0][1]*e[1][0]
	return planeBoxOverlap(normal, v[0], boxhalfsize)
}
