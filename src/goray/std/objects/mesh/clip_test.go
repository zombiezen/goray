//
//	goray/std/objects/mesh/clip_test.go
//	goray
//
//	Created by Ross Light on 2010-12-27.
//

package mesh

import (
	"testing"
	"goray/logging"
	"goray/core/bound"
	"goray/core/vector"
)

type boxClipTest struct {
	Min, Max        [3]float64
	PolyIn, PolyOut []dVector
	Region          *bound.Bound
}

var boxTests = []boxClipTest{
	{
		Min:     [3]float64{-1, -1, -1},
		Max:     [3]float64{0, 0, 0},
		PolyIn:  []dVector{{0.1, 0.1, 0.1}, {0.9, 0.9, 0.9}, {0.9, 0.1, 0.9}, {0.1, 0.1, 0.1}},
		PolyOut: nil,
		Region:  nil,
	},
	{
		Min:     [3]float64{0, 0, 0},
		Max:     [3]float64{1, 1, 1},
		PolyIn:  []dVector{{0.1, 0.1, 0.1}, {0.9, 0.9, 0.9}, {0.9, 0.1, 0.9}, {0.1, 0.1, 0.1}},
		PolyOut: []dVector{{0.1, 0.1, 0.1}, {0.9, 0.9, 0.9}, {0.9, 0.1, 0.9}, {0.1, 0.1, 0.1}},
		Region:  bound.New(vector.New(0.1, 0.1, 0.1), vector.New(0.9, 0.9, 0.9)),
	},
}

func (test boxClipTest) Run(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("TEST PANIC: %v", err)
		}
	}()

	resultPoly, resultBound := triBoxClip(test.Min, test.Max, test.PolyIn)

	switch {
	case test.PolyOut == nil:
		if resultPoly != nil {
			t.Error("Resulting polygon not nil")
		}
	case resultPoly == nil:
		t.Error("Resulting polygon is nil")
	case len(resultPoly) != len(test.PolyOut):
		t.Errorf("Resulting polygon has %d polys, want %d.", len(resultPoly), len(test.PolyOut))
	default:
		for i, _ := range resultPoly {
			pa, pb := test.PolyOut[i], resultPoly[i]
			if pa[0] != pb[0] || pa[1] != pb[1] || pa[2] != pb[2] {
				t.Errorf("Vertex %d: (%.2f, %.2f, %.2f) != (%.2f, %.2f, %.2f) (Result)", i, pa[0], pa[1], pa[2], pb[0], pb[1], pb[2])
			}
		}
	}

	switch {
	case test.Region == nil:
		if resultBound != nil {
			t.Error("Resulting bound not nil")
		}
	case resultBound == nil:
		t.Error("Resulting bound is nil")
	default:
		amin, amax := test.Region.Get()
		bmin, bmax := resultBound.Get()
		if !(amin.X == bmin.X && amin.Y == bmin.Y && amin.Z == bmin.Z && amax.X == bmax.X && amax.Y == bmax.Y && amax.Z == bmax.Z) {
			t.Errorf("Bound: %v, wanted %v", resultBound, test.Region)
		}

	}
}

type testLogHandler struct {
	t *testing.T
}

func (self testLogHandler) Handle(rec logging.Record) {
	self.t.Logf("LOG: %s", rec)
}

func TestBoxClip(t *testing.T) {
	logging.MainLog.AddHandler(testLogHandler{t})
	defer func() {
		logging.MainLog = logging.NewLogger()
	}()

	for i, test := range boxTests {
		t.Logf("** Test [%d]", i)
		test.Run(t)
	}
}
