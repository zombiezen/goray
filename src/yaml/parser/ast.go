//
//	yaml/parser/ast.go
//	goray
//
//	Created by Ross Light on 2010-07-02.
//

package parser

import (
	"reflect"
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

// Data retrieval

func GetNodeBool(node Node) (b bool, ok bool) {
	b, ok = node.Data().(bool)
	return
}

func GetNodeFloat(node Node) (f float64, ok bool) {
	val := reflect.NewValue(node.Data())
	ok = true
	
	switch realVal := val.(type) {
	case *reflect.FloatValue:
		f = realVal.Get()
	case *reflect.IntValue:
		f = float64(realVal.Get())
	case *reflect.UintValue:
		f = float64(realVal.Get())
	default:
		ok = false
	}
	return
}

func GetNodeUint(node Node) (i uint64, ok bool) {
	val := reflect.NewValue(node.Data())
	ok = true
	
	switch realVal := val.(type) {
	case *reflect.UintValue:
		i = realVal.Get()
	case *reflect.IntValue:
		if realVal.Get() >= 0 {
			i = uint64(realVal.Get())
		} else {
			ok = false
		}
	default:
		ok = false
	}
	return
}

func GetNodeInt(node Node) (i int64, ok bool) {
	val := reflect.NewValue(node.Data())
	ok = true
	
	switch realVal := val.(type) {
	case *reflect.IntValue:
		i = realVal.Get()
	case *reflect.UintValue:
		i = int64(realVal.Get())
	default:
		ok = false
	}
	return
}

func GetNodeSequence(node Node) (seq []interface{}, ok bool) {
	seq, ok = node.Data().([]interface{})
	return
}

func GetNodeMap(node Node) (m map[interface{}]interface{}, ok bool) {
	m, ok = node.Data().(map[interface{}]interface{})
	return
}
