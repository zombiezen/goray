//
//	goray/stack/stack.go
//	goray
//
//	Created by Ross Light on 2010-05-23.
//

// The stack package provides a generic stack data type.
package stack

// An Element is a node in a stack's linked list.
type Element struct {
	Value interface{}

	next *Element
}

// Next returns the next element in the stack (in FILO order).
func (e *Element) Next() *Element { return e.next }

// Stack holds a linked list, FILO collection.  Stack's zero value is immediately usable.
type Stack struct {
	top *Element
}

// New creates a new stack.
func New() *Stack {
	return (&Stack{}).Init()
}

// Init clears out a stack.
func (s *Stack) Init() *Stack { s.top = nil; return s }

// Copy returns a shallow copy of the stack.
//
// The copy operation used here is very cheap because it simply points to the
// original font element.  This will not disrupt data unless you start
// modifying elements yourself.
func (s *Stack) Copy() *Stack { return &Stack{s.top} }

// Front returns the stack's most recently pushed element.
func (s *Stack) Front() *Element { return s.top }

// Empty returns whether the stack is empty.
func (s *Stack) Empty() bool { return s.top == nil }

// Iter returns a channel which can be used to iterate over the Stack in FILO order.
func (s *Stack) Iter() <-chan interface{} {
	c := make(chan interface{})
	go func() {
		defer close(c)
		for elem := s.top; elem != nil; elem = elem.Next() {
			c <- elem.Value
		}
	}()
	return c
}

// Top returns the most recently pushed value.
func (s *Stack) Top() interface{} {
	if s.top == nil {
		return nil
	}
	return s.top.Value
}

// Push adds a new Element to the stack.
func (s *Stack) Push(v interface{}) {
	s.top = &Element{v, s.top}
}

// Pop removes the most recently pushed value from the stack.
func (s *Stack) Pop() (value interface{}) {
	if s.top == nil {
		return
	}
	value = s.top.Value
	s.top = s.top.Next()
	return
}
