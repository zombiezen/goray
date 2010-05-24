//
//  stack.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

package stack

type Element struct {
	Value interface{}

	next *Element
}

func (e *Element) Next() *Element { return e.next }

type Stack struct {
	top *Element
}

func New() *Stack {
	return (&Stack{}).Init()
}

func (s *Stack) Init() *Stack    { s.top = nil; return s }
func (s *Stack) Front() *Element { return s.top }
func (s *Stack) Iter() <-chan interface{} {
	c := make(chan interface{})
	go func() {
		for elem := s.top; elem != nil; elem = elem.Next() {
			c <- elem.Value
		}
		close(c)
	}()
	return c
}

func (s *Stack) Top() (value interface{}, ok bool) {
	if s.top == nil {
		return nil, false
	}
	return s.top.Value, true
}

func (s *Stack) Push(v interface{}) {
	s.top = &Element{v, s.top}
}

func (s *Stack) Pop() (value interface{}, ok bool) {
	if s.top == nil {
		return
	}
	value = s.top.Value
	s.top = s.top.Next()
	ok = true
	return
}
