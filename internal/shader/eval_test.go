/*
	Copyright (c) 2011 Ross Light.
	Copyright (c) 2005 Mathias Wein, Alejandro Conty, and Alfredo de Greef.

	This file is part of goray.

	goray is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	goray is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with goray.  If not, see <http://www.gnu.org/licenses/>.
*/

package shader

import "testing"

type testNode struct {
	val  float64
	deps []Node
}

func (n *testNode) Eval(inputs []Result, params Params) Result {
	sum := n.val
	for _, input := range inputs {
		sum += input.Scalar()
	}
	return Result{sum}
}

func (n *testNode) EvalDerivative(inputs []Result, params Params) Result {
	return Result{}
}

func (n *testNode) ViewDependent() bool  { return false }
func (n *testNode) Dependencies() []Node { return n.deps }

func TestBasicEval(t *testing.T) {
	n := &testNode{
		0.0,
		[]Node{
			&testNode{2.0, nil},
			&testNode{-3.0, nil},
		},
	}
	r := Eval([]Node{n}, nil)
	if len(r) == 1 {
		if r[0].Scalar() != -1 {
			t.Errorf("Got %.2f (expected -1)", r[0].Scalar())
		}
	} else {
		t.Errorf("Got %d results (expected 1)", len(r))
	}
}
