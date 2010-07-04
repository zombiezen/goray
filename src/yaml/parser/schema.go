//
//	yaml/parser/schema.go
//	goray
//
//	Created by Ross Light on 2010-07-03.
//

package parser

import (
	"os"
	"regexp"
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
	CoreSchema     = SchemaFunc(coreSchema)
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

var (
	csNullPat  = regexp.MustCompile(`null|Null|NULL|~`)
	csBoolPat  = regexp.MustCompile(`(true|True|TRUE)|(false|False|FALSE)`)
	csDecPat   = regexp.MustCompile(`([-+]?)([0-9]+)`)
	csOctPat   = regexp.MustCompile(`0o([0-7]+)`)
	csHexPat   = regexp.MustCompile(`0x([0-9a-fA-F]+)`)
	csFloatPat = regexp.MustCompile(`([-+]?)(\.[0-9]+|[0-9]+(\.[0-9]*)?)([eE][-+]?[0-9]+)?`)
	csInfPat   = regexp.MustCompile(`([-+]?)\.(inf|Inf|INF)`)
	csNanPat   = regexp.MustCompile(`\.(nan|NaN|NAN)`)
)

func coreSchema(node Node) (tag string, err os.Error) {
	if scalar, ok := node.(*Scalar); ok {
		s := scalar.String()
		switch {
		case csNullPat.MatchString(s):
			return NullTag, nil
		case csBoolPat.MatchString(s):
			return BoolTag, nil
		case csDecPat.MatchString(s) || csOctPat.MatchString(s) || csHexPat.MatchString(s):
			return IntTag, nil
		case csFloatPat.MatchString(s) || csInfPat.MatchString(s) || csNanPat.MatchString(s):
			return FloatTag, nil
		default:
			return StringTag, nil
		}
	}

	return failsafeSchema(node)
}
