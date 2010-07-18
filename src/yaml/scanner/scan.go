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

// scanToNextToken discards any non-token characters from the reader.
func (s *Scanner) scanToNextToken() (err os.Error) {
	for {
		// Allow the BOM mark to start a line
		if err = s.reader.Cache(3); err != nil {
			return
		}
		if s.GetPosition().Column == 1 && s.reader.Check(0, "\xEF\xBB\xBF") {
			s.reader.Next(3)
		}

		// Eat whitespaces
		//
		// Tabs are allowed:
		// - in the flow context;
		// - in the block context, but not at the beginning of the line or
		//   after '-', '?', or ':' (complex value).
		if err = s.reader.Cache(1); err != nil {
			return
		}

		for s.reader.Check(0, " ") || (s.reader.Check(0, "\t") && (s.flowLevel > 0 || !s.simpleKeyAllowed)) {
			s.reader.Next(1)
			if err = s.reader.Cache(1); err != nil {
				return
			}
		}

		// Eat comment until end of line
		if s.reader.Check(0, "#") {
			for !s.reader.CheckBreak(0) && s.reader.Len() > 0 {
				s.reader.Next(1)
				if err = s.reader.Cache(1); err != nil {
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
	startPos := s.GetPosition()
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
		t.Start, t.End = startPos, s.GetPosition()
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
		t.Start, t.End = startPos, s.GetPosition()
		t.Handle, t.Prefix = handle, prefix
		tok = t
	default:
		err = os.NewError("Unrecognized directive: " + name)
		return
	}

	// Eat the rest of the line, including comments
	s.reader.SkipSpaces()
	if err = s.reader.CacheFull(1); err != nil {
		return
	}
	if s.reader.Check(0, "#") {
		for s.reader.Len() != 0 && !s.reader.CheckBreak(0) {
			s.reader.Next(1)
			if err = s.reader.CacheFull(1); err != nil {
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
	if err = s.reader.CacheFull(1); err != nil {
		return
	}

	for isLetter(s.reader.Bytes()[0]) {
		if _, err = io.Copyn(nameBuf, s.reader, 1); err != nil {
			return
		}
		if err = s.reader.CacheFull(1); err != nil {
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
		if err = s.reader.CacheFull(1); err != nil {
			return
		}
		for s.reader.CheckDigit(0) && numLen < 9 {
			b, _ := s.reader.ReadByte()
			n = n*10 + int(b-'0')
			numLen++
			if err = s.reader.CacheFull(1); err != nil {
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
	s.reader.SkipSpaces()

	// Scan handle
	if handle, err = s.scanTagHandle(true); err != nil {
		return
	}

	// Handle whitespace
	if err = s.reader.CacheFull(1); err != nil {
		return
	}
	if !s.reader.CheckSpace(0) {
		err = os.NewError("Did not find expected whitespace")
		return
	}
	s.reader.SkipSpaces()

	// Scan prefix
	if prefix, err = s.scanTagURI(); err != nil {
		return
	}
	if prefix == "" {
		err = os.NewError("Did not find expected tag URI")
		return
	}

	// Expect whitespace or line break
	if err = s.reader.Cache(1); err != nil {
		return
	}
	if !s.reader.CheckBlank(0) {
		err = os.NewError("Did not find expected whitespace or line break")
		return
	}

	return
}

func (s *Scanner) scanAnchor(kind token.Token) (tok ValueToken, err os.Error) {
	startPos := s.GetPosition()
	// Eat indicator
	s.reader.Next(1)
	// Consume value
	valueBuf := new(bytes.Buffer)
	if err = s.reader.CacheFull(1); err != nil {
		return
	}
	for s.reader.CheckLetter(0) {
		b, _ := s.reader.ReadByte()
		valueBuf.WriteByte(b)
		if err = s.reader.CacheFull(1); err != nil {
			return
		}
	}
	if valueBuf.Len() == 0 || !(s.reader.Len() == 0 || s.reader.CheckBreak(0) || s.reader.CheckSpace(0) || s.reader.CheckAny(0, "?:,]}%@`")) {
		err = os.NewError("Did not find expected alphabetic or numeric character")
		return
	}
	// Create token
	tok.Kind = kind
	tok.Start, tok.End = startPos, s.GetPosition()
	tok.Value = valueBuf.String()
	return
}

func (s *Scanner) scanTag() (tok TagToken, err os.Error) {
	var handle, suffix string

	startPos := s.GetPosition()
	if err = s.reader.Cache(2); err != nil {
		return
	}
	if s.reader.Check(1, "<") {
		// Verbatim tag (i.e. "!<foo:bar>")
		// Eat "!<"
		s.reader.Next(2)
		// Consume tag value
		if suffix, err = s.scanTagURI(); err != nil {
			return
		}
		if suffix == "" {
			err = os.NewError("Did not find expected tag URI")
			return
		}
		// Check for ">" and eat it
		if !s.reader.Check(0, ">") {
			err = os.NewError("Did not find expected '>'")
			return
		}
		s.reader.Next(1)
	} else {
		// Shorthand tag (i.e. "!suffix" or "!handle!suffix"
		// Try to scan a handle
		if handle, err = s.scanTagHandle(false); err != nil {
			return
		}
		// Check if it is a handle
		if len(handle) >= 2 && handle[0] == '!' && handle[len(handle)-1] == '!' {
			// Scan the suffix now
			if suffix, err = s.scanTagURI(); err != nil {
				return
			}
			if suffix == "" {
				err = os.NewError("Did not find expected tag URI")
				return
			}
		} else {
			// It wasn't a handle.  Scan the rest of the tag.
			if suffix, err = s.scanTagURI(); err != nil {
				return
			}
			handle, suffix = "!", handle[1:]+suffix
			// Special case: the "!" tag.
			if suffix == "" {
				handle, suffix = "", "!"
			}
		}
	}

	// Check the current character which ends the tag.
	if err = s.reader.Cache(1); err != nil {
		return
	}
	if !s.reader.CheckBlank(0) {
		err = os.NewError("Did not find expected whitespace or line break")
		return
	}

	// Create token
	tok.Kind = token.TAG
	tok.Start, tok.End = startPos, s.GetPosition()
	tok.Handle, tok.Suffix = handle, suffix
	err = nil
	return
}

func (s *Scanner) scanTagHandle(directive bool) (handle string, err os.Error) {
	handleBuf := new(bytes.Buffer)

	// Check the initial "!" character
	if err = s.reader.CacheFull(1); err != nil {
		return
	}
	if !s.reader.Check(0, "!") {
		err = os.NewError("Did not find expected '!'")
		return
	}

	// Copy the "!" character
	io.Copyn(handleBuf, s.reader, 1)

	// Copy all subsequent alphabetical and numerical characters
	if err = s.reader.Cache(1); err != nil {
		return
	}
	for s.reader.CheckLetter(0) {
		io.Copyn(handleBuf, s.reader, 1)
		if err = s.reader.Cache(1); err != nil {
			return
		}
	}

	// Check if the trailing character is "!" and copy it
	if s.reader.Check(0, "!") {
		io.Copyn(handleBuf, s.reader, 1)
	} else {
		// It's either the "!" tag or not really a tag handle.  If it's a %TAG
		// directive, it's an error.  If it's a tag token, it must be a part of
		// a URI.
		if directive && handleBuf.String() == "!" {
			err = os.NewError("Did not find expected '!'")
			return
		}
	}

	handle = handleBuf.String()
	return
}

func (s *Scanner) scanTagURI() (uri string, err os.Error) {
	uriBuf := new(bytes.Buffer)

	if err = s.reader.Cache(1); err != nil {
		return
	}

	for s.reader.CheckWord(0) || s.reader.CheckAny(0, ";/?:@&=+$,.!~*'()[]%") {
		if s.reader.Check(0, "%") {
			if err = s.scanURIEscape(uriBuf); err != nil {
				return
			}
		} else {
			io.Copyn(uriBuf, s.reader, 1)
		}

		if err = s.reader.Cache(1); err != nil {
			return
		}
	}

	uri = uriBuf.String()
	return
}

func (s *Scanner) scanURIEscape(buf *bytes.Buffer) (err os.Error) {
	for escapeWidth, left := 0, 1; left > 0; left-- {
		// Check for a URI-escaped octet
		if err = s.reader.CacheFull(3); err != nil {
			return
		}
		if !(s.reader.Check(0, "%") && s.reader.CheckHexDigit(1) && s.reader.CheckHexDigit(2)) {
			err = os.NewError("Did not find URI-escaped octet")
			return
		}
		// Get the octet
		digits := s.reader.Next(3)[1:3]
		digit1, _ := asHex(digits[0])
		digit2, _ := asHex(digits[1])
		octet := byte(digit1<<4) | byte(digit2)
		// If it is the leading octet, determine the length of the UTF-8 sequence
		if escapeWidth == 0 {
			switch {
			case octet&0x80 == 0x00:
				escapeWidth = 1
			case octet&0xE0 == 0xC0:
				escapeWidth = 2
			case octet&0xF0 == 0xE0:
				escapeWidth = 3
			case octet&0xF8 == 0xF0:
				escapeWidth = 4
			default:
				err = os.NewError("Found an incorrect leading UTF-8 octet")
				return
			}
			left = escapeWidth // `left' will be decremented on the loop's end
		} else {
			// For multi-byte UTF-8 sequences, all octets after the first one
			// must have the two most significant bits set to 10.
			if octet&0xC0 != 0x80 {
				err = os.NewError("Found an incorrect UTF-8 octet")
				return
			}
		}
		// Copy octet
		buf.WriteByte(octet)
	}
	return
}

func (s *Scanner) scanBlockScalar(style int) (tok ScalarToken, err os.Error) {
	const (
		chompStrip = -1
		chompClip  = 0
		chompKeep  = +1
	)

	var breaks []byte

	chomping := chompClip
	indent := 0
	increment := 0
	leadingBlank := false
	trailingBlank := false

	valueBuf := new(bytes.Buffer)
	leadingBreak := new(bytes.Buffer)
	trailingBreaks := new(bytes.Buffer)

	// Eat the indicator
	startPos := s.GetPosition()
	s.reader.Next(1)

	// Scan for additional block scalar indicators (order doesn't matter)
	if err = s.reader.CacheFull(1); err != nil {
		return
	}
	scanChomp := func() {
		switch {
		case s.reader.Check(0, "+"):
			chomping = chompKeep
		case s.reader.Check(0, "-"):
			chomping = chompStrip
		default:
			return
		}
		s.reader.Next(1)
	}
	scanIndent := func() os.Error {
		if s.reader.CheckDigit(0) {
			b, _ := s.reader.ReadByte()
			increment, _ = asDecimal(b)
			if increment == 0 {
				return os.NewError("Found an indentation equal to 0")
			}
		}
		return nil
	}
	if s.reader.CheckAny(0, "+-") {
		scanChomp()
		if err = s.reader.CacheFull(1); err != nil {
			return
		}
		if err = scanIndent(); err != nil {
			return
		}
	} else if s.reader.CheckDigit(0) {
		if err = scanIndent(); err != nil {
			return
		}
		if err = s.reader.CacheFull(1); err != nil {
			return
		}
		scanChomp()
	}

	// Eat whitespaces and comments to the end of the line
	s.reader.SkipSpaces()

	if err = s.reader.CacheFull(1); err != nil {
		return
	}
	if s.reader.Check(0, "#") {
		for !s.reader.CheckBreak(0) && s.reader.Len() > 0 {
			s.reader.Next(1)
			if err = s.reader.Cache(1); err != nil {
				return
			}
		}
	}

	// Ensure that we're at EOL
	if !s.reader.CheckBlank(0) {
		err = os.NewError("Did not find expected comment or line break")
		return
	}
	// Eat line break
	s.reader.ReadBreak()
	endPos := s.GetPosition()
	// Set the indentation level if it was specified
	if increment != 0 {
		if s.indent > 0 {
			indent = s.indent + increment
		} else {
			indent = increment
		}
	}

	// Scan the leading line breaks and determine indentation level, if needed
	if breaks, endPos, err = s.scanBlockScalarBreaks(&indent, startPos); err != nil {
		return
	}
	trailingBreaks.Write(breaks)

	// Scan content
	if err = s.reader.Cache(1); err != nil {
		return
	}
	for s.GetPosition().Column == indent && s.reader.Len() > 0 {
		// We are at the beginning of a non-empty line
		trailingBlank = s.reader.CheckSpace(0)
		// Do we need to fold the leading line break?
		if style == FoldedScalarStyle && leadingBreak.Len() > 0 && !leadingBlank && !trailingBlank {
			if trailingBreaks.Len() == 0 {
				valueBuf.WriteByte(' ')
			}
			leadingBreak.Reset()
		} else {
			io.Copy(valueBuf, leadingBreak)
		}
		// Append the remaining line breaks
		io.Copy(valueBuf, trailingBreaks)
		// Is it a leading whitespace?
		leadingBlank = s.reader.CheckSpace(0)
		// Consume the current line
		for !s.reader.CheckBreak(0) {
			io.Copyn(valueBuf, s.reader, 1)
			if err = s.reader.Cache(1); err != nil {
				return
			}
		}
		// Consume the line break
		if breaks, err = s.reader.ReadBreak(); err != nil {
			return
		}
		leadingBreak.Write(breaks)
		// Eat the following indentation spaces and line breaks
		if breaks, endPos, err = s.scanBlockScalarBreaks(&indent, startPos); err != nil {
			return
		}
		trailingBreaks.Write(breaks)
	}

	// Chomp the tail
	if chomping != chompStrip {
		io.Copy(valueBuf, leadingBreak)
	}
	if chomping == chompKeep {
		io.Copy(valueBuf, trailingBreaks)
	}

	// Create token
	tok.Kind = token.SCALAR
	tok.Start, tok.End = startPos, endPos
	tok.Value = valueBuf.String()
	tok.Style = style
	err = nil
	return
}

func (s *Scanner) scanBlockScalarBreaks(indent *int, startPos token.Position) (breaks []byte, endPos token.Position, err os.Error) {
	maxIndent := 0
	endPos = s.GetPosition()
	breaksBuffer := new(bytes.Buffer)

	for {
		// Eat the indentation spaces
		if err = s.reader.Cache(1); err != nil {
			return
		}
		for (*indent == 0 || s.GetPosition().Column < *indent) && s.reader.Check(0, " ") {
			s.reader.Next(1)
			if err = s.reader.Cache(1); err != nil {
				return
			}
		}
		if s.GetPosition().Column > maxIndent {
			maxIndent = s.GetPosition().Column
		}
		// Check for a tab character messing the indentation
		if (*indent == 0 || s.GetPosition().Column < *indent) && s.reader.Check(0, "\t") {
			err = os.NewError("Found a tab character where an indentation space is expected")
			return
		}
		// Have we found a non-empty line?
		if !s.reader.CheckBreak(0) || s.reader.Len() == 0 {
			break
		}
		// Consume the line break
		var brBytes []byte
		if brBytes, err = s.reader.ReadBreak(); err != nil {
			return
		}
		breaksBuffer.Write(brBytes)
		endPos = s.GetPosition()
	}

	// Determine the indentation level, if needed
	if *indent == 0 {
		*indent = maxIndent
		if *indent < s.indent+1 {
			*indent = s.indent + 1
		}
		if *indent < 1 {
			*indent = 1
		}
	}

	breaks = breaksBuffer.Bytes()
	err = nil
	return
}

func (s *Scanner) scanFlowScalar(style int) (tok ScalarToken, err os.Error) {
	startPos := s.GetPosition()
	valueBuf := new(bytes.Buffer)
	// Eat the left quote
	s.reader.Next(1)
	// Consume content
	for {
		leadingBlanks := false

		if err = s.reader.Cache(4); err != nil {
			return
		}
		err = nil

		// Better not be a document indicator.
		if s.GetPosition().Column == 1 && (s.reader.Check(0, "---") || s.reader.Check(0, "...")) && s.reader.CheckBlank(3) {
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
			if err = s.reader.Cache(2); err != nil {
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
		if err = s.reader.CacheFull(1); err != nil {
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
			if err = s.reader.CacheFull(1); err != nil {
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
	tok.Kind = token.SCALAR
	tok.Start, tok.End = startPos, s.GetPosition()
	tok.Style = style
	tok.Value = valueBuf.String()
	err = nil
	return
}

func (s *Scanner) scanEscapeSeq() (rune int, err os.Error) {
	if err = s.reader.CacheFull(2); err != nil {
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
		if err = s.reader.CacheFull(codeLength); err != nil {
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

func (s *Scanner) scanPlainScalar() (tok ScalarToken, err os.Error) {
	startPos, endPos := s.GetPosition(), s.GetPosition()
	valueBuf := new(bytes.Buffer)
	indent := s.indent + 1

	// Set up buffers
	leadingBlanks := false
	whitespaces := new(bytes.Buffer)
	leadingBreak := new(bytes.Buffer)
	trailingBreaks := new(bytes.Buffer)

	for {
		if err = s.reader.Cache(4); err != nil {
			return
		}
		err = nil
		// Stop on a document indicator
		if s.GetPosition().Column == 1 && (s.reader.Check(0, "---") || s.reader.Check(0, "...")) && s.reader.CheckBlank(3) {
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
			endPos = s.GetPosition()
			if err = s.reader.Cache(2); err != nil {
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
				if leadingBlanks && s.GetPosition().Column < indent && s.reader.Check(0, "\t") {
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
		if s.flowLevel == 0 && s.GetPosition().Column < indent {
			break
		}
	}

	// Create token
	tok.Kind = token.SCALAR
	tok.Start, tok.End = startPos, endPos
	tok.Style = PlainScalarStyle
	tok.Value = valueBuf.String()
	if leadingBlanks {
		s.simpleKeyAllowed = true
	}
	err = nil
	return
}
