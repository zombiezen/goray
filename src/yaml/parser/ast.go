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
}

type basicNode struct {
	pos token.Position
	tag string
}

func (n basicNode) Start() token.Position { return n.pos }
func (n basicNode) Tag() string           { return n.tag }

type KeyValuePair struct {
	Key, Value Node
}

type Mapping struct {
	basicNode
	Pairs []KeyValuePair
}

type Sequence struct {
	basicNode
	Nodes []Node
}

func (seq *Sequence) At(i int) Node {
	return seq.Nodes[i]
}

type Scalar struct {
	basicNode
	Value string
}

func (s *Scalar) String() string { return s.Value }
