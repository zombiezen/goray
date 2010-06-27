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
}

func newReader(r io.Reader) *reader {
	initPos := token.Position{
		Index: 0,
		Column: 1,
		Line: 1,
	}
	return &reader{new(bytes.Buffer), r, initPos}
}

func (r *reader) Cache(n int) os.Error {
	// Do we already have enough buffered?
	if r.Len() >= n {
		return nil
	}
	// Read more bytes
	_, err := io.Copyn(r, r.Reader, int64(n-r.Len()))
	if err == os.EOF {
		err = nil
	}
	return err
}

func (r *reader) Read(p []byte) (n int, err os.Error) {
	if r.Buffer.Len() > 0 {
		n, _ = r.Buffer.Read(p)
	} else {
		n, err = r.Reader.Read(p)
	}
	r.Pos.Index += n
	r.Pos.Column += n
	return
}

func (r *reader) ReadByte() (c byte, err os.Error) {
	if err = r.Cache(1); err != nil {
		return
	}
	r.Pos.Index++
	r.Pos.Column++
	return r.Bytes()[0], nil
}

func (r *reader) Next(n int) (bytes []byte) {
	bytes = r.Buffer.Next(n)
	r.Pos.Index += len(bytes)
	r.Pos.Column += len(bytes)
	return
}

func (r *reader) Check(st string) bool {
	if r.Len() < len(st) {
		return false
	}
	return st == string(r.Bytes()[0:len(st)])
}

func (r *reader) CheckBreak() bool {
	return r.Check("\n") || r.Check("\r")
}

func (r *reader) SkipBreak() {
	if r.Len() == 0 {
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
	r.Pos.Column = 0
	r.Pos.Line++
}
