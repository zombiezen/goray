package color

import "testing"

import (
	"image"
	"math"
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
