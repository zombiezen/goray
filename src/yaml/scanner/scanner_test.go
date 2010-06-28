package scanner

import (
	"bytes"
	"container/vector"
	"fmt"
	"testing"
	"yaml/token"
)

type scanTest struct {
	Name     string
	Input    string
	Expected []Token
}

func (t scanTest) String() string {
	return fmt.Sprintf("%s test", t.Name)
}

var scannerTests = []scanTest{
	scanTest{
		"Empty",
		"",
		[]Token{
			BasicToken{token.STREAM_START, token.Position{0, 1, 1}, token.Position{0, 1, 1}},
			BasicToken{token.STREAM_END, token.Position{0, 1, 1}, token.Position{0, 1, 1}},
		},
	},
	scanTest{
		"Basic",
		`%YAML 1.2
---
"Hello, World!"`,
		[]Token{
			BasicToken{token.STREAM_START, token.Position{0, 1, 1}, token.Position{0, 1, 1}},
			BasicToken{token.VERSION_DIRECTIVE, token.Position{0, 1, 1}, token.Position{8, 1, 9}},
			BasicToken{token.DOCUMENT_START, token.Position{10, 2, 1}, token.Position{12, 2, 3}},
			BasicToken{token.SCALAR, token.Position{14, 3, 1}, token.Position{28, 3, 15}},
			BasicToken{token.STREAM_END, token.Position{28, 3, 15}, token.Position{28, 3, 15}},
		},
	},
}

func posEq(a, b token.Position) bool {
	return a.Index == b.Index && a.Line == b.Line && a.Column == b.Column
}

func TestScanner(t *testing.T) {
	for _, test := range scannerTests {
		scanner := New(bytes.NewBufferString(test.Input))
		results := make(vector.Vector, 0, len(test.Expected))
		for results.Len() == 0 || results.At(results.Len()-1).(Token).GetKind() != token.STREAM_END {
			tok, err := scanner.Scan()
			if err != nil {
				t.Errorf("%v error: %v", test, err)
				break
			}
			results.Push(tok)
		}

		if len(test.Expected) == results.Len() {
			for i, val := range results {
				tok := val.(Token)
				expected := test.Expected[i]
				if tok.GetKind() != expected.GetKind() {
					t.Errorf("%v: got wrong token %v at %d", test, tok.GetKind(), i)
				}
				if !posEq(tok.GetStart(), expected.GetStart()) {
					t.Errorf("%v: token %d started at %v", test, i, tok.GetStart())
				}
				if !posEq(tok.GetEnd(), expected.GetEnd()) {
					t.Errorf("%v: token %d ended at %v", test, i, tok.GetStart())
				}
			}
		} else {
			t.Errorf("%v: Yielded %v", test, results)
		}
	}
}
