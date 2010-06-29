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

type reader struct {
	*bytes.Buffer
	Reader io.Reader
	Pos    token.Position
	lastCR bool
}

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

func (r *reader) Cache(n int) (err os.Error) {
	var nRead int64
	// Do we already have enough buffered?
	if r.Len() >= n {
		return nil
	}
	// Read more bytes
	fillSize := int64(n - r.Len())
	nRead, err = io.Copyn(r, r.Reader, fillSize)
	if err == nil && nRead < fillSize {
		err = io.ErrUnexpectedEOF
	}
	return
}

func (r *reader) Read(p []byte) (n int, err os.Error) {
	if r.Buffer.Len() > 0 {
		n, _ = r.Buffer.Read(p)
	} else {
		n, err = r.Reader.Read(p)
	}
	r.updatePos(p[0:n])
	return
}

func (r *reader) ReadByte() (c byte, err os.Error) {
	if err = r.Cache(1); err != nil {
		if err == io.ErrUnexpectedEOF {
			err = os.EOF
		}
		return
	}
	c = r.Next(1)[0] // Next will update the position for us
	return
}

func (r *reader) Next(n int) (bytes []byte) {
	bytes = r.Buffer.Next(n)
	r.updatePos(bytes)
	return
}

func (r *reader) Check(st string) bool {
	if r.Len() < len(st) {
		return false
	}
	return st == string(r.Bytes()[0:len(st)])
}

func (r *reader) CheckBreak() bool {
	if r.Len() == 0 {
		return false
	}
	return isBreak(r.Bytes()[0])
}

func (r *reader) CheckDigit() bool {
	if r.Len() == 0 {
		return false
	}
	return isDigit(r.Bytes()[0])
}

func (r *reader) CheckLetter() bool {
	if r.Len() == 0 {
		return false
	}
	return isLetter(r.Bytes()[0])
}

func (r *reader) CheckSpace() bool {
	if r.Len() == 0 {
		return false
	}
	return isWhitespace(r.Bytes()[0])
}

func (r *reader) SkipBreak() {
	if err := r.Cache(2); err != nil && (err != io.ErrUnexpectedEOF || r.Len() == 0) {
		return
	}

	switch r.Bytes()[0] {
	case '\n':
		r.Next(1)
	case '\r':
		if r.Len() > 1 && r.Bytes()[1] == '\n' {
			r.Next(2)
		} else {
			r.Next(1)
		}
	}
}

func (r *reader) SkipSpaces() {
	if err := r.Cache(1); err != nil {
		return
	}
	for r.CheckSpace() {
		r.Next(1)
		if err := r.Cache(1); err != nil {
			return
		}
	}
}
