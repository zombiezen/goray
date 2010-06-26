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
	"yaml/token"
)

type reader struct {
	*bytes.Buffer
	Reader io.Reader
	Pos    token.Position
}

func (r *reader) Cache(n int) os.Error {
	// Do we already have enough buffered?
	if r.Len() >= n {
		return nil
	}
	// Read more bytes
	_, err := io.Copyn(r, r.Reader, n-r.Len())
	if err == os.EOF {
		err = nil
	}
	return err
}

func (r *reader) Read(p []byte) (n int, err os.Error) {
	if r.Buffer.Len() > 0 {
		n, _ = r.Buffer.Read(p)
		return
	} else {
		return r.reader.Read(p)
	}
}

func (r *reader) ReadByte() (c byte, err os.Error) {
	if err = r.Cache(1); err != nil {
		return
	}
	return r.Bytes()[0], nil
}

func (r *reader) Next(n int) (bytes []byte) {
	bytes = r.Buffer.Next(n)
	r.pos.Index += len(bytes)
	r.pos.Column += len(bytes)
	return
}

func (r *reader) Check(st string) bool {
	if r.Len() < st {
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
	r.pos.Column = 0
	r.pos.Line++
}
