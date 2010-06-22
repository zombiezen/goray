package partition

import "testing"

import (
	"goray/fmath"
	"goray/logging"
	"goray/core/primitive"
	"goray/core/ray"
	"goray/core/vector"
	"goray/std/primitives/sphere"
)

func TestDepth(t *testing.T) {
	type depthTestCase struct {
		Name string
		Partitioner Partitioner
	}
	
	r := ray.New()
	r.SetFrom(vector.New(2, 0, 0))
	r.SetDir(vector.New(-1, 0, 0))
	
	sphereA := sphere.New(vector.New(1, 0, 0), 0.25, nil)
	sphereB := sphere.New(vector.New(0, 0, 0), 0.25, nil)
	sphereC := sphere.New(vector.New(-1, 0, 0), 0.25, nil)
	
	log := logging.NewLogger()
	cases := []depthTestCase{
		depthTestCase{"Simple", NewSimple([]primitive.Primitive{
			sphereB, sphereC, sphereA,
		})},
		depthTestCase{"kd-tree", NewKD([]primitive.Primitive{
			sphereB, sphereC, sphereA,
		}, log)},
		depthTestCase{"kd-tree leaf", NewKD([]primitive.Primitive{
			sphereB, sphereA,
		}, log)},
	}
	
	for _, c := range cases {
		coll := c.Partitioner.Intersect(r, fmath.Inf)
		if coll.Hit() {
			if coll.Primitive != sphereA {
				t.Errorf("%s partitioner fails depth test", c.Name)
			}
		} else {
			t.Errorf("%s partitioner won't collide; depth test skipped.", c.Name)
		}
	}
}
