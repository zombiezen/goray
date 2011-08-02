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

import (
	"image"
	"math"
	"testing"
)

type ColorTest struct {
	Color          Color
	R, G, B, A     float64
	QR, QG, QB, QA uint32
}

func (ct ColorTest) Do(t *testing.T) {
	if ct.Color.Red() != ct.R {
		t.Errorf("%v.Red() got %#v (wanted %#v)", ct.Color, ct.Color.Red(), ct.R)
	}
	if ct.Color.Green() != ct.G {
		t.Errorf("%v.Green() got %#v (wanted %#v)", ct.Color, ct.Color.Green(), ct.G)
	}
	if ct.Color.Blue() != ct.B {
		t.Errorf("%v.Blue() got %#v (wanted %#v)", ct.Color, ct.Color.Blue(), ct.B)
	}
	if a, ok := ct.Color.(Alpha); ok && a.Alpha() != ct.A {
		t.Errorf("%v.Alpha() got %#v (wanted %#v)", ct.Color, a.Alpha(), ct.A)
	}

	if icol, ok := ct.Color.(image.Color); ok {
		ir, ig, ib, ia := icol.RGBA()
		if ir != ct.QR {
			t.Errorf("%v.RGBA()[0] got %d (wanted %d)", ct.Color, ir, ct.QR)
		}
		if ig != ct.QG {
			t.Errorf("%v.RGBA()[1] got %d (wanted %d)", ct.Color, ig, ct.QG)
		}
		if ib != ct.QB {
			t.Errorf("%v.RGBA()[2] got %d (wanted %d)", ct.Color, ib, ct.QB)
		}
		if ia != ct.QA {
			t.Errorf("%v.RGBA()[3] got %d (wanted %d)", ct.Color, ia, ct.QA)
		}
	}
}

func TestGray(t *testing.T) {
	cases := []ColorTest{
		// Col       R     G     B     A     QR     QG     QB     QA
		{Gray(0.00), 0.00, 0.00, 0.00, 0.00, 0, 0, 0, math.MaxUint32},
		{Gray(0.25), 0.25, 0.25, 0.25, 0.00, 16383, 16383, 16383, math.MaxUint32},
		{Gray(0.33), 0.33, 0.33, 0.33, 0.00, 21626, 21626, 21626, math.MaxUint32},
		{Gray(0.50), 0.50, 0.50, 0.50, 0.00, 32767, 32767, 32767, math.MaxUint32},
		{Gray(0.75), 0.75, 0.75, 0.75, 0.00, 49151, 49151, 49151, math.MaxUint32},
		{Gray(1.00), 1.00, 1.00, 1.00, 0.00, 65535, 65535, 65535, math.MaxUint32},
	}
	for _, c := range cases {
		c.Do(t)
	}
}

func TestRGB(t *testing.T) {
	cases := []ColorTest{
		// Col                 R     G     B     A     QR     QG     QB     QA
		{RGB{0.25, 0.33, 0.5}, 0.25, 0.33, 0.50, 0.00, 16383, 21626, 32767, math.MaxUint32},
	}
	for _, c := range cases {
		c.Do(t)
	}
}

func TestRGBA(t *testing.T) {
	cases := []ColorTest{
		// Col                       R     G     B     A     QR     QG     QB     QA
		{RGBA{0.25, 0.33, 0.5, 1.0}, 0.25, 0.33, 0.50, 1.00, 16383, 21626, 32767, 65535},
		{RGBA{0.25, 0.33, 0.5, 0.5}, 0.25, 0.33, 0.50, 0.50, 8191, 10813, 16383, 32767},
		{RGBA{0.25, 0.33, 0.5, 0.0}, 0.25, 0.33, 0.50, 0.00, 0, 0, 0, 0},
	}
	for _, c := range cases {
		c.Do(t)
	}
}

func TestModel(t *testing.T) {
	c1 := image.NRGBA64Color{0x1234, 0x4321, 0x1111, 0x7fff}
	c2 := Model.Convert(c1)
	if c3, ok := c2.(AlphaColor); ok {
		modelColorEq(t, "Red", c3.Red(), 0x1234)
		modelColorEq(t, "Green", c3.Green(), 0x4321)
		modelColorEq(t, "Blue", c3.Blue(), 0x1111)
		modelColorEq(t, "Alpha", c3.Alpha(), 0x7fff)
	} else {
		t.Error("Model doesn't give back AlphaColor")
	}
}

func modelColorEq(t *testing.T, name string, channel float64, val uint16) {
	const threshold = 1e-4
	expected := float64(val) / math.MaxUint16
	if channel > expected+threshold || channel < expected-threshold {
		t.Errorf("%s() -> %#v (expected %#v)", name, channel, expected)
	}
}
