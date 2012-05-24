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

package logging

import "testing"

func TestCircularHandler(t *testing.T) {
	data := []string{"Foo", "Bar", "Baz"}
	h := NewCircularHandler(len(data))
	if len(h.Records()) != 0 {
		t.Error("New handler is not empty")
	}
	if h.Full() {
		t.Error("New handler is full")
	}
	// Check non-filled
	for i, s := range data {
		h.Handle(StringRecord(s))
		if i < len(data)-1 && h.Full() {
			t.Errorf("Handler is full after %d records", i+1)
		}
		r := h.Records()
		if len(r) != i+1 {
			t.Errorf("Record length wrong after %d records (length = %d)", i+1, len(r))
		} else {
			for j := 0; j <= i; j++ {
				if r[j].String() != data[j] {
					t.Errorf("%d added: Records()[%d] != %#v (got %#v)", i+1, j, data[j], r[j].String())
				}
			}
		}
	}
	// Check full
	h.Handle(StringRecord("Bob"))
	if !h.Full() {
		t.Error("Full buffer does not report full")
	}
	r := h.Records()
	if len(r) != len(data) {
		t.Errorf("Records() on full buffer gives length %d", len(r))
		return
	}
	for i := 0; i < len(data); i++ {
		var s, expected string
		s = r[i].String()
		if i < len(data)-1 {
			expected = data[i+1]
		} else {
			expected = "Bob"
		}
		if s != expected {
			t.Errorf("full: Records[%d] != %#v (got %#v)", i, expected, s)
		}
	}
}
