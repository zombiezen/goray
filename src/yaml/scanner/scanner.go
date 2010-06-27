//
//	yaml/scanner/scanner.go
//	goray
//
//	Created by Ross Light on 2010-06-24.
//

package scanner

import (
	"container/list"
	"container/vector"
	"fmt"
	"io"
	"os"
	"unicode"
	"yaml/token"
)

type simpleKey struct {
	Possible bool
	Required bool
	Pos      token.Position
}

type Scanner struct {
	reader     *reader
	tokenQueue *list.List
	started    bool
	ended      bool

	indent      int
	indentStack vector.IntVector

	simpleKeyAllowed bool
	simpleKeyStack   vector.Vector

	flowLevel uint
}

func New(r io.Reader) (s *Scanner) {
	s = new(Scanner)
	s.reader = newReader(r)
	s.tokenQueue = list.New()
	return s
}

func (s *Scanner) Scan() (result Token, err os.Error) {
	if s.ended {
		return
	}
	if s.tokenQueue.Len() == 0 {
		if err = s.fetch(); err != nil {
			return
		}
		if s.tokenQueue.Len() == 0 {
			err = os.NewError("Fetch returned no tokens")
			return
		}
	}
	elem := s.tokenQueue.Front()
	result = elem.Value.(Token)
	s.tokenQueue.Remove(elem)
	return
}

func (s *Scanner) fetch() (err os.Error) {
	if !s.started {
		s.streamStart()
		return
	}

	if err = s.scanToNextToken(); err != nil {
		return
	}
	if err = s.removeStaleSimpleKeys(); err != nil {
		return
	}

	s.unrollIndent(s.reader.Pos.Column)

	if err = s.reader.Cache(4); err != nil {
		if err == os.EOF && !s.ended {
			err = s.streamEnd()
			return
		} else if err == io.ErrUnexpectedEOF {
			// Only have a handful of characters left?  That's cool, we handle
			// that.
			err = nil
		} else {
			return
		}
	}

	checkany := func(chars string) bool {
		currByte := s.reader.Bytes()[0]
		for _, c := range chars {
			if currByte == byte(c) {
				return true
			}
		}
		return false
	}
	blank := func(i int) bool {
		return unicode.IsSpace(int(s.reader.Bytes()[i]))
	}
	blankz := func(i int) bool {
		if s.reader.Len() <= i {
			return true
		}
		return blank(i)
	}

	switch {
	case s.reader.Pos.Column == 1 && s.reader.Check("%"):
		return s.fetchDirective()
	case s.reader.Pos.Column == 1 && s.reader.Check("---") && blankz(3):
		return s.fetchDocumentIndicator(token.DOCUMENT_START)
	case s.reader.Pos.Column == 1 && s.reader.Check("...") && blankz(3):
		return s.fetchDocumentIndicator(token.DOCUMENT_END)
	case s.reader.Check("["):
		return s.fetchFlowCollectionStart(token.FLOW_SEQUENCE_START)
	case s.reader.Check("{"):
		return s.fetchFlowCollectionStart(token.FLOW_MAPPING_START)
	case s.reader.Check("]"):
		return s.fetchFlowCollectionEnd(token.FLOW_SEQUENCE_END)
	case s.reader.Check("}"):
		return s.fetchFlowCollectionEnd(token.FLOW_MAPPING_END)
	case s.reader.Check(","):
		return s.fetchFlowEntry()
	case s.reader.Check("-") && blankz(2):
		return s.fetchBlockEntry()
	case s.reader.Check("?") && (s.flowLevel > 0 || blankz(2)):
		return s.fetchKey()
	case s.reader.Check(":") && (s.flowLevel > 0 || blankz(2)):
		return s.fetchValue()
	case s.reader.Check("*"):
		return s.fetchAnchor(token.ALIAS)
	case s.reader.Check("&"):
		return s.fetchAnchor(token.ANCHOR)
	case s.reader.Check("!"):
		return s.fetchTag()
	case s.reader.Check("|") && s.flowLevel == 0:
		return s.fetchBlockScalar(false)
	case s.reader.Check(">") && s.flowLevel == 1:
		return s.fetchBlockScalar(true)
	case s.reader.Check("'"):
		return s.fetchFlowScalar(false)
	case s.reader.Check("\""):
		return s.fetchFlowScalar(true)
	case !(blankz(0) || checkany("-?:,[]{}#&*!|>'\"%@`")) || (s.reader.Check("-") && !blank(1)) || (s.flowLevel == 0 && checkany("?:") && !blankz(1)):
		return s.fetchPlainScalar()
	default:
		err = os.NewError(fmt.Sprintf("Unrecognized token: %c", s.reader.Bytes()[0]))
	}
	return
}

func (s *Scanner) scanToNextToken() (err os.Error) {
	for {
		if err = s.reader.Cache(1); err != nil {
			return
		}

		// TODO: BOM

		// Eat whitespaces
		//
		// Tabs are allowed:
		// - in the flow context;
		// - in the block context, but not at the beginning of the line or
		//   after '-', '?', or ':' (complex value).
		for s.reader.Check(" ") || (s.reader.Check("\t") && (s.flowLevel > 0 || !s.simpleKeyAllowed)) {
			s.reader.Next(1)
			if err = s.reader.Cache(1); err != nil {
				return
			}
		}

		// Eat comment until end of line
		if s.reader.Check("#") {
			for !s.reader.CheckBreak() {
				s.reader.Next(1)
				if err = s.reader.Cache(1); err != nil {
					return
				}
			}
		}

		// If it's a line break, eat it.
		if s.reader.CheckBreak() {
			if err = s.reader.Cache(2); err != nil {
				return
			}
			s.reader.SkipBreak()
			// In the block context, a new line may start a simple key.
			if s.flowLevel == 0 {
				s.simpleKeyAllowed = true
			}
		} else {
			// We found a token.
			break
		}
	}
	return nil
}

func (s *Scanner) removeStaleSimpleKeys() (err os.Error) {
	for val := range s.simpleKeyStack.Iter() {
		key := val.(*simpleKey)

		// A simple key is:
		// - limited to a single line
		// - shorter than 1024 characters
		if key.Possible && (key.Pos.Line < s.reader.Pos.Line || key.Pos.Index+1024 < s.reader.Pos.Index) {
			if key.Required {
				return os.NewError("Could not find expected ':'")
			}
			key.Possible = false
		}
	}
	return
}

func (s *Scanner) saveSimpleKey() (err os.Error) {
	required := s.flowLevel == 0 && s.indent == s.reader.Pos.Column
	if s.simpleKeyAllowed {
		key := simpleKey{
			Possible: true,
			Required: required,
			Pos:      s.reader.Pos,
		}
		if err = s.removeSimpleKey(); err != nil {
			return
		}
		s.simpleKeyStack.Set(s.simpleKeyStack.Len()-1, &key)
	}
	return nil
}

func (s *Scanner) removeSimpleKey() (err os.Error) {
	key := s.simpleKeyStack.At(s.simpleKeyStack.Len() - 1).(*simpleKey)
	if key.Possible && key.Required {
		return os.NewError("Could not find expected ':'")
	}
	key.Possible = false
	return nil
}

func (s *Scanner) increaseFlowLevel() {
	s.simpleKeyStack.Push(new(simpleKey))
	s.flowLevel++
}

func (s *Scanner) decreaseFlowLevel() {
	if s.flowLevel > 0 {
		s.flowLevel--
		s.simpleKeyStack.Pop()
	}
}

func (s *Scanner) rollIndent(column, queueIndex int, kind token.Token, pos token.Position) {
	if s.flowLevel > 0 {
		return
	}

	if s.indent < column {
		// Push the current indentation level to the stack and set the new
		// indentation level.
		s.indentStack.Push(s.indent)
		s.indent = column
		tok := BasicToken{
			Kind:  kind,
			Start: pos,
			End:   pos,
		}
		if queueIndex == -1 {
			s.tokenQueue.PushBack(tok)
		} else {
			elem := s.tokenQueue.Front()
			for i := 0; i < queueIndex; i++ {
				elem = elem.Next()
			}
			s.tokenQueue.InsertBefore(tok, elem)
		}
	}
}

func (s *Scanner) unrollIndent(column int) {
	// In flow context, do nothing.
	if s.flowLevel > 0 {
		return
	}

	for s.indent > column {
		s.tokenQueue.PushBack(BasicToken{
			Kind:  token.BLOCK_END,
			Start: s.reader.Pos,
			End:   s.reader.Pos,
		})
		s.indent = s.indentStack.Pop()
	}
}

func (s *Scanner) streamStart() {
	s.indent = 0
	s.simpleKeyStack.Push(new(simpleKey))
	s.simpleKeyAllowed = true
	s.started = true
	s.tokenQueue.PushBack(BasicToken{
		Kind:  token.STREAM_START,
		Start: s.reader.Pos,
		End:   s.reader.Pos,
	})
}

func (s *Scanner) streamEnd() (err os.Error) {
	s.ended = true
	// Force new line
	if s.reader.Pos.Column != 1 {
		s.reader.Pos.Column = 1
		s.reader.Pos.Line++
	}
	// Reset indentation level
	s.unrollIndent(0)
	// Reset simple keys
	if err = s.removeSimpleKey(); err != nil {
		return
	}
	s.simpleKeyAllowed = false
	// End the stream
	s.tokenQueue.PushBack(BasicToken{
		Kind:  token.STREAM_END,
		Start: s.reader.Pos,
		End:   s.reader.Pos,
	})
	return nil
}

func (s *Scanner) fetchDirective() (err os.Error) {
	// Reset indentation level
	s.unrollIndent(0)
	// Reset simple keys
	if err = s.removeSimpleKey(); err != nil {
		return
	}
	s.simpleKeyAllowed = false
	// TODO: Create token
	return
}

func (s *Scanner) fetchDocumentIndicator(kind token.Token) (err os.Error) {
	// Reset indentation level
	s.unrollIndent(0)
	// Reset simple keys
	if err = s.removeSimpleKey(); err != nil {
		return
	}
	s.simpleKeyAllowed = false
	// Consume the token
	startPos := s.reader.Pos
	s.reader.Next(3)
	endPos := s.reader.Pos
	// Create the scanner token
	s.tokenQueue.PushBack(BasicToken{
		Kind:  kind,
		Start: startPos,
		End:   endPos,
	})
	return
}

func (s *Scanner) fetchFlowCollectionStart(kind token.Token) (err os.Error) {
	return
}

func (s *Scanner) fetchFlowCollectionEnd(kind token.Token) (err os.Error) {
	return
}

func (s *Scanner) fetchFlowEntry() (err os.Error) {
	return
}

func (s *Scanner) fetchBlockEntry() (err os.Error) {
	return
}

func (s *Scanner) fetchKey() (err os.Error) {
	return
}

func (s *Scanner) fetchValue() (err os.Error) {
	return
}

func (s *Scanner) fetchAnchor(kind token.Token) (err os.Error) {
	return
}

func (s *Scanner) fetchTag() (err os.Error) {
	return
}

func (s *Scanner) fetchBlockScalar(folded bool) (err os.Error) {
	return
}

func (s *Scanner) fetchFlowScalar(doubleQuoted bool) (err os.Error) {
	return
}

func (s *Scanner) fetchPlainScalar() (err os.Error) {
	return
}
