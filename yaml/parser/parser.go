//
//	parser/parser.go
//	goyaml
//
//	Created by Ross Light on 2010-07-02.
//

/*
	The parser package takes a sequence of tokens and transforms it into a
	representation graph (and optionally into native data structures).  This
	corresponds to the composing and constructing stages in the YAML 1.2
	specification.
*/
package parser

import (
	"bitbucket.org/zombiezen/goray/yaml/scanner"
	"bitbucket.org/zombiezen/goray/yaml/token"
	"container/list"
	"errors"
	"fmt"
	"io"
)

// An Error is a high-level parser error, including the node.
type Error struct {
	Err  error
	Node Node
}

func (err Error) Error() string {
	return fmt.Sprintf("%v: %v", err.Node.Start(), err.Err)
}

func wrapError(origErr error, n Node) (err Error) {
	var ok bool
	if err, ok = origErr.(Error); ok {
		return
	}
	return Error{origErr, n}
}

// A Schema determines the tag for a node without an explicit tag.
type Schema interface {
	Resolve(node Node, userData interface{}) (tag string, err error)
}

// A Constructor converts a Node into a Go data structure.
type Constructor interface {
	Construct(node Node, userData interface{}) (data interface{}, err error)
}

const DefaultPrefix = "tag:yaml.org,2002:"

// A Parser transforms a YAML stream into data structures.
type Parser struct {
	scanner     *scanner.Scanner
	scanQueue   *list.List
	doc         *Document
	tagPrefixes map[string]string
	anchors     map[string]Node
	schema      Schema
	constructor Constructor
	userData    interface{}
}

// New creates a new parser that reads from the given reader.
func New(r io.Reader, s Schema, c Constructor, userData interface{}) *Parser {
	return NewWithScanner(scanner.New(r), s, c, userData)
}

// NewWithScanner creates a new parser that reads from an existing scanner.
func NewWithScanner(scan *scanner.Scanner, schema Schema, con Constructor, userData interface{}) (p *Parser) {
	p = new(Parser)
	p.scanner = scan
	p.scanQueue = list.New()
	p.schema = schema
	p.constructor = con
	p.userData = userData
	return p
}

// reset reverts document-specific variables to their initial state.
func (parser *Parser) reset() {
	parser.tagPrefixes = make(map[string]string)
	parser.tagPrefixes["!!"] = DefaultPrefix
	parser.anchors = make(map[string]Node)
}

func (parser *Parser) scan() (tok scanner.Token, err error) {
	if parser.scanQueue.Len() > 0 {
		// We have some tokens queued up
		elem := parser.scanQueue.Front()
		tok = elem.Value.(scanner.Token)
		parser.scanQueue.Remove(elem)
		return
	}
	return parser.scanner.Scan()
}

func (parser *Parser) peek() (tok scanner.Token, err error) {
	if parser.scanQueue.Len() > 0 {
		tok = parser.scanQueue.Front().Value.(scanner.Token)
		return
	}
	tok, err = parser.scanner.Scan()
	if err == nil {
		parser.unscan(tok)
	}
	return
}

func (parser *Parser) unscan(tok scanner.Token) {
	parser.scanQueue.PushFront(tok)
}

// Parse returns all of the documents in the stream.
func (parser *Parser) Parse() (docs []*Document, err error) {
	docs = make([]*Document, 0, 1)

docLoop:
	for {
		var newDoc *Document

		// Parse document
		newDoc, err = parser.ParseDocument()
		if err != nil {
			return
		}

		// Add document to array
		if cap(docs) < len(docs)+1 {
			newDocs := make([]*Document, len(docs), cap(docs)*2)
			copy(newDocs, docs)
			docs = newDocs
		}
		docs = docs[0 : len(docs)+1]
		docs[len(docs)-1] = newDoc

		// Eat up any end document tokens
	endDocLoop:
		for {
			var tok scanner.Token

			if tok, err = parser.scan(); err != nil {
				return
			}

			switch tok.GetKind() {
			case token.DOCUMENT_END:
				// Ignore
			case token.STREAM_END:
				break docLoop
			default:
				parser.unscan(tok)
				break endDocLoop
			}
		}
	}
	return
}

// ParseDocument returns the next document in the stream.
func (parser *Parser) ParseDocument() (doc *Document, err error) {
	if parser.doc != nil {
		err = errors.New("ParseDocument called in the middle of parsing another document")
		return
	}

	doc = new(Document)
	doc.MajorVersion, doc.MinorVersion = 1, 2
	parser.doc = doc
	parser.reset()

	if err = parser.parseDocHeader(); err != nil {
		return
	}
	if doc.Content, err = parser.parseNode(); err != nil {
		return
	}
	parser.doc = nil
	return
}

func (parser *Parser) parseDocHeader() (err error) {
	startedHeader := false

	for {
		var tok scanner.Token
		tok, err = parser.scan()
		if err != nil {
			return
		}

		switch tok.GetKind() {
		case token.STREAM_START:
			// Ignore
		case token.VERSION_DIRECTIVE:
			startedHeader = true
			versionDir := tok.(scanner.VersionDirective)
			parser.doc.MajorVersion, parser.doc.MinorVersion = versionDir.Major, versionDir.Minor
		case token.TAG_DIRECTIVE:
			startedHeader = true
			if err = parser.handleTagDirective(tok.(scanner.TagDirective)); err != nil {
				return
			}
		case token.DOCUMENT_START:
			return
		case token.DOCUMENT_END:
			if startedHeader {
				err = errors.New("End document token appeared after directives")
				return
			}
		default:
			parser.unscan(tok)
			return
		}
	}

	// This should never be reached.
	return nil
}

func (parser *Parser) handleTagDirective(directive scanner.TagDirective) (err error) {
	if _, hasPrefix := parser.tagPrefixes[directive.Handle]; hasPrefix {
		err = errors.New("Tag directive would redefine " + directive.Handle)
		return
	}
	parser.tagPrefixes[directive.Handle] = directive.Prefix
	return
}

func (parser *Parser) parseNode() (node Node, err error) {
	anchor, tag, err := parser.parseNodeProperties()
	if err != nil {
		return
	}

	tok, err := parser.peek()
	if err != nil {
		return
	}

	switch tok.GetKind() {
	case token.SCALAR:
		node, err = parser.parseScalar()
	case token.BLOCK_SEQUENCE_START:
		node, err = parser.parseSequence(true)
	case token.FLOW_SEQUENCE_START:
		node, err = parser.parseSequence(false)
	case token.BLOCK_MAPPING_START:
		node, err = parser.parseMapping(true)
	case token.FLOW_MAPPING_START:
		node, err = parser.parseMapping(false)
	case token.ALIAS:
		parser.scan() // Eat alias token
		return parser.resolveAlias(tok.String())
	case token.ANCHOR, token.TAG:
		err = errors.New("Invalid node properties")
		return
	default:
		// Create an empty node
		node = &Empty{new(basicNode)}
		node.(*Empty).pos = tok.GetStart()
	}

	if err == nil {
		var data interface{}

		// Set up tag
		if tag == "" {
			tag, err = parser.schema.Resolve(node, parser.userData)
			if err != nil {
				err = wrapError(err, node)
				return
			}
		}
		node.setTag(tag)

		// Construct data
		data, err = parser.constructor.Construct(node, parser.userData)
		if err != nil {
			err = wrapError(err, node)
			return
		}
		node.setData(data)

		if anchor != "" {
			parser.anchors[anchor] = node
		}
	} else if node != nil {
		err = wrapError(err, node)
	}

	return
}

func (parser *Parser) parseNodeProperties() (anchor, tag string, err error) {
	var tok, tagTok scanner.Token

	if tok, err = parser.scan(); err != nil {
		return
	}

	if tok.GetKind() == token.ANCHOR {
		anchor = tok.String()

		if tok, err = parser.peek(); err != nil {
			return
		}
		if tok.GetKind() == token.TAG {
			tagTok = tok
			parser.scan()
		}
	} else if tok.GetKind() == token.TAG {
		tagTok = tok

		if tok, err = parser.peek(); err != nil {
			return
		}
		if tok.GetKind() == token.ANCHOR {
			anchor = tok.String()
			parser.scan()
		}
	} else {
		parser.unscan(tok)
	}

	if tagTok != nil {
		tag, err = parser.resolveTag(tagTok.(scanner.TagToken))
	}

	return
}

func (parser *Parser) resolveTag(tok scanner.TagToken) (tag string, err error) {
	if tok.Handle == "" {
		return tok.Suffix, nil
	}
	if prefix, ok := parser.tagPrefixes[tok.Handle]; ok {
		return prefix + tok.Suffix, nil
	}

	err = errors.New("Unrecognized tag handle: " + tok.Handle)
	return
}

func (parser *Parser) resolveAlias(anchor string) (node Node, err error) {
	node, ok := parser.anchors[anchor]
	if !ok {
		err = errors.New("Invalid alias: " + anchor)
	}
	return
}

func (parser *Parser) parseScalar() (node *Scalar, err error) {
	tok, err := parser.scan()
	if err != nil {
		return
	}

	node = new(Scalar)
	node.basicNode = new(basicNode)
	node.pos = tok.GetStart()
	node.Value = tok.String()
	return
}

func (parser *Parser) parseSequence(block bool) (node *Sequence, err error) {
	var tok scanner.Token
	var sep, sentinel token.Token

	// Discard opening token
	if tok, err = parser.scan(); err != nil {
		return
	}

	if block {
		sep, sentinel = token.BLOCK_ENTRY, token.BLOCK_END
	} else {
		sep, sentinel = token.FLOW_ENTRY, token.FLOW_SEQUENCE_END
	}

	node = new(Sequence)
	node.basicNode = new(basicNode)
	node.pos = tok.GetStart()
	node.Nodes = make([]Node, 0, 2)
	childCount := 0

	// Block sequences have a leading separator
	if block {
		tok, err = parser.scan()
		if tok.GetKind() != sep {
			err = errors.New("Expected leading block entry token")
			return
		}
	}

	for err == nil && tok.GetKind() != sentinel {
		var child Node

		// Scan child
		if child, err = parser.parseNode(); err != nil {
			return
		}

		// Add child to node
		if cap(node.Nodes) < childCount+1 {
			newNodes := make([]Node, childCount, cap(node.Nodes)*2)
			copy(newNodes, node.Nodes)
			node.Nodes = newNodes
		}
		node.Nodes = node.Nodes[0 : childCount+1]
		node.Nodes[childCount] = child
		childCount++

		// Check for separator
		tok, err = parser.scan()
		if err != nil {
			return
		} else if tok.GetKind() != sep && tok.GetKind() != sentinel {
			err = errors.New("Unexpected token in sequence: " + tok.String())
			return
		}
	}
	return
}

func (parser *Parser) parseMapping(block bool) (node *Mapping, err error) {
	var tok scanner.Token
	var sentinel token.Token

	// Discard opening token
	if tok, err = parser.scan(); err != nil {
		return
	}

	if block {
		sentinel = token.BLOCK_END
	} else {
		sentinel = token.FLOW_MAPPING_END
	}

	node = new(Mapping)
	node.basicNode = new(basicNode)
	node.pos = tok.GetStart()
	node.Pairs = make([]KeyValuePair, 0, 2)
	pairCount := 0

	for err == nil && tok.GetKind() != sentinel {
		var pair KeyValuePair

		// Scan key
		tok, err = parser.scan()
		if err != nil {
			return
		} else if tok.GetKind() != token.KEY {
			err = errors.New("Expected key")
			return
		}
		pair.Key, err = parser.parseNode()
		if err != nil {
			return
		}

		// Scan value
		tok, err = parser.scan()
		if err != nil {
			return
		} else if tok.GetKind() != token.VALUE {
			err = errors.New("Expected value")
			return
		}
		pair.Value, err = parser.parseNode()
		if err != nil {
			return
		}

		// Add pair to node
		if cap(node.Pairs) < pairCount+1 {
			newPairs := make([]KeyValuePair, pairCount, cap(node.Pairs)*2)
			copy(newPairs, node.Pairs)
			node.Pairs = newPairs
		}
		node.Pairs = node.Pairs[0 : pairCount+1]
		node.Pairs[pairCount] = pair
		pairCount++

		// Find closure
		// Wouldn't it be great if it were always this easy?
		tok, err = parser.scan()
		if err != nil {
			return
		} else if !(tok.GetKind() == sentinel || (!block && tok.GetKind() == token.FLOW_ENTRY) || (block && tok.GetKind() == token.KEY)) {
			err = errors.New("Unexpected token in mapping: " + tok.GetKind().String() + " " + tok.String())
			return
		}

		if tok.GetKind() == token.KEY {
			parser.unscan(tok)
		}
	}

	return
}
