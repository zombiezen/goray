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

package color

import "testing"

func TestBlack(t *testing.T) {
	if !IsBlack(Black) {
		t.Error("Black constant is not black")
	}
	if IsBlack(White) {
		t.Error("White constant is black")
	}
	if !IsBlack(RGB{0, 0, 0}) {
		t.Error("RGB{0, 0, 0} is not black")
	}
	if !IsBlack(RGBA{0, 0, 0, 0}) {
		t.Error("RGBA{0, 0, 0, 0} is not black")
	}
	if IsBlack(RGB{1, 0, 0}) {
		t.Error("RGB{1, 0, 0} is black")
	}
}

type energyTest struct {
	Color    Color
	Expected float64
}

func TestEnergy(t *testing.T) {
	cases := []energyTest{
		{Black, 0},
		{White, 1},
		{Red, 1 / 3.0},
		{Green, 1 / 3.0},
		{Blue, 1 / 3.0},
		{RGB{0.25, 0.70, 0.55}, 0.5},
	}
	for _, c := range cases {
		result := Energy(c.Color)
		if result != c.Expected {
			t.Errorf("%v energy is %#v (expected %#v)", c.Color, result, c.Expected)
		}
	}
}

type invertTest struct {
	Input, Output Color
}

func colorEq(c1, c2 Color) bool {
	return c1.Red() == c2.Red() && c1.Green() == c2.Green() && c1.Blue() == c2.Blue()
}

func TestInvert(t *testing.T) {
	cases := []invertTest{
		{Black, Black},
		{White, White},
		{Red, Red},
		{Green, Green},
		{Blue, Blue},
		{RGB{2.0, 2.0, 2.0}, RGB{0.5, 0.5, 0.5}},
		{RGB{1.0 / 3.0, 0.5, 10.0}, RGB{3.0, 2.0, 0.1}},
	}
	for _, c := range cases {
		result := Invert(c.Input)
		if !colorEq(result, c.Output) {
			t.Errorf("%#v inverted is %#v (expected %#v)", c.Input, result, c.Output)
		}
	}
}

type binaryTest struct {
	C1, C2   Color
	Expected Color
}

func (bt binaryTest) Do(t *testing.T, name string, f func(c1, c2 Color) Color) {
	result := f(bt.C1, bt.C2)
	if !colorEq(result, bt.Expected) {
		t.Errorf("%s(%#v, %#v) -> %#v (expected %#v)", name, bt.C1, bt.C2, result, bt.Expected)
	}
}

func TestAdd(t *testing.T) {
	cases := []binaryTest{
		{Black, Black, Black},
		{Black, White, White},
		{White, Black, White},
		{Gray(2.0), Gray(2.0), Gray(4.0)},
		{Red, Green, Yellow},
		{Red, Blue, Magenta},
		{Green, Blue, Cyan},
		{Red, Cyan, White},
		{Yellow, Blue, White},
		{Green, Magenta, White},
	}
	for _, c := range cases {
		c.Do(t, "Add", Add)
	}
}

func TestSub(t *testing.T) {
	cases := []binaryTest{
		{Black, Black, Black},
		{Black, White, Gray(-1.0)},
		{White, Black, White},
		{Gray(5.0), Gray(2.0), Gray(3.0)},
		{Yellow, Green, Red},
		{Yellow, Red, Green},
		{Magenta, Blue, Red},
		{Magenta, Red, Blue},
		{Cyan, Green, Blue},
		{Cyan, Blue, Green},
		{White, Red, Cyan},
		{White, Yellow, Blue},
		{White, Green, Magenta},
	}
	for _, c := range cases {
		c.Do(t, "Sub", Sub)
	}
}

func TestMul(t *testing.T) {
	cases := []binaryTest{
		{Black, Black, Black},
		{White, Black, Black},
		{Black, White, Black},
		{White, White, White},
		{Gray(5.0), Gray(2.0), Gray(10.0)},
		{Red, Green, Black},
		{Yellow, Red, Red},
	}
	for _, c := range cases {
		c.Do(t, "Mul", Mul)
	}
}

type binaryAlphaTest struct {
	C1, C2   AlphaColor
	Expected AlphaColor
}

func alphaColorEq(c1, c2 AlphaColor) bool {
	return c1.Alpha() == c2.Alpha() && colorEq(c1, c2)
}

func (bt binaryAlphaTest) Do(t *testing.T, name string, f func(c1, c2 AlphaColor) AlphaColor) {
	result := f(bt.C1, bt.C2)
	if !alphaColorEq(result, bt.Expected) {
		t.Errorf("%s(%#v, %#v) -> %#v (expected %#v)", name, bt.C1, bt.C2, result, bt.Expected)
	}
}

func TestAddAlpha(t *testing.T) {
	cases := []binaryAlphaTest{
		{RGBA{0, 0, 0, 0}, RGBA{0, 0, 0, 0}, RGBA{0, 0, 0, 0}},
		{RGBA{0, 0, 0, 0}, RGBA{1, 1, 1, 1}, RGBA{1, 1, 1, 1}},
		{RGBA{1, 1, 1, 1}, RGBA{0, 0, 0, 0}, RGBA{1, 1, 1, 1}},
		{RGBA{2, 2, 2, 2}, RGBA{2, 2, 2, 2}, RGBA{4, 4, 4, 4}},
		{RGBA{1, 0, 0, 1}, RGBA{0, 1, 0, 1}, RGBA{1, 1, 0, 2}},
		{RGBA{1, 0, 0, 1}, RGBA{0, 0, 1, 1}, RGBA{1, 0, 1, 2}},
		{RGBA{0, 1, 0, 1}, RGBA{0, 0, 1, 1}, RGBA{0, 1, 1, 2}},
		{RGBA{1, 0, 0, 1}, RGBA{0, 1, 1, 1}, RGBA{1, 1, 1, 2}},
		{RGBA{1, 1, 0, 1}, RGBA{0, 0, 1, 1}, RGBA{1, 1, 1, 2}},
		{RGBA{0, 1, 0, 1}, RGBA{1, 0, 1, 1}, RGBA{1, 1, 1, 2}},
	}
	for _, c := range cases {
		c.Do(t, "AddAlpha", AddAlpha)
	}
}
