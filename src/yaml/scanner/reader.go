//
//	yaml/scanner/reader.go
//	goray
//
//	Created by Ross Light on 2010-06-25.
//

package scanner

import (
	"bytes"
	"io"
	"os"
	"yaml/token"
)

// reader is an arbitrarily buffered reader.
type reader struct {
	Buffer *bytes.Buffer
	Reader io.Reader
	Pos    token.Position
	lastCR bool
}

// newReader creates a new reader.
func newReader(r io.Reader) *reader {
	initPos := token.Position{
		Index:  0,
		Column: 1,
		Line:   1,
	}
	return &reader{
		Buffer: new(bytes.Buffer),
		Reader: r,
		Pos:    initPos,
		lastCR: false,
	}
}

// updatePos updates the Pos field of the reader.  The data passed into the
// method is what was most recently read.
func (r *reader) updatePos(data []byte) {
	r.Pos.Index += len(data)
	// If the last character from the previous update was a CR and the first
	// character isn't a LF, then increment line number. (LFs always increment
	// line number.)
	if len(data) > 0 && r.lastCR && data[0] != '\n' {
		r.Pos.Column = 1
		r.Pos.Line++
	}
	// Update column and line information
	for i, b := range data {
		switch b {
		case '\n':
			r.Pos.Column = 1
			r.Pos.Line++
		case '\r':
			switch {
			case i+1 >= len(data):
				// The last byte in the data is a CR. Crap.
				r.lastCR = true
			case data[i+1] == '\n':
				// This is a CRLF.  Do nothing (the next iteration will catch it).
			default:
				// This is a naked CR (Mac-style).
				r.Pos.Column = 1
				r.Pos.Line++
			}
		default:
			r.Pos.Column++
		}
	}
}

// Cache reads characters into the buffer until the buffer contains n
// characters, or an error occurs, whichever comes first.  An EOF will not
// return an error; the cache will simply contain less characters.
func (r *reader) Cache(n int) (err os.Error) {
	// Do we already have enough buffered?
	if r.Len() >= n {
		return nil
	}
	// Read more bytes
	fillSize := int64(n - r.Len())
	_, err = io.Copyn(r.Buffer, r.Reader, fillSize)
	return
}

// CacheFull is the same as Cache, except it will return an io.ErrUnexpectedEOF
// if an EOF is discovered.
func (r *reader) CacheFull(n int) (err os.Error) {
	err = r.Cache(n)
	if err == nil && r.Len() < n {
		err = io.ErrUnexpectedEOF
	}
	return
}

// Read reads up to len(p) bytes into p.  If the buffer is not empty, then any
// bytes from the buffer will be returned first.
func (r *reader) Read(p []byte) (n int, err os.Error) {
	if r.Buffer.Len() > 0 {
		n, _ = r.Buffer.Read(p)
	} else {
		n, err = r.Reader.Read(p)
	}
	r.updatePos(p[0:n])
	return
}

// ReadByte reads a single byte.  If the buffer is not empty, then the byte
// returned will be the buffer's first byte.
func (r *reader) ReadByte() (c byte, err os.Error) {
	if err = r.CacheFull(1); err != nil {
		if err == io.ErrUnexpectedEOF {
			err = os.EOF
		}
		return
	}
	c = r.Next(1)[0] // Next will update the position for us
	return
}

// ReadBreak reads a line break at the current position.  This will correctly
// handle CRLF sequences.
func (r *reader) ReadBreak() (bytes []byte, err os.Error) {
	if err := r.Cache(2); err != nil || r.Len() == 0 {
		return
	}

	switch r.Bytes()[0] {
	case '\n':
		bytes = r.Next(1)
	case '\r':
		if r.Len() > 1 && r.Bytes()[1] == '\n' {
			bytes = r.Next(2)
		} else {
			bytes = r.Next(1)
		}
	}
	return
}

// Next returns a slice containing the next n bytes in the buffer.  If n is
// greater than the size of the buffer, the entire buffer's contents will be
// returned.
func (r *reader) Next(n int) (bytes []byte) {
	bytes = r.Buffer.Next(n)
	r.updatePos(bytes)
	return
}

// Bytes returns a slice of the contents of the buffer.
func (r *reader) Bytes() []byte { return r.Buffer.Bytes() }
// Len returns the number of bytes in the buffer.
func (r *reader) Len() int { return r.Buffer.Len() }

// Check returns whether the buffer contains the given substring at i.  If the
// buffer is not big enough to possibly contain the substring, Check returns
// false.
func (r *reader) Check(i int, st string) bool {
	if r.Len() < i+len(st) {
		return false
	}
	return st == string(r.Bytes()[i:i+len(st)])
}

// CheckAny returns whether the character at i in the buffer is any one of the
// characters in chars.
func (r *reader) CheckAny(i int, chars string) bool {
	if r.Len() <= i {
		return false
	}
	currByte := r.Bytes()[i]
	for _, c := range chars {
		if currByte == byte(c) {
			return true
		}
	}
	return false
}

// CheckBreak returns whether the character at i in the buffer is a line break.
func (r *reader) CheckBreak(i int) bool {
	if r.Len() <= i {
		return false
	}
	return isBreak(r.Bytes()[i])
}

// CheckDigit returns whether the character at i in the buffer is an ASCII digit.
func (r *reader) CheckDigit(i int) bool {
	if r.Len() <= i {
		return false
	}
	return isDigit(r.Bytes()[i])
}

// CheckHexDigit returns whether the character at i in the buffer is a hexadecimal digit.
func (r *reader) CheckHexDigit(i int) bool {
	if r.Len() <= i {
		return false
	}
	return isHexDigit(r.Bytes()[i])
}

// CheckLetter returns whether the character at i in the buffer is an ASCII letter.
func (r *reader) CheckLetter(i int) bool {
	if r.Len() <= i {
		return false
	}
	return isLetter(r.Bytes()[i])
}

// CheckWord returns whether the character at i in the buffer is an ASCII alphanumeric.
func (r *reader) CheckWord(i int) bool {
	if r.Len() <= i {
		return false
	}
	return isWordChar(r.Bytes()[i])
}

// CheckSpace returns whether the character at i in the buffer is a space or tab character.
func (r *reader) CheckSpace(i int) bool {
	if r.Len() <= i {
		return false
	}
	return isWhitespace(r.Bytes()[i])
}

// CheckBlank returns whether the buffer ends before the given index, or if it
// doesn't, whether that index contains whitespace or a linebreak.
func (r *reader) CheckBlank(i int) bool {
	if r.Len() <= i {
		return true
	}
	return r.CheckSpace(i) || r.CheckBreak(i)
}

// SkipSpaces discards characters until it finds a non-whitespace character or
// an error.
func (r *reader) SkipSpaces() {
	if err := r.CacheFull(1); err != nil {
		return
	}
	for r.CheckSpace(0) {
		r.Next(1)
		if err := r.CacheFull(1); err != nil {
			return
		}
	}
}
