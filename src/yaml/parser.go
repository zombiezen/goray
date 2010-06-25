//
//	yaml/parser.go
//	goray
//
//	Created by Ross Light on 2010-06-24.
//

package yaml

import (
	"io"
	"os"
)

const (
	parser_STREAM_START_STATE = iota
	parser_IMPLICIT_DOCUMENT_START_STATE
	parser_DOCUMENT_START_STATE
	parser_DOCUMENT_CONTENT_STATE
	parser_DOCUMENT_END_STATE
	parser_BLOCK_NODE_STATE
	parser_BLOCK_NODE_OR_INDENTLESS_SEQUENCE_STATE
	parser_FLOW_NODE_STATE
	parser_BLOCK_SEQUENCE_FIRST_ENTRY_STATE
	parser_BLOCK_SEQUENCE_ENTRY_STATE
	parser_INDENTLESS_SEQUENCE_ENTRY_STATE
	parser_BLOCK_MAPPING_FIRST_KEY_STATE
	parser_BLOCK_MAPPING_KEY_STATE
	parser_BLOCK_MAPPING_VALUE_STATE
	parser_FLOW_SEQUENCE_FIRST_ENTRY_STATE
	parser_FLOW_SEQUENCE_ENTRY_STATE
	parser_FLOW_SEQUENCE_ENTRY_MAPPING_KEY_STATE
	parser_FLOW_SEQUENCE_ENTRY_MAPPING_VALUE_STATE
	parser_FLOW_SEQUENCE_ENTRY_MAPPING_END_STATE
	parser_FLOW_MAPPING_FIRST_KEY_STATE
	parser_FLOW_MAPPING_KEY_STATE
	parser_FLOW_MAPPING_VALUE_STATE
	parser_FLOW_MAPPING_EMPTY_VALUE_STATE
	parser_END_STATE
)

type Parser struct {
	reader io.Reader
	state int
}

func (p *Parser) Next() (Event, os.Error) {
	switch state {
	case parser_STREAM_START_STATE:
		return p.parseStreamStart()
	case parser_IMPLICIT_DOCUMENT_START_STATE:
	case parser_DOCUMENT_START_STATE:
	case parser_DOCUMENT_CONTENT_STATE:
	case parser_DOCUMENT_END_STATE:
	case parser_BLOCK_NODE_STATE:
	case parser_BLOCK_NODE_OR_INDENTLESS_SEQUENCE_STATE:
	case parser_FLOW_NODE_STATE:
	case parser_BLOCK_SEQUENCE_FIRST_ENTRY_STATE:
	case parser_BLOCK_SEQUENCE_ENTRY_STATE:
	case parser_INDENTLESS_SEQUENCE_ENTRY_STATE:
	case parser_BLOCK_MAPPING_FIRST_KEY_STATE:
	case parser_BLOCK_MAPPING_KEY_STATE:
	case parser_BLOCK_MAPPING_VALUE_STATE:
	case parser_FLOW_SEQUENCE_FIRST_ENTRY_STATE:
	case parser_FLOW_SEQUENCE_ENTRY_STATE:
	case parser_FLOW_SEQUENCE_ENTRY_MAPPING_KEY_STATE:
	case parser_FLOW_SEQUENCE_ENTRY_MAPPING_VALUE_STATE:
	case parser_FLOW_SEQUENCE_ENTRY_MAPPING_END_STATE:
	case parser_FLOW_MAPPING_FIRST_KEY_STATE:
	case parser_FLOW_MAPPING_KEY_STATE:
	case parser_FLOW_MAPPING_VALUE_STATE:
	case parser_FLOW_MAPPING_EMPTY_VALUE_STATE:
	case parser_END_STATE:
		return nil, nil
	}
	
	return nil, os.NewError("Parser entered unknown state")
}

func (p *Parser) parseStreamStart() (Event, os.Error) {
}

func Parse(r io.Reader) <-chan Event {
	ch := make(chan Event)
	go func() {
		parser := NewParser(r)
		for parser.state != parser_END_STATE {
			evt, _ := p.Next()
			ch <- evt
		}
	}()
	return ch
}
