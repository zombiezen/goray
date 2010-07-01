//
//	yaml/scanner/scan.go
//	goray
//
//	Created by Ross Light on 2010-06-29.
//

package scanner

import (
	"bytes"
	"io"
	"os"
	"yaml/token"
)

func (s *Scanner) scanToNextToken() (err os.Error) {
	for {
		if err = s.reader.Cache(1); err != nil {
			if err == io.ErrUnexpectedEOF {
				// Our "next" token is the end of the stream.  This isn't a failure.
				err = nil
			}
			return
		}

		// TODO: BOM

		// Eat whitespaces
		//
		// Tabs are allowed:
		// - in the flow context;
		// - in the block context, but not at the beginning of the line or
		//   after '-', '?', or ':' (complex value).
		for s.reader.Check(0, " ") || (s.reader.Check(0, "\t") && (s.flowLevel > 0 || !s.simpleKeyAllowed)) {
			s.reader.Next(1)
			if err = s.reader.Cache(1); err != nil {
				if err == io.ErrUnexpectedEOF {
					err = nil
				}
				return
			}
		}

		// Eat comment until end of line
		if s.reader.Check(0, "#") {
			for !s.reader.CheckBreak(0) {
				s.reader.Next(1)
				if err = s.reader.Cache(1); err != nil {
					if err == io.ErrUnexpectedEOF {
						err = nil
					}
					return
				}
			}
		}

		// If it's a line break, eat it.
		if s.reader.CheckBreak(0) {
			s.reader.ReadBreak()
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

func (s *Scanner) scanDirective() (tok Token, err os.Error) {
	startPos := s.reader.Pos
	// Eat '%'
	s.reader.Next(1)
	// Scan the name
	name, err := s.scanDirectiveName()
	if err != nil {
		return
	}

	switch name {
	case "YAML":
		var major, minor int
		major, minor, err = s.scanVersionDirectiveValue()
		if err != nil {
			return
		}
		t := VersionDirective{}
		t.Kind = token.VERSION_DIRECTIVE
		t.Start, t.End = startPos, s.reader.Pos
		t.Major, t.Minor = major, minor
		tok = t
	case "TAG":
		var handle, prefix string
		handle, prefix, err = s.scanTagDirectiveValue()
		if err != nil {
			return
		}
		t := TagDirective{}
		t.Kind = token.TAG_DIRECTIVE
		t.Start, t.End = startPos, s.reader.Pos
		t.Handle, t.Prefix = handle, prefix
		tok = t
	default:
		err = os.NewError("Unrecognized directive: " + name)
		return
	}

	// Eat the rest of the line, including comments
	s.reader.SkipSpaces()
	if err = s.reader.Cache(1); err != nil {
		return
	}
	if s.reader.Check(0, "#") {
		for s.reader.Len() != 0 && !s.reader.CheckBreak(0) {
			s.reader.Next(1)
			if err = s.reader.Cache(1); err != nil {
				return
			}
		}
	}
	if !s.reader.CheckBreak(0) {
		err = os.NewError("Directive did not end with comment or line break")
		return
	}
	s.reader.ReadBreak()
	return tok, nil
}

func (s *Scanner) scanDirectiveName() (name string, err os.Error) {
	nameBuf := new(bytes.Buffer)
	if err = s.reader.Cache(1); err != nil {
		return
	}

	for isLetter(s.reader.Bytes()[0]) {
		if _, err = io.Copyn(nameBuf, s.reader, 1); err != nil {
			return
		}
		if err = s.reader.Cache(1); err != nil {
			return
		}
	}

	if nameBuf.Len() == 0 {
		err = os.NewError("Directive name not found")
		return
	}

	if !isWhitespace(s.reader.Bytes()[0]) {
		err = os.NewError("Unexpected non-alphabetical character")
		return
	}

	name = nameBuf.String()
	return
}

func (s *Scanner) scanVersionDirectiveValue() (major, minor int, err os.Error) {
	scanNumber := func() (n int, err os.Error) {
		numLen := 0
		if err = s.reader.Cache(1); err != nil {
			return
		}
		for s.reader.CheckDigit(0) && numLen < 9 {
			b, _ := s.reader.ReadByte()
			n = n*10 + int(b-'0')
			numLen++
			if err = s.reader.Cache(1); err != nil {
				return 0, err
			}
		}

		if numLen == 0 {
			err = os.NewError("No number found")
		}
		return
	}
	s.reader.SkipSpaces()
	if major, err = scanNumber(); err != nil {
		return
	}
	if !s.reader.Check(0, ".") {
		err = os.NewError("Did not find .")
		return
	}
	s.reader.Next(1)
	if minor, err = scanNumber(); err != nil {
		return
	}
	return
}

func (s *Scanner) scanTagDirectiveValue() (handle, prefix string, err os.Error) {
	// TODO
	return
}

func (s *Scanner) scanAnchor(kind token.Token) (tok Token, err os.Error) {
	startPos := s.reader.Pos
	// Eat indicator
	s.reader.Next(1)
	// Consume value
	valueBuf := new(bytes.Buffer)
	if err = s.reader.Cache(1); err != nil {
		return
	}
	for s.reader.CheckLetter(0) {
		b, _ := s.reader.ReadByte()
		valueBuf.WriteByte(b)
		if err = s.reader.Cache(1); err != nil {
			return
		}
	}
	if valueBuf.Len() == 0 || !(s.reader.Len() == 0 || s.reader.CheckBreak(0) || s.reader.CheckSpace(0) || s.reader.CheckAny(0, "?:,]}%@`")) {
		err = os.NewError("Did not find expected alphabetic or numeric character")
		return
	}
	// Create token
	tok = ValueToken{
		BasicToken{
			Kind:  kind,
			Start: startPos,
			End:   s.reader.Pos,
		},
		valueBuf.String(),
	}
	return
}

func (s *Scanner) scanFlowScalar(style int) (tok Token, err os.Error) {
	startPos := s.reader.Pos
	valueBuf := new(bytes.Buffer)
	// Eat the left quote
	s.reader.Next(1)
	// Consume content
	for {
		leadingBlanks := false

		if err = s.reader.Cache(4); err != nil && (err != io.ErrUnexpectedEOF || s.reader.Len() == 0) {
			return
		}
		err = nil

		// Better not be a document indicator.
		if s.reader.Pos.Column == 1 && (s.reader.Check(0, "---") || s.reader.Check(0, "...")) && s.reader.CheckBlank(3) {
			err = os.NewError("Unexpected end of document")
			return
		}
		// Consume non-blanks
	nonBlankLoop:
		for !s.reader.CheckBlank(0) {
			switch {
			// Escaped single quote
			case style == SingleQuotedScalarStyle && s.reader.Check(0, "''"):
				valueBuf.WriteByte('\'')
				s.reader.Next(2)
			// Right quote
			case (style == SingleQuotedScalarStyle && s.reader.Check(0, "'")) || (style == DoubleQuotedScalarStyle && s.reader.Check(0, "\"")):
				break nonBlankLoop
			// Escaped line break
			case style == DoubleQuotedScalarStyle && s.reader.Check(0, "\\") && s.reader.CheckBreak(1):
				s.reader.Next(1)
				s.reader.ReadBreak()
				leadingBlanks = false
				break nonBlankLoop
			// Escape sequence
			case style == DoubleQuotedScalarStyle && s.reader.Check(0, "\\"):
				var rune int
				if rune, err = s.scanEscapeSeq(); err == nil {
					valueBuf.WriteRune(rune)
				} else {
					return
				}
			// Normal character
			default:
				b, _ := s.reader.ReadByte()
				valueBuf.WriteByte(b)
			}

			// Get ready for next non-blank
			if err = s.reader.Cache(2); err != nil && (err != io.ErrUnexpectedEOF || s.reader.Len() == 0) {
				return
			}
		}

		// Are we at the end of the scalar?
		if (style == SingleQuotedScalarStyle && s.reader.Check(0, "'")) || (style == DoubleQuotedScalarStyle && s.reader.Check(0, "\"")) {
			break
		}

		// Consume blank characters
		whitespaces := new(bytes.Buffer)
		leadingBreak := new(bytes.Buffer)
		trailingBreaks := new(bytes.Buffer)
		if err = s.reader.Cache(1); err != nil {
			return
		}
		for s.reader.CheckBlank(0) || s.reader.CheckBreak(0) {
			if s.reader.CheckBlank(0) {
				if !leadingBlanks {
					b, _ := s.reader.ReadByte()
					whitespaces.WriteByte(b)
				} else {
					s.reader.Next(1)
				}
			} else {
				var bytes []byte
				if !leadingBlanks {
					whitespaces.Reset()
					if bytes, err = s.reader.ReadBreak(); err == nil {
						leadingBreak.Write(bytes)
					} else {
						return
					}
				} else {
					if bytes, err = s.reader.ReadBreak(); err == nil {
						trailingBreaks.Write(bytes)
					} else {
						return
					}
				}
			}
			if err = s.reader.Cache(1); err != nil {
				return
			}
		}

		// Join the whitespaces or fold line breaks
		if leadingBlanks {
			if leadingBreak.Bytes()[0] == '\n' {
				// We need to fold line breaks
				if trailingBreaks.Len() == 0 {
					valueBuf.WriteByte(' ')
				} else {
					io.Copy(valueBuf, trailingBreaks)
				}
			} else {
				io.Copy(valueBuf, leadingBreak)
				io.Copy(valueBuf, trailingBreaks)
				leadingBreak.Reset()
				trailingBreaks.Reset()
			}
		} else {
			io.Copy(valueBuf, whitespaces)
		}
	}

	// Eat the right quote
	s.reader.Next(1)
	// Create token
	{
		scalarTok := ScalarToken{}
		scalarTok.Kind = token.SCALAR
		scalarTok.Start, scalarTok.End = startPos, s.reader.Pos
		scalarTok.Style = style
		scalarTok.Value = valueBuf.String()
		tok = scalarTok
	}
	err = nil
	return
}

func (s *Scanner) scanEscapeSeq() (rune int, err os.Error) {
	if err = s.reader.Cache(2); err != nil {
		return
	}
	codeLength := 0
	switch s.reader.Bytes()[1] {
	case '0':
		rune = '\x00'
	case 'a':
		rune = '\x07'
	case 't', '\t':
		rune = '\x09'
	case 'n':
		rune = '\x0A'
	case 'v':
		rune = '\x0B'
	case 'f':
		rune = '\x0C'
	case 'r':
		rune = '\x0D'
	case 'e':
		rune = '\x1B'
	case ' ':
		rune = '\x20'
	case '"':
		rune = '"'
	case '\'':
		rune = '\''
	case '\\':
		rune = '\\'
	// NEL (#x85)
	case 'N':
		rune = '\u0085'
	// #xA0
	case '_':
		rune = '\u00A0'
	// LS (#x2028)
	case 'L':
		rune = '\u2028'
	// PS (#x2029)
	case 'P':
		rune = '\u2029'
	case 'x':
		codeLength = 2
	case 'u':
		codeLength = 4
	case 'U':
		codeLength = 8
	default:
		err = os.NewError("Unrecognized escape sequence")
		return
	}
	s.reader.Next(2)

	// Is this a hex escape?
	if codeLength > 0 {
		// Scan character value
		if err = s.reader.Cache(codeLength); err != nil {
			return
		}
		for k := 0; k < codeLength; k++ {
			b, _ := s.reader.ReadByte()
			if digit, ok := asHex(b); ok {
				rune = (rune << 4) | digit
			} else {
				err = os.NewError("Did not find expected hex digit")
				return
			}
		}
		// Check value
		if (rune >= 0xD800 && rune <= 0xDFFF) || rune > 0x10FFFF {
			err = os.NewError("Found invalid unicode character")
			return
		}
	}
	return
}

func (s *Scanner) scanPlainScalar() (tok Token, err os.Error) {
	startPos, endPos := s.reader.Pos, s.reader.Pos
	valueBuf := new(bytes.Buffer)
	indent := s.indent + 1

	// Set up buffers
	leadingBlanks := false
	whitespaces := new(bytes.Buffer)
	leadingBreak := new(bytes.Buffer)
	trailingBreaks := new(bytes.Buffer)

	for {
		if err = s.reader.Cache(4); err != nil && (err != io.ErrUnexpectedEOF || s.reader.Len() == 0) {
			return
		}
		err = nil
		// Stop on a document indicator
		if s.reader.Pos.Column == 1 && (s.reader.Check(0, "---") || s.reader.Check(0, "...")) && s.reader.CheckBlank(3) {
			break
		}
		// Check for a comment
		if s.reader.Check(0, "#") {
			break
		}
		// Consume non-blanks
		for !s.reader.CheckBlank(0) {
			// Check for 'x:x' in the flow context
			if s.flowLevel > 0 && s.reader.Check(0, ":") && !s.reader.CheckBlank(1) {
				err = os.NewError("Found unexpected ':'")
				return
			}
			// Check for indicators that may end a plain scalar
			if (s.reader.Check(0, ":") && s.reader.CheckBlank(1)) || (s.flowLevel > 0 && s.reader.CheckAny(0, ",:?[]{}")) {
				break
			}
			// Check if we need to join whitespaces and breaks
			if leadingBlanks {
				// Do we need to fold line breaks?
				if leadingBreak.Bytes()[0] == '\n' {
					// We need to fold line breaks
					if trailingBreaks.Len() == 0 {
						valueBuf.WriteByte(' ')
					} else {
						io.Copy(valueBuf, trailingBreaks)
					}
					leadingBreak.Reset()
				} else {
					io.Copy(valueBuf, leadingBreak)
					io.Copy(valueBuf, trailingBreaks)
				}

				leadingBlanks = false
			} else if whitespaces.Len() > 0 {
				io.Copy(valueBuf, whitespaces)
			}
			// Copy the character
			b, _ := s.reader.ReadByte()
			valueBuf.WriteByte(b)
			endPos = s.reader.Pos
			if err = s.reader.Cache(2); err != nil && err != io.ErrUnexpectedEOF {
				return
			}
		}

		// Is it the end?
		if !(s.reader.CheckSpace(0) || s.reader.CheckBreak(0)) {
			break
		}

		// Consume blank characters
		if err = s.reader.Cache(1); err != nil {
			return
		}
		for s.reader.CheckSpace(0) || s.reader.CheckBreak(0) {
			if s.reader.CheckSpace(0) {
				// Check for abusive tabs
				if leadingBlanks && s.reader.Pos.Column < indent && s.reader.Check(0, "\t") {
					err = os.NewError("Found a tab character that violates indentation")
					return
				}

				if !leadingBlanks {
					b, _ := s.reader.ReadByte()
					whitespaces.WriteByte(b)
				} else {
					s.reader.Next(1)
				}
			} else {
				var bytes []byte
				if !leadingBlanks {
					whitespaces.Reset()
					if bytes, err = s.reader.ReadBreak(); err == nil {
						leadingBreak.Write(bytes)
						leadingBlanks = true
					} else {
						return
					}
				} else {
					if bytes, err = s.reader.ReadBreak(); err == nil {
						trailingBreaks.Write(bytes)
					} else {
						return
					}
				}
			}
			if err = s.reader.Cache(1); err != nil {
				return
			}
		}

		// Check indentation level
		if s.flowLevel == 0 && s.reader.Pos.Column < indent {
			break
		}
	}

	// Create token
	{
		scalarTok := ScalarToken{}
		scalarTok.Kind = token.SCALAR
		scalarTok.Start, scalarTok.End = startPos, endPos
		scalarTok.Style = PlainScalarStyle
		scalarTok.Value = valueBuf.String()
		tok = scalarTok
	}
	if leadingBlanks {
		s.simpleKeyAllowed = true
	}
	err = nil
	return
}
