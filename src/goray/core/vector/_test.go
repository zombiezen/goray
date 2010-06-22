package vector

import "testing"
import "goray/fmath"

func TestNew(t *testing.T) {
	comps := []float{1.5, -5.7, 4.0}
	v := New(comps[0], comps[1], comps[2])

	if !fmath.Eq(v.X, comps[0]) {
		t.Error("X component does not match")
	}
	if !fmath.Eq(v.Y, comps[1]) {
		t.Error("Y component does not match")
	}
	if !fmath.Eq(v.Z, comps[2]) {
		t.Error("Z component does not match")
	}
}

func TestNormalize(t *testing.T) {
	comps := []float{1.0, 2.0, -2.0}
	length := 3.0
	v := New(comps[0], comps[1], comps[2])
	vn := v.Normalize()
	if vn.Length() != 1.0 {
		t.Error("Length is not 1")
	}
	if vn.X != comps[0]/length {
		t.Error("X component is incorrect")
	}
	if vn.Y != comps[1]/length {
		t.Error("Y component is incorrect")
	}
	if vn.Z != comps[2]/length {
		t.Error("Z component is incorrect")
	}
}

func TestLength(t *testing.T) {
	type lengthTest struct {
		Comps  [3]float
		Length float
	}

	tests := []lengthTest{
		lengthTest{[3]float{0.0, 0.0, 0.0}, 0.0},

		lengthTest{[3]float{1.0, 0.0, 0.0}, 1.0},
		lengthTest{[3]float{0.0, 1.0, 0.0}, 1.0},
		lengthTest{[3]float{0.0, 0.0, 1.0}, 1.0},

		lengthTest{[3]float{-1.0, 0.0, 0.0}, 1.0},
		lengthTest{[3]float{0.0, -1.0, 0.0}, 1.0},
		lengthTest{[3]float{0.0, 0.0, -1.0}, 1.0},

		lengthTest{[3]float{3.0, -4.0, 0.0}, 5.0},
		lengthTest{[3]float{1.0, 2.0, -2.0}, 3.0},
		lengthTest{[3]float{3.14, 20.7, 0.5}, 20.942769635365803},
	}

	for _, ltest := range tests {
		v := New(ltest.Comps[0], ltest.Comps[1], ltest.Comps[2])
		if lensqr := ltest.Length * ltest.Length; !fmath.Eq(v.LengthSqr(), lensqr) {
			t.Error("LengthSqr failed for %v (wanted %.2f, got %.2f)", v, lensqr, v.LengthSqr())
		}
		if !fmath.Eq(v.Length(), ltest.Length) {
			t.Error("Length failed for %v (wanted %.2f, got %.2f)", v, ltest.Length, v.Length())
		}
	}
}

func TestAbs(t *testing.T) {
	if v := New(0, 0, 0).Abs(); v.X != 0 || v.Y != 0 || v.Z != 0 {
		t.Error("Zero vector incorrect")
	}
	if v := New(1, 2, 3).Abs(); v.X != 1 || v.Y != 2 || v.Z != 3 {
		t.Error("All positive vector incorrect")
	}
	if v := New(-1, -2, -3).Abs(); v.X != 1 || v.Y != 2 || v.Z != 3 {
		t.Error("All negative vector incorrect")
	}
	if v := New(-1, 2, -3).Abs(); v.X != 1 || v.Y != 2 || v.Z != 3 {
		t.Error("Mixed vector incorrect")
	}
}

func TestIsZero(t *testing.T) {
	if v := New(0, 0, 0); !v.IsZero() {
		t.Error("Zero vector is not zero")
	}
	if v := New(1, 0, 0); v.IsZero() {
		t.Error("Positive X vector is zero")
	}
	if v := New(-1, 0, 0); v.IsZero() {
		t.Error("Negative X vector is zero")
	}
}
