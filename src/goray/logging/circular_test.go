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
