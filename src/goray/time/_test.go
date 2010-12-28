package time

import "goray/fmath"
import "testing"

func TestSplit(t *testing.T) {
	if h, m, s := Time(20.4 * Second).Split(); !fmath.Eq(float(s), 20.4) || m != 0 || h != 0 {
		t.Error("Standalone seconds does not work")
	}
	if h, m, s := Time(42 * Minute + 3.14 * Second).Split(); !fmath.Eq(float(s), 3.14) || m != 42 || h != 0 {
		t.Error("Minutes and seconds does not work")
	}
	if h, m, s := Time(12 * Hour + 42 * Minute + 3.14 * Second).Split(); !fmath.Eq(float(s), 3.14) || m != 42 || h != 12 {
		t.Error("Hours, minutes, and seconds does not work")
	}
}

func TestString(t *testing.T) {
	if s := Time(20.5 * Second).String(); s != "20.500s" {
		t.Error("Standalone seconds does not work")
	}
	if s := Time(3 * Minute + 2.01 * Second).String(); s != "03:02.01" {
		t.Error("Minutes and seconds does not work")
	}
	if s := Time(4 * Hour + 3 * Minute + 2.01 * Second).String(); s != "4:03:02.01" {
		t.Error("Hours, minutes, and seconds does not work")
	}
}