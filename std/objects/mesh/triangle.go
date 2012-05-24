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

package mesh

import (
	"fmt"
	"math"

	"bitbucket.org/zombiezen/goray"
	"bitbucket.org/zombiezen/goray/bound"
	"bitbucket.org/zombiezen/goray/logging"
	"bitbucket.org/zombiezen/goray/vector"
)

// Triangle stores information for a single triangle.
type Triangle struct {
	v     [3]int // v containes the vertex indices in the mesh's array.
	n     [3]int // n contains the normal indices in the mesh's array (if per-vertex normals are enabled).
	uv    [3]int // uv contains the UV indices in the mesh's array (if UV is enabled).
	index int

	normal   vector.Vector3D
	material goray.Material
	mesh     *Mesh
}

var _ goray.Primitive = &Triangle{}

// NewTriangle creates a new triangle.
func NewTriangle(a, b, c int, m *Mesh) (tri *Triangle) {
	tri = &Triangle{
		v:     [3]int{a, b, c},
		n:     [3]int{-1, -1, -1},
		uv:    [3]int{-1, -1, -1},
		index: -1,
		mesh:  m,
	}
	tri.CalculateNormal()
	return tri
}

func (tri *Triangle) String() string {
	v := tri.getVertices()
	return fmt.Sprintf("Triangle{%v %v %v N:%v}", v[0], v[1], v[2], tri.normal)
}

func (tri *Triangle) getVertices() (v [3]vector.Vector3D) {
	for i := 0; i < 3; i++ {
		v[i] = tri.mesh.vertices[tri.v[i]]
	}
	return
}

func (tri *Triangle) getNormals() (n [3]vector.Vector3D) {
	for i := 0; i < 3; i++ {
		if tri.n[i] >= 0 && tri.mesh.normals != nil {
			n[i] = tri.mesh.normals[tri.n[i]]
		}
	}
	return
}

func (tri *Triangle) getUVs() (uv [3]UV) {
	for i := 0; i < 3; i++ {
		if tri.uv[i] >= 0 && tri.mesh.uvs != nil {
			uv[i] = tri.mesh.uvs[tri.uv[i]]
		}
	}
	return
}

func (tri *Triangle) Intersect(r goray.Ray) (coll goray.Collision) {
	coll.Ray = r
	rayDepth, u, v := intersect(
		[3]float64(tri.mesh.vertices[tri.v[0]]),
		[3]float64(tri.mesh.vertices[tri.v[1]]),
		[3]float64(tri.mesh.vertices[tri.v[2]]),
		[3]float64(r.Dir),
		[3]float64(r.From),
	)
	if rayDepth < 0 {
		return
	}

	coll.Primitive = tri
	coll.RayDepth = rayDepth
	coll.UserData = interface{}([2]float64{u, v})
	return
}

func (tri *Triangle) Surface(coll goray.Collision) (sp goray.SurfacePoint) {
	sp.GeometricNormal = tri.normal
	vert := tri.getVertices()

	// The u and v in intersection code are actually v and w
	dat := coll.UserData.([2]float64)
	v, w := dat[0], dat[1]
	u := 1.0 - v - w

	if tri.mesh.normals != nil {
		n := tri.getNormals()
		sp.Normal = vector.Add(vector.ScalarMul(n[0], u), vector.ScalarMul(n[1], v), vector.ScalarMul(n[2], w)).Normalize()
	} else {
		sp.Normal = tri.normal
	}

	sp.HasOrco = tri.mesh.hasOrco
	if tri.mesh.hasOrco {
		// TODO: Yafaray uses index+1 for each one of the vertices. Why?
		sp.OrcoPosition = vector.Add(vector.ScalarMul(vert[0], u), vector.ScalarMul(vert[1], v), vector.ScalarMul(vert[2], w))
		sp.OrcoNormal = vector.Cross(vector.Sub(vert[1], vert[0]), vector.Sub(vert[2], vert[0])).Normalize()
	} else {
		sp.OrcoPosition = coll.Point()
		sp.OrcoNormal = sp.GeometricNormal
	}

	if tri.mesh.uvs != nil {
		// u, v, and w are actually the barycentric coords, not some UVs.
		uv := tri.getUVs()
		sp.U = u*uv[0][0] + v*uv[1][0] + w*uv[2][0]
		sp.V = u*uv[0][1] + v*uv[1][1] + w*uv[2][1]

		// Calculate world vectors
		du1, du2 := uv[0][0]-uv[2][0], uv[1][0]-uv[2][0]
		dv1, dv2 := uv[0][1]-uv[2][1], uv[1][1]-uv[2][1]
		det := du1*dv2 - dv1*du2

		if det != 0.0 {
			invdet := 1.0 / det
			dp1, dp2 := vector.Sub(vert[0], vert[2]), vector.Sub(vert[1], vert[2])
			sp.WorldU = vector.Sub(vector.ScalarMul(dp1, dv2*invdet), vector.ScalarMul(dp2, dv1*invdet))
			sp.WorldV = vector.Sub(vector.ScalarMul(dp2, du1*invdet), vector.ScalarMul(dp1, du2*invdet))
		} else {
			sp.WorldU, sp.WorldV = vector.Vector3D{}, vector.Vector3D{}
		}
	} else {
		sp.U, sp.V = u, v
		sp.WorldU, sp.WorldV = vector.Sub(vert[1], vert[0]), vector.Sub(vert[2], vert[0])
	}

	sp.Object = tri.mesh
	sp.Primitive = tri
	sp.Light = tri.mesh.light
	sp.Material = tri.material

	sp.SurfaceU, sp.SurfaceV = u, v
	sp.PrimitiveNumber = tri.index
	sp.Position = coll.Point()

	sp.NormalU, sp.NormalV = vector.CreateCS(sp.Normal)
	sp.ShadingU[vector.X] = vector.Dot(sp.NormalU, sp.WorldU)
	sp.ShadingU[vector.Y] = vector.Dot(sp.NormalV, sp.WorldU)
	sp.ShadingU[vector.Z] = vector.Dot(sp.Normal, sp.WorldU)
	sp.ShadingV[vector.X] = vector.Dot(sp.NormalU, sp.WorldV)
	sp.ShadingV[vector.Y] = vector.Dot(sp.NormalV, sp.WorldV)
	sp.ShadingV[vector.Z] = vector.Dot(sp.Normal, sp.WorldV)

	return
}

func (tri *Triangle) Bound() bound.Bound {
	var minPt, maxPt vector.Vector3D
	v := tri.getVertices()
	for axis := vector.X; axis <= vector.Z; axis++ {
		minPt[axis] = math.Min(math.Min(v[0][axis], v[1][axis]), v[2][axis])
		maxPt[axis] = math.Max(math.Max(v[0][axis], v[1][axis]), v[2][axis])
	}
	return bound.Bound{minPt, maxPt}
}

func (tri *Triangle) IntersectsBound(bd bound.Bound) bool {
	var points [3][3]float64
	vert := tri.getVertices()
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			points[i][j] = vert[i][j]
		}
	}

	ctr := bd.Center()
	return triBoxOverlap([3]float64(ctr), bd.HalfSize(), points)
}

func (tri *Triangle) Material() goray.Material { return tri.material }

func (tri *Triangle) Clip(bound bound.Bound, axis vector.Axis, lower bool, oldData interface{}) (clipped bound.Bound, newData interface{}) {
	if axis >= 0 {
		return tri.clipPlane(bound, axis, lower, oldData)
	}
	return tri.clipBox(bound)
}

func (tri *Triangle) clipPlane(bound bound.Bound, axis vector.Axis, lower bool, oldData interface{}) (clipped bound.Bound, newData interface{}) {
	defer func() {
		if err := recover(); err != nil {
			logging.Debug(logging.MainLog, "Clip plane fault: %v", err)
			clipped, newData = tri.clipBox(bound)
		}
	}()

	var poly []vector.Vector3D
	if oldData == nil {
		v := tri.getVertices()
		poly = []vector.Vector3D{v[0], v[1], v[2], v[0]}
	} else {
		poly = oldData.([]vector.Vector3D)
	}

	var split float64
	if lower {
		split = bound.Min[axis]
	} else {
		split = bound.Max[axis]
	}

	newData, clipped = triPlaneClip(axis, split, lower, poly)
	return
}

func (tri *Triangle) clipBox(bd bound.Bound) (clipped bound.Bound, newData interface{}) {
	defer func() {
		if err := recover(); err != nil {
			clipped, newData = bound.Bound{}, nil
			logging.Warning(logging.MainLog, "Clip panic: %v", err)
		}
	}()

	v := tri.getVertices()
	poly := []vector.Vector3D{v[0], v[1], v[2], v[0]}
	newData, clipped = triBoxClip([3]float64(bd.Min), [3]float64(bd.Max), poly)
	return
}

// The rest of these are non-interface triangle-specific methods.

func (tri *Triangle) SetMaterial(mat goray.Material) { tri.material = mat }
func (tri *Triangle) SetNormals(a, b, c int)         { tri.n = [3]int{a, b, c} }
func (tri *Triangle) ClearNormals()                  { tri.n = [3]int{-1, -1, -1} }
func (tri *Triangle) SetUVs(a, b, c int)             { tri.uv = [3]int{a, b, c} }
func (tri *Triangle) ClearUVs()                      { tri.uv = [3]int{-1, -1, -1} }
func (tri *Triangle) GetNormal() vector.Vector3D     { return tri.normal }

func (tri *Triangle) CalculateNormal() {
	v := tri.getVertices()
	tri.normal = vector.Cross(vector.Sub(v[1], v[0]), vector.Sub(v[2], v[0])).Normalize()
}

func (tri *Triangle) SurfaceArea() float64 {
	v := tri.getVertices()
	edge1, edge2 := vector.Sub(v[1], v[0]), vector.Sub(v[2], v[0])
	return vector.Cross(edge1, edge2).Length() * 0.5
}
