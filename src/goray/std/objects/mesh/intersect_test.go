// intersect_test.go

package mesh

import "testing"
import (
	"goray/core/ray"
	"goray/core/vector"
)

type IntersectTest struct {
	Vertices [3][3]float64
	RayDir   [3]float64
	RayFrom  [3]float64

	RayDepth float64
	U, V     float64
}

func TestIntersect(t *testing.T) {
	testList := []IntersectTest{
		{
			Vertices: [3][3]float64{
				{1.0, 0.0, 0.0},
				{1.0, 1.0, 0.0},
				{1.0, 0.0, 1.0},
			},
			RayDir:   [3]float64{1.0, 0.0, 0.0},
			RayFrom:  [3]float64{0.0, 0.0, 0.0},
			RayDepth: 1.0,
			U:        0.0,
			V:        0.0,
		},
		{
			Vertices: [3][3]float64{
				{0.0, 1.0, 0.0},
				{1.0, 1.0, 0.0},
				{0.0, 1.0, 1.0},
			},
			RayDir:   [3]float64{0.0, 1.0, 0.0},
			RayFrom:  [3]float64{0.0, 0.0, 0.0},
			RayDepth: 1.0,
			U:        0.0,
			V:        0.0,
		},
		{
			Vertices: [3][3]float64{
				{0.0, 0.0, 1.0},
				{0.0, 1.0, 1.0},
				{1.0, 0.0, 1.0},
			},
			RayDir:   [3]float64{0.0, 0.0, 1.0},
			RayFrom:  [3]float64{0.0, 0.0, 0.0},
			RayDepth: 1.0,
			U:        0.0,
			V:        0.0,
		},
		{
			Vertices: [3][3]float64{
				{1.565772, -0.227881, -0.856351},
				{0.480624, 1.452136, -0.856351},
				{2.433322, 0.332482, 0.856351},
			},
			RayDir:   [3]float64{0.211504, 0.558421, -0.802142},
			RayFrom:  [3]float64{1.339351, 0.225915, -0.059020},
			RayDepth: 0.44048257340316493,
			U:        0.3300573931174704,
			V:        0.2592403276257273,
		},
	}
	for i, itest := range testList {
		rayDepth, u, v := intersect(itest.Vertices[0], itest.Vertices[1], itest.Vertices[2], itest.RayDir, itest.RayFrom)
		if rayDepth != itest.RayDepth || u != itest.U || v != itest.V {
			t.Errorf("intersect() [%d] failed, wanted %f, %f, %f but got %#v, %#v, %#v", i, itest.RayDepth, itest.U, itest.V, rayDepth, u, v)
		}
	}
}

func BenchmarkIntersect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		intersect(
			[3]float64{1.565772, -0.227881, -0.856351}, // A
			[3]float64{0.480624, 1.452136, -0.856351},  // B
			[3]float64{2.433322, 0.332482, 0.856351},   // C
			[3]float64{0.211504, 0.558421, -0.802142},  // Dir
			[3]float64{1.339351, 0.225915, -0.059020},  // From
		)
	}
}

func BenchmarkIntersectMethod(b *testing.B) {
	mesh := New(1, false)
	mesh.SetData([]vector.Vector3D{
		{1.565772, -0.227881, -0.856351},
		{0.480624, 1.452136, -0.856351},
		{2.433322, 0.332482, 0.856351},
	}, nil, nil)
	tri := NewTriangle(0, 1, 2, mesh)
	mesh.AddTriangle(tri)

	for i := 0; i < b.N; i++ {
		tri.Intersect(ray.Ray{
			Dir: vector.Vector3D{0.211504, 0.558421, -0.802142},
			From: vector.Vector3D{1.339351, 0.225915, -0.059020},
		})
	}
}
