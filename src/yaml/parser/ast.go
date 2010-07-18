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

// Document stores data related to a single document in a stream.
type Document struct {
	MajorVersion int
	MinorVersion int
	Content      Node
}

// Node defines a node of the representation graph.
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

// An Empty indicates a lack of a value.
type Empty struct {
	*basicNode
}

// A KeyValuePair stores a single pair in a mapping.
type KeyValuePair struct {
	Key, Value Node
}

// Mapping stores a key-value mapping of nodes.
type Mapping struct {
	*basicNode
	Pairs []KeyValuePair
}

// Len returns the number of pairs in the mapping.
func (m *Mapping) Len() int { return len(m.Pairs) }

// Get returns the value for a scalar that has the same content as key.
func (m *Mapping) Get(key string) (value Node, ok bool) {
	for _, pair := range m.Pairs {
		if pairKey, convOk := pair.Key.Data().(string); convOk && pairKey == key {
			return pair.Value, true
		}
	}
	return nil, false
}

// Map returns the content of the mapping as a native Go map.
func (node *Mapping) Map() (m map[interface{}]interface{}) {
	m = make(map[interface{}]interface{}, len(node.Pairs))
	for _, pair := range node.Pairs {
		m[pair.Key.Data()] = pair.Value.Data()
	}
	return
}

// Sequence stores an ordered collection of nodes.
type Sequence struct {
	*basicNode
	Nodes []Node
}

// At returns the node at the given index.
func (seq *Sequence) At(i int) Node { return seq.Nodes[i] }

// Len returns the number of nodes in the sequence.
func (seq *Sequence) Len() int      { return len(seq.Nodes) }

// Slice returns the content of the sequence as a native Go slice.
func (seq *Sequence) Slice() (s []interface{}) {
	s = make([]interface{}, len(seq.Nodes))
	for i, n := range seq.Nodes {
		s[i] = n.Data()
	}
	return
}

// Scalar stores a text value.
type Scalar struct {
	*basicNode
	Value string
}

// String returns the original string that was used to construct the value.
func (s *Scalar) String() string { return s.Value }
