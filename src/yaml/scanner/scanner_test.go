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
		"Plain literal",
		`Hello, World!`,
		[]Token{
			BasicToken{token.STREAM_START, token.Position{0, 1, 1}, token.Position{0, 1, 1}},
			ValueToken{BasicToken{token.SCALAR, token.Position{0, 1, 1}, token.Position{13, 1, 14}}, "Hello, World!"},
			BasicToken{token.STREAM_END, token.Position{13, 2, 1}, token.Position{13, 2, 1}},
		},
	},
	scanTest{
		"Comment",
		`# Comment`,
		[]Token{
			BasicToken{token.STREAM_START, token.Position{0, 1, 1}, token.Position{0, 1, 1}},
			BasicToken{token.STREAM_END, token.Position{9, 2, 1}, token.Position{9, 2, 1}},
		},
	},
	scanTest{
		"Comment with content",
		"# Comment\nHello",
		[]Token{
			BasicToken{token.STREAM_START, token.Position{0, 1, 1}, token.Position{0, 1, 1}},
			ValueToken{BasicToken{token.SCALAR, token.Position{10, 2, 1}, token.Position{15, 2, 6}}, "Hello"},
			BasicToken{token.STREAM_END, token.Position{15, 3, 1}, token.Position{15, 3, 1}},
		},
	},
	scanTest{
		"Flow sequence",
		`[Foo, Bar, Baz]`,
		[]Token{
			BasicToken{token.STREAM_START, token.Position{0, 1, 1}, token.Position{0, 1, 1}},
			BasicToken{token.FLOW_SEQUENCE_START, token.Position{0, 1, 1}, token.Position{1, 1, 2}},

			ValueToken{BasicToken{token.SCALAR, token.Position{1, 1, 2}, token.Position{4, 1, 5}}, "Foo"},
			BasicToken{token.FLOW_ENTRY, token.Position{4, 1, 5}, token.Position{5, 1, 6}},

			ValueToken{BasicToken{token.SCALAR, token.Position{6, 1, 7}, token.Position{9, 1, 10}}, "Bar"},
			BasicToken{token.FLOW_ENTRY, token.Position{9, 1, 10}, token.Position{10, 1, 11}},

			ValueToken{BasicToken{token.SCALAR, token.Position{11, 1, 12}, token.Position{14, 1, 15}}, "Baz"},

			BasicToken{token.FLOW_SEQUENCE_END, token.Position{14, 1, 15}, token.Position{15, 1, 16}},
			BasicToken{token.STREAM_END, token.Position{15, 2, 1}, token.Position{15, 2, 1}},
		},
	},
	scanTest{
		"Flow mapping",
		`{Spam: Eggs, Knights: Ni}`,
		[]Token{
			BasicToken{token.STREAM_START, token.Position{0, 1, 1}, token.Position{0, 1, 1}},
			BasicToken{token.FLOW_MAPPING_START, token.Position{0, 1, 1}, token.Position{1, 1, 2}},

			BasicToken{token.KEY, token.Position{1, 1, 2}, token.Position{1, 1, 2}},
			ValueToken{BasicToken{token.SCALAR, token.Position{1, 1, 2}, token.Position{5, 1, 6}}, "Spam"},
			BasicToken{token.VALUE, token.Position{5, 1, 6}, token.Position{6, 1, 7}},
			ValueToken{BasicToken{token.SCALAR, token.Position{7, 1, 8}, token.Position{11, 1, 12}}, "Eggs"},
			BasicToken{token.FLOW_ENTRY, token.Position{11, 1, 12}, token.Position{12, 1, 13}},

			BasicToken{token.KEY, token.Position{13, 1, 14}, token.Position{13, 1, 14}},
			ValueToken{BasicToken{token.SCALAR, token.Position{13, 1, 14}, token.Position{20, 1, 21}}, "Knights"},
			BasicToken{token.VALUE, token.Position{20, 1, 21}, token.Position{21, 1, 22}},
			ValueToken{BasicToken{token.SCALAR, token.Position{22, 1, 23}, token.Position{24, 1, 25}}, "Ni"},

			BasicToken{token.FLOW_MAPPING_END, token.Position{24, 1, 25}, token.Position{25, 1, 26}},
			BasicToken{token.STREAM_END, token.Position{25, 2, 1}, token.Position{25, 2, 1}},
		},
	},
	scanTest{
		"Block sequence",
		"- Foo\n- Bar\n- Baz",
		[]Token{
			BasicToken{token.STREAM_START, token.Position{0, 1, 1}, token.Position{0, 1, 1}},
			BasicToken{token.BLOCK_SEQUENCE_START, token.Position{0, 1, 1}, token.Position{0, 1, 1}},

			BasicToken{token.BLOCK_ENTRY, token.Position{0, 1, 1}, token.Position{1, 1, 2}},
			ValueToken{BasicToken{token.SCALAR, token.Position{2, 1, 3}, token.Position{5, 1, 6}}, "Foo"},

			BasicToken{token.BLOCK_ENTRY, token.Position{6, 2, 1}, token.Position{7, 2, 2}},
			ValueToken{BasicToken{token.SCALAR, token.Position{8, 2, 3}, token.Position{11, 2, 6}}, "Bar"},

			BasicToken{token.BLOCK_ENTRY, token.Position{12, 3, 1}, token.Position{13, 3, 2}},
			ValueToken{BasicToken{token.SCALAR, token.Position{14, 3, 3}, token.Position{17, 3, 6}}, "Baz"},

			BasicToken{token.BLOCK_END, token.Position{17, 4, 1}, token.Position{17, 4, 1}},
			BasicToken{token.STREAM_END, token.Position{17, 4, 1}, token.Position{17, 4, 1}},
		},
	},
	scanTest{
		"Block mapping",
		"a: b\nc: d",
		[]Token{
			BasicToken{token.STREAM_START, token.Position{0, 1, 1}, token.Position{0, 1, 1}},
			BasicToken{token.BLOCK_MAPPING_START, token.Position{0, 1, 1}, token.Position{0, 1, 1}},

			BasicToken{token.KEY, token.Position{0, 1, 1}, token.Position{0, 1, 1}},
			ValueToken{BasicToken{token.SCALAR, token.Position{0, 1, 1}, token.Position{1, 1, 2}}, "a"},
			BasicToken{token.VALUE, token.Position{1, 1, 2}, token.Position{2, 1, 3}},
			ValueToken{BasicToken{token.SCALAR, token.Position{3, 1, 4}, token.Position{4, 1, 5}}, "b"},

			BasicToken{token.KEY, token.Position{5, 2, 1}, token.Position{5, 2, 1}},
			ValueToken{BasicToken{token.SCALAR, token.Position{5, 2, 1}, token.Position{6, 2, 2}}, "c"},
			BasicToken{token.VALUE, token.Position{6, 2, 2}, token.Position{7, 2, 3}},
			ValueToken{BasicToken{token.SCALAR, token.Position{8, 2, 4}, token.Position{9, 2, 5}}, "d"},

			BasicToken{token.BLOCK_END, token.Position{9, 3, 1}, token.Position{9, 3, 1}},
			BasicToken{token.STREAM_END, token.Position{9, 3, 1}, token.Position{9, 3, 1}},
		},
	},
	scanTest{
		"Basic document",
		`%YAML 1.2
---
"Hello, World!"
...`,
		[]Token{
			BasicToken{token.STREAM_START, token.Position{0, 1, 1}, token.Position{0, 1, 1}},
			BasicToken{token.VERSION_DIRECTIVE, token.Position{0, 1, 1}, token.Position{9, 1, 10}},
			BasicToken{token.DOCUMENT_START, token.Position{10, 2, 1}, token.Position{13, 2, 4}},
			ValueToken{BasicToken{token.SCALAR, token.Position{14, 3, 1}, token.Position{29, 3, 16}}, "Hello, World!"},
			BasicToken{token.DOCUMENT_END, token.Position{30, 4, 1}, token.Position{33, 4, 4}},
			BasicToken{token.STREAM_END, token.Position{33, 5, 1}, token.Position{33, 5, 1}},
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
					t.Errorf("%v: token %d started at %v (expected %v)", test, i, tok.GetStart(), expected.GetStart())
				}
				if !posEq(tok.GetEnd(), expected.GetEnd()) {
					t.Errorf("%v: token %d ended at %v (expected %v)", test, i, tok.GetEnd(), expected.GetEnd())
				}
				if _, ok := expected.(ValueToken); ok && tok.String() != expected.String() {
					t.Errorf("%v: token %d had wrong value %#v (expected %#v)", test, i, tok.String(), expected.String())
				}
			}
		} else {
			t.Errorf("%v: Yielded %v", test, results)
		}
	}
}
