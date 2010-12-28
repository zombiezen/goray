package intersect

import "testing"

import (
	"goray/fmath"
	"goray/core/primitive"
	"goray/core/ray"
	"goray/core/vector"
	"goray/std/primitives/sphere"
)

func TestDepth(t *testing.T) {
	type depthTestCase struct {
		Name        string
		Intersect Interface
	}

	r := ray.New()
	r.SetFrom(vector.New(2, 0, 0))
	r.SetDir(vector.New(-1, 0, 0))

	sphereA := sphere.New(vector.New(1, 0, 0), 0.25, nil)
	sphereB := sphere.New(vector.New(0, 0, 0), 0.25, nil)
	sphereC := sphere.New(vector.New(-1, 0, 0), 0.25, nil)

	cases := []depthTestCase{
		depthTestCase{"Simple", NewSimple(
			[]primitive.Primitive{sphereB, sphereC, sphereA},
		)},
		depthTestCase{"kd-tree", NewKD(
			[]primitive.Primitive{sphereB, sphereC, sphereA},
			nil,
		)},
		depthTestCase{"kd-tree leaf", NewKD(
			[]primitive.Primitive{sphereB, sphereA},
			nil,
		)},
	}

	for _, c := range cases {
		coll := c.Intersect.Intersect(r, fmath.Inf)
		if coll.Hit() {
			if coll.Primitive != sphereA {
				t.Errorf("%s intersect fails depth test", c.Name)
			}
		} else {
			t.Errorf("%s intersect won't collide; depth test skipped.", c.Name)
		}
	}
}