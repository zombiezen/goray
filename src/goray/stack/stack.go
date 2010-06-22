//
//  goray/stack/stack.go
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
func (s *Stack) Copy() *Stack    { return &Stack{s.top} }
func (s *Stack) Front() *Element { return s.top }
func (s *Stack) Empty() bool     { return s.top == nil }
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

func (s *Stack) Top() interface{} {
	if s.top == nil {
		return nil
	}
	return s.top.Value
}

func (s *Stack) Push(v interface{}) {
	s.top = &Element{v, s.top}
}

func (s *Stack) Pop() (value interface{}) {
	if s.top == nil {
		return
	}
	value = s.top.Value
	s.top = s.top.Next()
	return
}
