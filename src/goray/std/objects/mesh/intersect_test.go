// intersect_test.go

package mesh

import "testing"

type IntersectTest struct {
	Vertices [3][3]float64
	RayDir [3]float64
	RayFrom [3]float64

	RayDepth float64
	U, V float64
}

func TestIntersect(t *testing.T) {
	testList := []IntersectTest{
		{
			Vertices: [3][3]float64{
				{1.0, 0.0, 0.0},
				{1.0, 1.0, 0.0},
				{1.0, 0.0, 1.0},
			},
			RayDir: [3]float64{1.0, 0.0, 0.0},
			RayFrom: [3]float64{0.0, 0.0, 0.0},
			RayDepth: 1.0,
			U: 0.0,
			V: 0.0,
		},
		{
			Vertices: [3][3]float64{
				{0.0, 1.0, 0.0},
				{1.0, 1.0, 0.0},
				{0.0, 1.0, 1.0},
			},
			RayDir: [3]float64{0.0, 1.0, 0.0},
			RayFrom: [3]float64{0.0, 0.0, 0.0},
			RayDepth: 1.0,
			U: 0.0,
			V: 0.0,
		},
		{
			Vertices: [3][3]float64{
				{0.0, 0.0, 1.0},
				{0.0, 1.0, 1.0},
				{1.0, 0.0, 1.0},
			},
			RayDir: [3]float64{0.0, 0.0, 1.0},
			RayFrom: [3]float64{0.0, 0.0, 0.0},
			RayDepth: 1.0,
			U: 0.0,
			V: 0.0,
		},
	}
	for i, itest := range testList {
		rayDepth, u, v := intersect(itest.Vertices[0], itest.Vertices[1], itest.Vertices[2], itest.RayDir, itest.RayFrom)
		if rayDepth != itest.RayDepth || u != itest.U || v != itest.V {
			t.Errorf("intersect() [%d] failed, wanted %.3f, %.3f, %.3f but got %.3f, %.3f, %.3f", i, itest.RayDepth, itest.U, itest.V, rayDepth, u, v)
		}
	}
}
