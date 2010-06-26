//
//	yaml/scanner/scanner.go
//	goray
//
//	Created by Ross Light on 2010-06-24.
//

package scanner

import (
	"fmt"
	"io"
	"os"
	"yaml/token"
)

type Scanner struct {
	reader           *reader
	started          bool
	flowLevel        uint
	simpleKeyAllowed bool
}

func New(r io.Reader) (s *Scanner) {
	s = new(Scanner)
	s.reader = newReader(r)
	return s
}

func (s *Scanner) Scan() (result Token, err os.Error) {
	return s.Next()
}

func (s *Scanner) Next() (result Token, err os.Error) {
	if !s.started {
		return s.streamStart()
	}

	if err = s.scanToNextToken(); err != nil {
		return
	}
	
	err = os.NewError(fmt.Sprintf("Unrecognized token: %c", s.reader.Bytes()[0]))
	return
}

func (s *Scanner) streamStart() (result Token, err os.Error) {
	s.started = true
	result = BasicToken{
		Kind: token.STREAM_START,
		Start: s.reader.Pos,
		End: s.reader.Pos,
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
