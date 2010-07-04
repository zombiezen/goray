//
//	yaml/parser/schema.go
//	goray
//
//	Created by Ross Light on 2010-07-03.
//

package parser

import (
	"os"
)

const (
	YAMLDefaultPrefix = "tag:yaml.org,2002:"

	StringTag   = YAMLDefaultPrefix + "str"
	SequenceTag = YAMLDefaultPrefix + "seq"
	MappingTag  = YAMLDefaultPrefix + "map"

	NullTag  = YAMLDefaultPrefix + "null"
	BoolTag  = YAMLDefaultPrefix + "bool"
	IntTag   = YAMLDefaultPrefix + "int"
	FloatTag = YAMLDefaultPrefix + "float"
)

func init() {
	m := make(ConstructorMap)
	m[StringTag] = ConstructorFunc(constructStr)
	m[SequenceTag] = ConstructorFunc(constructSeq)
	m[MappingTag] = ConstructorFunc(constructMap)
	DefaultConstructor = m
}

// SCHEMAS

type Schema interface {
	Resolve(Node) (tag string, err os.Error)
}

type SchemaFunc func(Node) (string, os.Error)

func (f SchemaFunc) Resolve(n Node) (string, os.Error) { return f(n) }

var (
	FailsafeSchema Schema = SchemaFunc(failsafeSchema)
)

func failsafeSchema(node Node) (tag string, err os.Error) {
	switch node.(type) {
	case *Scalar:
		tag = StringTag
	case *Sequence:
		tag = SequenceTag
	case *Mapping:
		tag = MappingTag
	default:
		err = os.NewError("Unrecognized node given to failsafe schema")
	}
	return
}

// CONSTRUCTORS

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
