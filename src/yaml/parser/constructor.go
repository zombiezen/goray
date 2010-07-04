//
//	yaml/parser/constructor.go
//	goray
//
//	Created by Ross Light on 2010-07-04.
//

package parser

import (
	"os"
)

func init() {
	m := make(ConstructorMap)
	m[StringTag] = ConstructorFunc(constructStr)
	m[SequenceTag] = ConstructorFunc(constructSeq)
	m[MappingTag] = ConstructorFunc(constructMap)
	DefaultConstructor = m
}

type Constructor interface {
	Construct(Node) (data interface{}, err os.Error)
}

type ConstructorFunc func(Node) (interface{}, os.Error)

func (f ConstructorFunc) Construct(n Node) (interface{}, os.Error) { return f(n) }

type ConstructorMap map[string]Constructor

func (m ConstructorMap) Construct(n Node) (data interface{}, err os.Error) {
	if c, ok := m[n.Tag()]; ok {
		return c.Construct(n)
	}
	err = os.NewError("Constructor has no rule for " + n.Tag())
	return
}

var DefaultConstructor Constructor

func constructStr(n Node) (data interface{}, err os.Error) {
	node, ok := n.(*Scalar)
	if !ok {
		err = os.NewError("Non-scalar given to string")
		return
	}
	data = node.String()
	return
}

func constructSeq(n Node) (data interface{}, err os.Error) {
	node, ok := n.(*Sequence)
	if !ok {
		err = os.NewError("Non-sequence given to sequence")
		return
	}

	a := make([]interface{}, node.Len())
	for i := 0; i < node.Len(); i++ {
		a[i] = node.At(i).Data()
	}
	data = a
	return
}

func constructMap(n Node) (data interface{}, err os.Error) {
	node, ok := n.(*Mapping)
	if !ok {
		err = os.NewError("Non-mapping given to map")
		return
	}

	m := make(map[interface{}]interface{}, node.Len())
	for _, pair := range node.Pairs {
		k := pair.Key.Data()
		v := pair.Value.Data()
		m[k] = v
	}
	data = m
	return
}
