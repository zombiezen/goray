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

package time

import "testing"

func TestSplit(t *testing.T) {
	if h, m, s := Time(20.4 * Second).Split(); s != 20.4 || m != 0 || h != 0 {
		t.Error("Standalone seconds does not work")
	}
	if h, m, s := Time(42*Minute + 3.14*Second).Split(); s > 3.145 || s < 3.135 || m != 42 || h != 0 {
		t.Errorf("Minutes and seconds does not work (got %d, %d, %.3f)", h, m, s)
	}
	if h, m, s := Time(12*Hour + 42*Minute + 3.14*Second).Split(); s > 3.145 || s < 3.135 || m != 42 || h != 12 {
		t.Errorf("Hours, minutes, and seconds does not work (got %d, %d, %.3f)", h, m, s)
	}
}

func TestString(t *testing.T) {
	if s := Time(20.5 * Second).String(); s != "20.500s" {
		t.Error("Standalone seconds does not work")
	}
	if s := Time(3*Minute + 2.01*Second).String(); s != "03:02.01" {
		t.Error("Minutes and seconds does not work")
	}
	if s := Time(4*Hour + 3*Minute + 2.01*Second).String(); s != "4:03:02.01" {
		t.Error("Hours, minutes, and seconds does not work")
	}
}
