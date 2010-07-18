//
//	yaml/scanner/result.go
//	goray
//
//	Created by Ross Light on 2010-06-25.
//

package scanner

import (
	"fmt"
	"yaml/token"
)

// Token holds data for a single lexical unit in a YAML document.
type Token interface {
	GetKind() token.Token
	GetStart() token.Position
	GetEnd() token.Position
	String() string
}

// BasicToken holds the essential information for a Token.
type BasicToken struct {
	Kind       token.Token
	Start, End token.Position
}

func (t BasicToken) GetKind() token.Token     { return t.Kind }
func (t BasicToken) GetStart() token.Position { return t.Start }
func (t BasicToken) GetEnd() token.Position   { return t.End }

func (t BasicToken) String() string {
	return fmt.Sprintf("%v %v", t.Start, t.Kind)
}

// ValueToken defines a token that holds a string value.  It is used for anchors, aliases, and scalars.
type ValueToken struct {
	BasicToken
	Value string
}

func (t ValueToken) String() string { return t.Value }

// TagToken holds information for a token of TAG type (i.e. a tag property).
type TagToken struct {
	BasicToken
	Handle string
	Suffix string
}

func (t TagToken) String() string { return t.Handle + t.Suffix }

// Scalar styles
const (
	AnyScalarStyle = iota
	PlainScalarStyle
	SingleQuotedScalarStyle
	DoubleQuotedScalarStyle
	LiteralScalarStyle
	FoldedScalarStyle
)

// A ScalarToken holds a value and the style as it appeared in the YAML document.
type ScalarToken struct {
	ValueToken
	Style int
}

// VersionDirective stores a %YAML directive.
type VersionDirective struct {
	BasicToken
	Major, Minor int
}

func (vd VersionDirective) String() string {
	return fmt.Sprintf("%%YAML %d.%d", vd.Major, vd.Minor)
}

// TagDirective stores a %TAG directive.
type TagDirective struct {
	BasicToken
	Handle string
	Prefix string
}

func (td TagDirective) String() string {
	return fmt.Sprintf("%%TAG %s %s", td.Handle, td.Prefix)
}
