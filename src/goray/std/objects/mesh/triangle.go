//
//	goray/std/objects/mesh/triangle.go
//	goray
//
//	Created by Ross Light on 2010-12-26.
//

package mesh

import (
	"fmt"
	"math"
	"goray/logging"
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
	// Explanation: <http://softsurfer.com/Archive/algorithm_0105/algorithm_0105.htm#Segment-Triangle>
	coll.Ray = r
	a, b, c := tri.mesh.vertices[tri.va], tri.mesh.vertices[tri.vb], tri.mesh.vertices[tri.vc]
	rayDepth, u, v := intersect([3]float64(a), [3]float64(b), [3]float64(c), [3]float64(r.Dir), [3]float64(r.From))
	if rayDepth < 0 {
		return
	}

	coll.Primitive = tri
	coll.RayDepth = rayDepth
	coll.UserData = interface{}([2]float64{u, v})
	return
}

func (tri *Triangle) GetSurface(coll primitive.Collision) (sp surface.Point) {
	sp.GeometricNormal = tri.normal
	a, b, c := tri.getVertices()
	dat := coll.UserData.([2]float64)
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
		sp.U = u*uvA[0] + v*uvB[0] + w*uvC[0]
		sp.V = u*uvA[1] + v*uvB[1] + w*uvC[1]

		// Calculate world vectors
		du1, du2 := uvA[0]-uvC[0], uvB[0]-uvC[0]
		dv1, dv2 := uvA[1]-uvC[1], uvB[1]-uvC[1]
		det := du1*dv2 - dv1*du2

		if det != 0.0 {
			invdet := 1.0 / det
			dp1, dp2 := vector.Sub(a, c), vector.Sub(b, c)
			sp.WorldU = vector.Sub(vector.ScalarMul(dp1, dv2*invdet), vector.ScalarMul(dp2, dv1*invdet))
			sp.WorldV = vector.Sub(vector.ScalarMul(dp2, du1*invdet), vector.ScalarMul(dp1, du2*invdet))
		} else {
			sp.WorldU, sp.WorldV = vector.Vector3D{}, vector.Vector3D{}
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
	sp.ShadingU[vector.X] = vector.Dot(sp.NormalU, sp.WorldU)
	sp.ShadingU[vector.Y] = vector.Dot(sp.NormalV, sp.WorldU)
	sp.ShadingU[vector.Z] = vector.Dot(sp.Normal, sp.WorldU)
	sp.ShadingV[vector.X] = vector.Dot(sp.NormalU, sp.WorldV)
	sp.ShadingV[vector.Y] = vector.Dot(sp.NormalV, sp.WorldV)
	sp.ShadingV[vector.Z] = vector.Dot(sp.Normal, sp.WorldV)

	return
}

func (tri *Triangle) GetBound() *bound.Bound {
	var minPt, maxPt vector.Vector3D
	a, b, c := tri.getVertices()
	for axis := vector.X; axis <= vector.Z; axis++ {
		minPt[axis] = math.Fmin(math.Fmin(a[axis], b[axis]), c[axis])
		maxPt[axis] = math.Fmax(math.Fmax(a[axis], b[axis]), c[axis])
	}
	return bound.New(minPt, maxPt)
}

func (tri *Triangle) IntersectsBound(bd *bound.Bound) bool {
	var points [3][3]float64
	a, b, c := tri.getVertices()

	for axis := vector.X; axis <= vector.Z; axis++ {
		points[0][axis] = a[axis]
		points[1][axis] = b[axis]
		points[2][axis] = c[axis]
	}
	ctr := bd.GetCenter()
	return triBoxOverlap([3]float64(ctr), bd.GetHalfSize(), points)
}

func (tri *Triangle) GetMaterial() material.Material { return tri.material }

func (tri *Triangle) Clip(bound *bound.Bound, axis vector.Axis, lower bool, oldData interface{}) (clipped *bound.Bound, newData interface{}) {
	if axis >= 0 {
		return tri.clipPlane(bound, axis, lower, oldData)
	}
	return tri.clipBox(bound)
}

func (tri *Triangle) clipPlane(bound *bound.Bound, axis vector.Axis, lower bool, oldData interface{}) (clipped *bound.Bound, newData interface{}) {
	defer func() {
		if err := recover(); err != nil {
			logging.Debug(logging.MainLog, "Clip plane fault: %v", err)
			clipped, newData = tri.clipBox(bound)
		}
	}()

	var poly []vector.Vector3D
	if oldData == nil {
		a, b, c := tri.mesh.vertices[tri.va], tri.mesh.vertices[tri.vb], tri.mesh.vertices[tri.vc]
		poly = []vector.Vector3D{a, b, c, a}
	} else {
		poly = oldData.([]vector.Vector3D)
	}

	var split float64
	if lower {
		split = bound.GetMin()[axis]
	} else {
		split = bound.GetMax()[axis]
	}

	newData, clipped = triPlaneClip(axis, split, lower, poly)
	return
}

func (tri *Triangle) clipBox(bound *bound.Bound) (clipped *bound.Bound, newData interface{}) {
	defer func() {
		if err := recover(); err != nil {
			clipped, newData = nil, nil
			logging.Warning(logging.MainLog, "Clip panic: %v", err)
		}
	}()

	a, b, c := tri.mesh.vertices[tri.va], tri.mesh.vertices[tri.vb], tri.mesh.vertices[tri.vc]
	poly := []vector.Vector3D{a, b, c, a}
	newData, clipped = triBoxClip([3]float64(bound.GetMin()), [3]float64(bound.GetMax()), poly)
	return
}

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

func (tri *Triangle) GetSurfaceArea() float64 {
	a, b, c := tri.getVertices()
	edge1, edge2 := vector.Sub(b, a), vector.Sub(c, a)
	return vector.Cross(edge1, edge2).Length() * 0.5
}
