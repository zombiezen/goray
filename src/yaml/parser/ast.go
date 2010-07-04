//
//	yaml/parser/ast.go
//	goray
//
//	Created by Ross Light on 2010-07-02.
//

package parser

import (
	"yaml/token"
)

type Document struct {
	MajorVersion int
	MinorVersion int
	Content      Node
}

type Node interface {
	Start() token.Position
	Tag() string
	Data() interface{}

	setTag(string)
	setData(interface{})
}

type basicNode struct {
	pos     token.Position
	tag     string
	hasData bool
	data    interface{}
}

func (n basicNode) Start() token.Position  { return n.pos }
func (n basicNode) Tag() string            { return n.tag }
func (n basicNode) Data() interface{}      { return n.data }
func (n *basicNode) setTag(t string)       { n.tag = t }
func (n *basicNode) setData(d interface{}) { n.hasData = true; n.data = d }

type Empty struct {
	*basicNode
}

type KeyValuePair struct {
	Key, Value Node
}

type Mapping struct {
	*basicNode
	Pairs []KeyValuePair
}

func (m *Mapping) Len() int { return len(m.Pairs) }

func (m *Mapping) Get(key string) (value Node, ok bool) {
	for _, pair := range m.Pairs {
		if pairKey, convOk := pair.Key.Data().(string); convOk && pairKey == key {
			return pair.Value, true
		}
	}
	return nil, false
}

type Sequence struct {
	*basicNode
	Nodes []Node
}

func (seq *Sequence) At(i int) Node { return seq.Nodes[i] }
func (seq *Sequence) Len() int      { return len(seq.Nodes) }

type Scalar struct {
	*basicNode
	Value string
}

func (s *Scalar) String() string { return s.Value }
