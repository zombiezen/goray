package vector

import "testing"

func TestNormalize(t *testing.T) {
	comps := []float64{1.0, 2.0, -2.0}
	length := float64(3.0)
	v := Vector3D{comps[0], comps[1], comps[2]}
	vn := v.Normalize()
	if vn.Length() != 1.0 {
		t.Error("Length is not 1")
	}
	for axis := X; axis <= Z; axis++ {
		if vn[axis] != comps[axis]/length {
			t.Error("X component is incorrect")
		}
	}
}

func TestLength(t *testing.T) {
	type lengthTest struct {
		Vec    Vector3D
		Length float64
	}

	tests := []lengthTest{
		lengthTest{[3]float64{0.0, 0.0, 0.0}, 0.0},

		lengthTest{[3]float64{1.0, 0.0, 0.0}, 1.0},
		lengthTest{[3]float64{0.0, 1.0, 0.0}, 1.0},
		lengthTest{[3]float64{0.0, 0.0, 1.0}, 1.0},

		lengthTest{[3]float64{-1.0, 0.0, 0.0}, 1.0},
		lengthTest{[3]float64{0.0, -1.0, 0.0}, 1.0},
		lengthTest{[3]float64{0.0, 0.0, -1.0}, 1.0},

		lengthTest{[3]float64{3.0, -4.0, 0.0}, 5.0},
		lengthTest{[3]float64{1.0, 2.0, -2.0}, 3.0},
		lengthTest{[3]float64{3.14, 20.7, 0.5}, 20.942769635365803},
	}

	for _, ltest := range tests {
		if lensqr := ltest.Length * ltest.Length; ltest.Vec.LengthSqr() != lensqr {
			t.Error("LengthSqr failed for %v (wanted %.2f, got %.2f)", ltest.Vec, lensqr, ltest.Vec.LengthSqr())
		}
		if ltest.Vec.Length() != ltest.Length {
			t.Error("Length failed for %v (wanted %.2f, got %.2f)", ltest.Vec, ltest.Length, ltest.Vec.Length())
		}
	}
}

func TestAbs(t *testing.T) {
	var v Vector3D

	v = Vector3D{0, 0, 0}.Abs()
	if v[X] != 0 || v[Y] != 0 || v[Z] != 0 {
		t.Error("Zero vector incorrect")
	}

	v = Vector3D{1, 2, 3}.Abs()
	if v[X] != 1 || v[Y] != 2 || v[Z] != 3 {
		t.Error("All positive vector incorrect")
	}

	v = Vector3D{-1, -2, -3}.Abs()
	if v[X] != 1 || v[Y] != 2 || v[Z] != 3 {
		t.Error("All negative vector incorrect")
	}

	v = Vector3D{-1, 2, -3}.Abs()
	if v[X] != 1 || v[Y] != 2 || v[Z] != 3 {
		t.Error("Mixed vector incorrect")
	}
}

func TestIsZero(t *testing.T) {
	var v Vector3D

	v = Vector3D{0, 0, 0}
	if !v.IsZero() {
		t.Error("Zero vector is not zero")
	}

	v = Vector3D{1, 0, 0}
	if v.IsZero() {
		t.Error("Positive X vector is zero")
	}

	v = Vector3D{-1, 0, 0}
	if v.IsZero() {
		t.Error("Negative X vector is zero")
	}
}
