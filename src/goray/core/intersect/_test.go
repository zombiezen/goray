package intersect

import "testing"

import (
	"math"
	"goray/core/primitive"
	"goray/core/ray"
	"goray/core/vector"
	"goray/std/primitives/sphere"
)

func TestDepth(t *testing.T) {
	type depthTestCase struct {
		Name      string
		Intersect Interface
	}

	r := ray.Ray{
		From: vector.Vector3D{2, 0, 0},
		Dir:  vector.Vector3D{-1, 0, 0},
		TMax: -1.0,
	}

	sphereA := sphere.New(vector.Vector3D{1, 0, 0}, 0.25, nil)
	sphereB := sphere.New(vector.Vector3D{0, 0, 0}, 0.25, nil)
	sphereC := sphere.New(vector.Vector3D{-1, 0, 0}, 0.25, nil)

	cases := []depthTestCase{
		{"Simple", NewSimple(
			[]primitive.Primitive{sphereB, sphereC, sphereA},
		)},
		{"kd-tree", NewKD(
			[]primitive.Primitive{sphereB, sphereC, sphereA},
			nil,
		)},
		{"kd-tree leaf", NewKD(
			[]primitive.Primitive{sphereB, sphereA},
			nil,
		)},
	}

	for _, c := range cases {
		coll := c.Intersect.Intersect(r, math.Inf(1))
		if coll.Hit() {
			if coll.Primitive != sphereA {
				t.Errorf("%s intersect fails depth test", c.Name)
			}
		} else {
			t.Errorf("%s intersect won't collide; depth test skipped.", c.Name)
		}
	}
}
