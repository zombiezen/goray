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
	"fmt"
	"math"

	"bitbucket.org/zombiezen/goray/bound"
	"bitbucket.org/zombiezen/goray/log"
	"bitbucket.org/zombiezen/goray/vecutil"
	"bitbucket.org/zombiezen/math3/vec64"
)

// Triangle stores information for a single triangle.
type Triangle struct {
	v     [3]int // v containes the vertex indices in the mesh's array.
	n     [3]int // n contains the normal indices in the mesh's array (if per-vertex normals are enabled).
	uv    [3]int // uv contains the UV indices in the mesh's array (if UV is enabled).
	index int

	normal   vec64.Vector
	material Material
	mesh     *Mesh
}

var _ Primitive = &Triangle{}

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

func (tri *Triangle) getVertices() (v [3]vec64.Vector) {
	for i := 0; i < 3; i++ {
		v[i] = tri.mesh.vertices[tri.v[i]]
	}
	return
}

func (tri *Triangle) getNormals() (n [3]vec64.Vector) {
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

func (tri *Triangle) Intersect(r Ray) (coll Collision) {
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

func (tri *Triangle) Surface(coll Collision) (sp SurfacePoint) {
	sp.GeometricNormal = tri.normal
	vert := tri.getVertices()

	// The u and v in intersection code are actually v and w
	dat := coll.UserData.([2]float64)
	v, w := dat[0], dat[1]
	u := 1.0 - v - w

	if tri.mesh.normals != nil {
		n := tri.getNormals()
		sp.Normal = vec64.Sum(n[0].Scale(u), n[1].Scale(v), n[2].Scale(w)).Normalize()
	} else {
		sp.Normal = tri.normal
	}

	sp.HasOrco = tri.mesh.hasOrco
	if tri.mesh.hasOrco {
		// TODO: Yafaray uses index+1 for each one of the vertices. Why?
		sp.OrcoPosition = vec64.Sum(vert[0].Scale(u), vert[1].Scale(v), vert[2].Scale(w))
		sp.OrcoNormal = vec64.Cross(vec64.Sub(vert[1], vert[0]), vec64.Sub(vert[2], vert[0])).Normalize()
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
			dp1, dp2 := vec64.Sub(vert[0], vert[2]), vec64.Sub(vert[1], vert[2])
			sp.WorldU = vec64.Sub(dp1.Scale(dv2*invdet), dp2.Scale(dv1*invdet))
			sp.WorldV = vec64.Sub(dp2.Scale(du1*invdet), dp1.Scale(du2*invdet))
		} else {
			sp.WorldU, sp.WorldV = vec64.Vector{}, vec64.Vector{}
		}
	} else {
		sp.U, sp.V = u, v
		sp.WorldU, sp.WorldV = vec64.Sub(vert[1], vert[0]), vec64.Sub(vert[2], vert[0])
	}

	sp.Object = tri.mesh
	sp.Primitive = tri
	sp.Light = tri.mesh.light
	sp.Material = tri.material

	sp.SurfaceU, sp.SurfaceV = u, v
	sp.PrimitiveNumber = tri.index
	sp.Position = coll.Point()

	sp.NormalU, sp.NormalV = vecutil.CreateCS(sp.Normal)
	sp.ShadingU[0] = vec64.Dot(sp.NormalU, sp.WorldU)
	sp.ShadingU[1] = vec64.Dot(sp.NormalV, sp.WorldU)
	sp.ShadingU[2] = vec64.Dot(sp.Normal, sp.WorldU)
	sp.ShadingV[0] = vec64.Dot(sp.NormalU, sp.WorldV)
	sp.ShadingV[1] = vec64.Dot(sp.NormalV, sp.WorldV)
	sp.ShadingV[2] = vec64.Dot(sp.Normal, sp.WorldV)

	return
}

func (tri *Triangle) Bound() (bd bound.Bound) {
	v := tri.getVertices()
	for axis := range bd.Min {
		bd.Min[axis] = math.Min(math.Min(v[0][axis], v[1][axis]), v[2][axis])
		bd.Max[axis] = math.Max(math.Max(v[0][axis], v[1][axis]), v[2][axis])
	}
	return
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

func (tri *Triangle) Material() Material { return tri.material }

func (tri *Triangle) Clip(bound bound.Bound, axis vecutil.Axis, lower bool, oldData interface{}) (clipped bound.Bound, newData interface{}) {
	if axis >= 0 {
		return tri.clipPlane(bound, axis, lower, oldData)
	}
	return tri.clipBox(bound)
}

func (tri *Triangle) clipPlane(bound bound.Bound, axis vecutil.Axis, lower bool, oldData interface{}) (clipped bound.Bound, newData interface{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Debugf("Clip plane fault: %v", err)
			clipped, newData = tri.clipBox(bound)
		}
	}()

	var poly []vec64.Vector
	if oldData == nil {
		v := tri.getVertices()
		poly = []vec64.Vector{v[0], v[1], v[2], v[0]}
	} else {
		poly = oldData.([]vec64.Vector)
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
			log.Warningf("Clip panic: %v", err)
		}
	}()

	v := tri.getVertices()
	poly := []vec64.Vector{v[0], v[1], v[2], v[0]}
	newData, clipped = triBoxClip([3]float64(bd.Min), [3]float64(bd.Max), poly)
	return
}

// The rest of these are non-interface triangle-specific methods.

func (tri *Triangle) SetMaterial(mat Material) { tri.material = mat }
func (tri *Triangle) SetNormals(a, b, c int)   { tri.n = [3]int{a, b, c} }
func (tri *Triangle) ClearNormals()            { tri.n = [3]int{-1, -1, -1} }
func (tri *Triangle) SetUVs(a, b, c int)       { tri.uv = [3]int{a, b, c} }
func (tri *Triangle) ClearUVs()                { tri.uv = [3]int{-1, -1, -1} }
func (tri *Triangle) GetNormal() vec64.Vector  { return tri.normal }

func (tri *Triangle) CalculateNormal() {
	v := tri.getVertices()
	tri.normal = vec64.Cross(vec64.Sub(v[1], v[0]), vec64.Sub(v[2], v[0])).Normalize()
}

func (tri *Triangle) SurfaceArea() float64 {
	v := tri.getVertices()
	edge1, edge2 := vec64.Sub(v[1], v[0]), vec64.Sub(v[2], v[0])
	return vec64.Cross(edge1, edge2).Length() * 0.5
}
