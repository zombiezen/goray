//
//	yaml/token/token.go
//	goray
//
//	Created by Ross Light on 2010-06-24.
//

package token

import "fmt"

const (
	NO_TOKEN = iota

	STREAM_START
	STREAM_END

	VERSION_DIRECTIVE
	TAG_DIRECTIVE
	DOCUMENT_START
	DOCUMENT_END

	BLOCK_SEQUENCE_START
	BLOCK_MAPPING_START
	BLOCK_END

	FLOW_SEQUENCE_START
	FLOW_SEQUENCE_END
	FLOW_MAPPING_START
	FLOW_MAPPING_END

	BLOCK_ENTRY
	FLOW_ENTRY
	KEY
	VALUE

	ALIAS
	ANCHOR
	TAG
	SCALAR
)

type Token int

func (t Token) String() string {
	switch t {
	case NO_TOKEN:
		return "NO_TOKEN"
	case STREAM_START:
		return "STREAM_START"
	case STREAM_END:
		return "STREAM_END"
	case VERSION_DIRECTIVE:
		return "VERSION_DIRECTIVE"
	case TAG_DIRECTIVE:
		return "TAG_DIRECTIVE"
	case DOCUMENT_START:
		return "DOCUMENT_START"
	case DOCUMENT_END:
		return "DOCUMENT_END"
	case BLOCK_SEQUENCE_START:
		return "BLOCK_SEQUENCE_START"
	case BLOCK_MAPPING_START:
		return "BLOCK_MAPPING_START"
	case BLOCK_END:
		return "BLOCK_END"
	case FLOW_SEQUENCE_START:
		return "FLOW_SEQUENCE_START"
	case FLOW_SEQUENCE_END:
		return "FLOW_SEQUENCE_END"
	case FLOW_MAPPING_START:
		return "FLOW_MAPPING_START"
	case FLOW_MAPPING_END:
		return "FLOW_MAPPING_END"
	case BLOCK_ENTRY:
		return "BLOCK_ENTRY"
	case FLOW_ENTRY:
		return "FLOW_ENTRY"
	case KEY:
		return "KEY"
	case VALUE:
		return "VALUE"
	case ALIAS:
		return "ALIAS"
	case ANCHOR:
		return "ANCHOR"
	case TAG:
		return "TAG"
	case SCALAR:
		return "SCALAR"
	}
	return fmt.Sprintf("Token(%d)", int(t))
}

type Position struct {
	Index  int
	Line   int
	Column int
}

func (pos Position) String() string {
	return fmt.Sprintf("%d:%d", pos.Line, pos.Column)
}
