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
	"testing"

	"bitbucket.org/zombiezen/math3/vec64"
)

var meshIntersectTests = []struct {
	Vertices [3]vec64.Vector
	RayDir   vec64.Vector
	RayFrom  vec64.Vector

	RayDepth float64
	U, V     float64
}{
	{
		Vertices: [3]vec64.Vector{
			{1.0, 0.0, 0.0},
			{1.0, 1.0, 0.0},
			{1.0, 0.0, 1.0},
		},
		RayDir:   vec64.Vector{1.0, 0.0, 0.0},
		RayFrom:  vec64.Vector{0.0, 0.0, 0.0},
		RayDepth: 1.0,
		U:        0.0,
		V:        0.0,
	},
	{
		Vertices: [3]vec64.Vector{
			{0.0, 1.0, 0.0},
			{1.0, 1.0, 0.0},
			{0.0, 1.0, 1.0},
		},
		RayDir:   vec64.Vector{0.0, 1.0, 0.0},
		RayFrom:  vec64.Vector{0.0, 0.0, 0.0},
		RayDepth: 1.0,
		U:        0.0,
		V:        0.0,
	},
	{
		Vertices: [3]vec64.Vector{
			{0.0, 0.0, 1.0},
			{0.0, 1.0, 1.0},
			{1.0, 0.0, 1.0},
		},
		RayDir:   vec64.Vector{0.0, 0.0, 1.0},
		RayFrom:  vec64.Vector{0.0, 0.0, 0.0},
		RayDepth: 1.0,
		U:        0.0,
		V:        0.0,
	},
	{
		Vertices: [3]vec64.Vector{
			{1.565772, -0.227881, -0.856351},
			{0.480624, 1.452136, -0.856351},
			{2.433322, 0.332482, 0.856351},
		},
		RayDir:   vec64.Vector{0.211504, 0.558421, -0.802142},
		RayFrom:  vec64.Vector{1.339351, 0.225915, -0.059020},
		RayDepth: 0.44048257340316493,
		U:        0.3300573931174704,
		V:        0.2592403276257273,
	},
}

func TestMeshIntersect(t *testing.T) {
	for i, itest := range meshIntersectTests {
		rayDepth, u, v := intersect(itest.Vertices[0], itest.Vertices[1], itest.Vertices[2], itest.RayDir, itest.RayFrom)
		if rayDepth != itest.RayDepth || u != itest.U || v != itest.V {
			t.Errorf("intersect() [%d] failed, wanted %f, %f, %f but got %#v, %#v, %#v", i, itest.RayDepth, itest.U, itest.V, rayDepth, u, v)
		}
	}
}

func TestGoMeshIntersect(t *testing.T) {
	for i, itest := range meshIntersectTests {
		rayDepth, u, v := intersect_go(itest.Vertices[0], itest.Vertices[1], itest.Vertices[2], itest.RayDir, itest.RayFrom)
		if rayDepth != itest.RayDepth || u != itest.U || v != itest.V {
			t.Errorf("intersect() [%d] failed, wanted %f, %f, %f but got %#v, %#v, %#v", i, itest.RayDepth, itest.U, itest.V, rayDepth, u, v)
		}
	}
}

func BenchmarkMeshIntersect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		intersect(
			vec64.Vector{1.565772, -0.227881, -0.856351}, // A
			vec64.Vector{0.480624, 1.452136, -0.856351},  // B
			vec64.Vector{2.433322, 0.332482, 0.856351},   // C
			vec64.Vector{0.211504, 0.558421, -0.802142},  // Dir
			vec64.Vector{1.339351, 0.225915, -0.059020},  // From
		)
	}
}

func BenchmarkGoMeshIntersect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		intersect_go(
			vec64.Vector{1.565772, -0.227881, -0.856351}, // A
			vec64.Vector{0.480624, 1.452136, -0.856351},  // B
			vec64.Vector{2.433322, 0.332482, 0.856351},   // C
			vec64.Vector{0.211504, 0.558421, -0.802142},  // Dir
			vec64.Vector{1.339351, 0.225915, -0.059020},  // From
		)
	}
}

func BenchmarkMeshIntersectMethod(b *testing.B) {
	mesh := NewMesh(1, false)
	mesh.SetData([]vec64.Vector{
		{1.565772, -0.227881, -0.856351},
		{0.480624, 1.452136, -0.856351},
		{2.433322, 0.332482, 0.856351},
	}, nil, nil)
	tri := NewTriangle(0, 1, 2, mesh)
	mesh.AddTriangle(tri)

	for i := 0; i < b.N; i++ {
		tri.Intersect(Ray{
			Dir:  vec64.Vector{0.211504, 0.558421, -0.802142},
			From: vec64.Vector{1.339351, 0.225915, -0.059020},
		})
	}
}
