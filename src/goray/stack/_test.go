package stack

import "testing"

func TestPush(t *testing.T) {
	s := New()
	data := []string{"Foo", "Bar", "Baz"}
	for i, datum := range data {
		s.Push(datum)

		if pushed, ok := s.Top().(string); ok {
			if pushed != datum {
				t.Errorf("Mismatched Top data for datum %d: %s", i, pushed)
			}
		} else {
			t.Errorf("Wrong Top data for datum %d: %V", i, s.Top())
		}
		if s.Front() != nil {
			if pushed, ok := s.Front().Value.(string); ok {
				if pushed != datum {
					t.Errorf("Mismatched Top data for datum %d: %s", i, pushed)
				}
			} else {
				t.Errorf("Wrong Top data for datum %d: %V", i, s.Top())
			}
		} else {
			t.Errorf("Front returns nil for datum %d", i)
		}
	}
}

func TestEmpty(t *testing.T) {
	s := New()
	if !s.Empty() {
		t.Fatal("New stack is not empty")
	}
	if s.Push("Foo"); s.Empty() {
		t.Error("Stack is still empty after first push")
	}
	if s.Push("Foo"); s.Empty() {
		t.Error("Stack is empty after second push")
	}
	if s.Push("Foo"); s.Empty() {
		t.Error("Stack is empty after third push")
	}

	if s.Pop(); s.Empty() {
		t.Error("Stack is empty after first pop")
	}
	if s.Pop(); s.Empty() {
		t.Error("Stack is empty after second pop")
	}
	if s.Pop(); !s.Empty() {
		t.Error("Stack is not empty after last pop")
	}
}

func TestCrash(t *testing.T) {
	s := New()
	if val := s.Pop(); val != nil {
		t.Error("Stack does not return nil from empty pop")
	}
	if !s.Empty() {
		t.Error("Empty pop causes the stack to become un-empty")
	}
	s.Push("Foo")
	s.Pop()
	if val := s.Pop(); val != nil {
		t.Error("Used stack doesn't return nil from empty pop")
	}
}

func TestIter(t *testing.T) {
	data := []string{"Foo", "Bar", "Baz"}
	s := New()
	for _, datum := range data {
		s.Push(datum)
	}
	i := len(data) - 1
	for val := range s.Iter() {
		if i < 0 {
			t.Error("Iter exceeded length")
		}
		if str, ok := val.(string); ok {
			if str != data[i] {
				t.Errorf("Mismatched data: %s (expected %s)", str, data[i])
			}
		} else {
			t.Error("Iter gives the wrong type")
		}
		i--
	}

	if i != -1 {
		t.Errorf("Iter gave wrong length: %d (expected %d)", len(data)-(i+1), len(data))
	}
}
