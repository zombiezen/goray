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
