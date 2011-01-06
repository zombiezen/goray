//
//	goray/core/shader/eval_test.go
//	goray
//
//	Created by Ross Light on 2011-01-06.
//

package shader

import "testing"

type testNode struct {
	val float64
	deps []Node
}

func (n *testNode) Eval(params map[string]interface{}, inputs []Result) Result {
	sum := n.val
	for _, input := range inputs {
		sum += input.Scalar()
	}
	return Result{sum}
}

func (n *testNode) EvalDerivative(params map[string]interface{}, inputs []Result) Result {
	return Result{}
}

func (n *testNode) ViewDependent() bool { return false }
func (n *testNode) Dependencies() []Node { return n.deps }

func TestBasicEval(t *testing.T) {
	n := &testNode{
		0.0,
		[]Node{
			&testNode{2.0, nil},
			&testNode{-3.0, nil},
		},
	}
	r := Eval(nil, []Node{n})
	if len(r) == 1 {
		if r[0].Scalar() != -1 {
			t.Errorf("Got %.2f (expected -1)", r[0].Scalar())
		}
	} else {
		t.Errorf("Got %d results (expected 1)", len(r))
	}
}
