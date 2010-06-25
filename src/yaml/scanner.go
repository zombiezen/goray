//
//	yaml/scanner.go
//	goray
//
//	Created by Ross Light on 2010-06-24.
//

package yaml

import (
	"bytes"
	"io"
	"os"
)

type scannerReader struct {
	reader io.Reader
	*bytes.Buffer
	pos Position
}

func (r *scannerReader) Cache(n int) os.Error {
	// Do we already have enough buffered?
	if r.Len() >= n {
		return nil
	}
	// Read more bytes
	_, err := io.Copyn(r, r.reader, n - r.Len())
	if err == os.EOF {
		err = nil
	}
	return err
}

func (r *scannerReader) Read(p []byte) (n int, err os.Error) {
	if r.Buffer.Len() > 0 {
		n, _ = r.Buffer.Read(p)
		return
	} else {
		return r.reader.Read(p)
	}
}

func (r *scannerReader) ReadByte() (c byte, err os.Error) {
	if err = r.Cache(1); err != nil {
		return
	}
	return r.Bytes()[0], nil
}

func (r *scannerReader) Next(n int) (bytes []byte) {
	bytes = r.Buffer.Next(n)
	r.pos.Index += len(bytes)
	r.pos.Column += len(bytes)
	return
}

func (r *scannerReader) Check(st string) bool {
	if r.Len() < st {
		return false
	}
	return st == string(r.Bytes()[0:len(st)])
}

func (r *scannerReader) CheckBreak() bool {
	return r.Check("\n") || r.Check("\r")
}

func (r *scannerReader) SkipBreak() {
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

type scanner struct {
	reader *scannerReader
	started bool
	flowLevel uint
	simpleKeyAllowed bool
}

type scanResult struct {
	Token Token
	Start, End Position
}

func newScanner(r io.Reader) (s *scanner) {
	s = new(scanner)
	s.reader = &scannerReader{reader: r, Buffer: new(bytes.Buffer)}
	s.reader.pos.Line = 1
	return s
}

func (s *scanner) Scan() (result scanResult, err os.Error) {
	
}

func (s *scanner) Next() (result scanResult, err os.Error) {
	if !s.started {
		return s.streamStart()
	}
	
	if err = s.scanToNextToken(); err != nil {
		return err
	}
}

func (s *scanner) streamStart() (result scanResult, err os.Error) {
	s.started = true
}

func (s *scanner) scanToNextToken() (err os.Error) {
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
