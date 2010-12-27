//
//	goray/std/objects/mesh/triangle.go
//	goray
//
//	Created by Ross Light on 2010-12-26.
//

package mesh

import (
	"fmt"
	"goray/logging"
	"goray/fmath"
	"goray/core/bound"
	"goray/core/material"
	"goray/core/primitive"
	"goray/core/ray"
	"goray/core/surface"
	"goray/core/vector"
)

// Triangle stores information for a single triangle.
type Triangle struct {
	va, vb, vc    int // va, vb, and vc are the vertex indices in the mesh's array.
	na, nb, nc    int // na, nb, and nc are the normal indices in the mesh's array (if per-vertex normals are enabled).
	uva, uvb, uvc int // uva, uvb, and uvc are the UV indices in the mesh's array (if UV is enabled).
	index         int
	normal        vector.Vector3D
	material      material.Material
	mesh          *Mesh
}

// NewTriangle creates a new triangle.
func NewTriangle(a, b, c int, m *Mesh) (tri *Triangle) {
	tri = &Triangle{
		va: a, vb: b, vc: c,
		na: -1, nb: -1, nc: -1,
		uva: -1, uvb: -1, uvc: -1,
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

func (tri *Triangle) getUVs() (a, b, c UV) {
	f := func(i int) UV {
		if i >= 0 && tri.mesh.uvs != nil {
			return tri.mesh.uvs[i]
		}
		return UV{}
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
	a, b, c := tri.getVertices()
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

	sp.HasOrco = tri.mesh.hasOrco
	if tri.mesh.hasOrco {
		// TODO: Yafaray uses index+1 for each one of the vertices. Why?
		sp.OrcoPosition = vector.Add(vector.ScalarMul(a, u), vector.ScalarMul(b, v), vector.ScalarMul(c, w))
		sp.OrcoNormal = vector.Cross(vector.Sub(b, a), vector.Sub(c, a)).Normalize()
	} else {
		sp.OrcoPosition = coll.GetPoint()
		sp.OrcoNormal = sp.GeometricNormal
	}

	if tri.mesh.uvs != nil {
		// u, v, and w are actually the barycentric coords, not some UVs.
		uvA, uvB, uvC := tri.getUVs()
		sp.U = u*uvA.U + v*uvB.U + w*uvC.U
		sp.V = u*uvA.V + v*uvB.V + w*uvC.V

		// Calculate world vectors
		du1, du2 := uvA.U-uvC.U, uvB.U-uvC.U
		dv1, dv2 := uvA.V-uvC.V, uvB.V-uvC.V
		det := du1*dv2 - dv1*du2

		if !fmath.Eq(det, 0.0) {
			invdet := 1.0 / det
			dp1, dp2 := vector.Sub(a, c), vector.Sub(b, c)
			sp.WorldU = vector.Sub(vector.ScalarMul(dp1, dv2*invdet), vector.ScalarMul(dp2, dv1*invdet))
			sp.WorldV = vector.Sub(vector.ScalarMul(dp2, du1*invdet), vector.ScalarMul(dp1, du2*invdet))
		} else {
			sp.WorldU, sp.WorldV = vector.New(0, 0, 0), vector.New(0, 0, 0)
		}
	} else {
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

//func (tri *Triangle) ClipToBound(bound [2][3]float, axis int) (clipped *bound.Bound, ok bool) { return }

func (tri *Triangle) GetMaterial() material.Material { return tri.material }

// The rest of these are non-interface triangle-specific methods.

func (tri *Triangle) SetMaterial(mat material.Material) { tri.material = mat }
func (tri *Triangle) SetNormals(a, b, c int)            { tri.na, tri.nb, tri.nc = a, b, c }
func (tri *Triangle) ClearNormals()                     { tri.na, tri.nb, tri.nc = -1, -1, -1 }
func (tri *Triangle) SetUVs(a, b, c int)                { tri.uva, tri.uvb, tri.uvc = a, b, c }
func (tri *Triangle) ClearUVs()                         { tri.uva, tri.uvb, tri.uvc = -1, -1, -1 }
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

func (tri *Triangle) ClipToBound(bound *bound.Bound, axis int, oldData interface{}) (clipped *bound.Bound, newData interface{}) {
	defer func() {
		if err := recover(); err != nil {
			clipped, newData = nil, nil
			logging.Warning(logging.MainLog, "ClipToBound panic: %v", err)
		}
	}()

	var poly []dVector
	if oldData == nil {
		a, b, c := tri.mesh.vertices[tri.va], tri.mesh.vertices[tri.vb], tri.mesh.vertices[tri.vc]
		poly = []dVector{vec2dvec(a), vec2dvec(b), vec2dvec(c), vec2dvec(a)}
	} else {
		poly = oldData.([]dVector)
	}

	if axis >= 0 {
		lower := (axis &^ 3) != 0
		axis = axis & 3

		var split float64
		if lower {
			split = float64(bound.GetMin().GetComponent(axis))
		} else {
			split = float64(bound.GetMax().GetComponent(axis))
		}

		newData, clipped = triPlaneClip(axis, split, lower, poly)
		return
	}

	// Full clip
	bMin := [3]float64(vec2dvec(bound.GetMin()))
	bMax := [3]float64(vec2dvec(bound.GetMax()))
	newData, clipped = triBoxClip(bMin, bMax, poly)
	return
}
