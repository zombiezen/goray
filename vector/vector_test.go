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

package vector

import "testing"

func TestAdd(t *testing.T) {
	tests := []struct {
		A, B, Result Vector3D
	}{
		{Vector3D{0, 0, 0}, Vector3D{0, 0, 0}, Vector3D{0, 0, 0}},
		{Vector3D{1, 2, 3}, Vector3D{4, 5, 6}, Vector3D{5, 7, 9}},
	}
	for _, tt := range tests {
		if r := Add(tt.A, tt.B); r != tt.Result {
			t.Errorf("Add(%v, %v) != %v (got %v)", tt.A, tt.B, tt.Result, r)
		}
	}
}

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Add(Vector3D{1, 2, 3}, Vector3D{4, 5, 6})
	}
}

func TestSub(t *testing.T) {
	tests := []struct {
		A, B, Result Vector3D
	}{
		{Vector3D{0, 0, 0}, Vector3D{0, 0, 0}, Vector3D{0, 0, 0}},
		{Vector3D{1, 2, 3}, Vector3D{6, 5, 4}, Vector3D{-5, -3, -1}},
	}
	for _, tt := range tests {
		if r := Sub(tt.A, tt.B); r != tt.Result {
			t.Errorf("Sub(%v, %v) != %v (got %v)", tt.A, tt.B, tt.Result, r)
		}
	}
}

func BenchmarkSub(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Sub(Vector3D{1, 2, 3}, Vector3D{4, 5, 6})
	}
}

func TestDot(t *testing.T) {
	tests := []struct {
		A, B   Vector3D
		Result float64
	}{
		{Vector3D{0, 0, 0}, Vector3D{0, 0, 0}, 0},
		{Vector3D{1, 2, 3}, Vector3D{4, 5, 6}, 32},
	}
	for _, tt := range tests {
		if r := Dot(tt.A, tt.B); r != tt.Result {
			t.Errorf("Dot(%v, %v) != %v (got %v)", tt.A, tt.B, tt.Result, r)
		}
	}
}

func BenchmarkDot(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Dot(Vector3D{1, 2, 3}, Vector3D{4, 5, 6})
	}
}

func TestNormalize(t *testing.T) {
	comps := []float64{1.0, 2.0, -2.0}
	length := 3.0
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
		Vec       Vector3D
		Length    float64
		LengthSqr float64
	}

	tests := []lengthTest{
		lengthTest{Vector3D{0.0, 0.0, 0.0}, 0.0, 0.0},

		lengthTest{Vector3D{1.0, 0.0, 0.0}, 1.0, 1.0},
		lengthTest{Vector3D{0.0, 1.0, 0.0}, 1.0, 1.0},
		lengthTest{Vector3D{0.0, 0.0, 1.0}, 1.0, 1.0},

		lengthTest{Vector3D{-1.0, 0.0, 0.0}, 1.0, 1.0},
		lengthTest{Vector3D{0.0, -1.0, 0.0}, 1.0, 1.0},
		lengthTest{Vector3D{0.0, 0.0, -1.0}, 1.0, 1.0},

		lengthTest{Vector3D{3.0, -4.0, 0.0}, 5.0, 25.0},
		lengthTest{Vector3D{1.0, 2.0, -2.0}, 3.0, 9.0},
		lengthTest{Vector3D{3.14, 20.7, 0.5}, 20.942769635365803, 438.59959999999995},
	}

	for _, ltest := range tests {
		if ltest.Vec.LengthSqr() != ltest.LengthSqr {
			t.Errorf("LengthSqr failed for %v (wanted %.2f, got %.2f)", ltest.Vec, ltest.LengthSqr, ltest.Vec.LengthSqr())
		}
		if ltest.Vec.Length() != ltest.Length {
			t.Errorf("Length failed for %v (wanted %.2f, got %.2f)", ltest.Vec, ltest.Length, ltest.Vec.Length())
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

func TestNegate(t *testing.T) {
	var v Vector3D

	v = Vector3D{0, 0, 0}.Negate()
	if v[X] != 0 || v[Y] != 0 || v[Z] != 0 {
		t.Error("Zero vector incorrect")
	}

	v = Vector3D{1, 2, 3}.Negate()
	if v[X] != -1 || v[Y] != -2 || v[Z] != -3 {
		t.Error("All positive vector incorrect")
	}

	v = Vector3D{-1, -2, -3}.Negate()
	if v[X] != 1 || v[Y] != 2 || v[Z] != 3 {
		t.Error("All negative vector incorrect")
	}

	v = Vector3D{-1, 2, -3}.Negate()
	if v[X] != 1 || v[Y] != -2 || v[Z] != 3 {
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
