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

type Token interface {
	GetKind() token.Token
	GetStart() token.Position
	GetEnd() token.Position
	String() string
}

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

/* ValueToken defines a token that holds a string value.  It is used for anchors, aliases, and scalars. */
type ValueToken struct {
	BasicToken
	Value string
}

func (t ValueToken) String() string { return t.Value }

type TagToken struct {
	BasicToken
	Handle string
	Suffix string
}

type ScalarToken struct {
	ValueToken
	// TODO: Style
}

type VersionDirective struct {
	BasicToken
	Major, Minor int
}

func (vd VersionDirective) String() string {
	return fmt.Sprintf("%%YAML %d.%d", vd.Major, vd.Minor)
}

type TagDirective struct {
	BasicToken
	Handle string
	Prefix string
}

func (td TagDirective) String() string {
	return fmt.Sprintf("%%TAG %s %s", td.Handle, td.Prefix)
}
