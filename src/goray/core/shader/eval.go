//
//	goray/core/shader/eval.go
//	goray
//
//	Created by Ross Light on 2011-01-06.
//

package shader

import (
	"container/list"
)

type evalNode struct {
	Node       Node
	Dependents []*evalNode
	DepCount   int
}

func (n *evalNode) Ready() bool {
	return n.DepCount == 0
}

func (n *evalNode) ClearDependency() {
	if n.DepCount > 0 {
		n.DepCount--
	}
}

type eqItem struct {
	Node      Node
	Dependent *evalNode
}

func buildTree(targets []Node) (nodes []*evalNode) {
	tree := make(map[Node]*evalNode)
	queue := list.New()
	for _, n := range targets {
		queue.PushBack(eqItem{n, nil})
	}

	for queue.Len() > 0 {
		item := queue.Remove(queue.Front()).(eqItem)

		en, ok := tree[item.Node]
		if !ok {
			deps := item.Node.Dependencies()
			en = &evalNode{
				Node:     item.Node,
				DepCount: len(deps),
			}
			tree[item.Node] = en

			for _, d := range deps {
				queue.PushBack(eqItem{d, en})
			}
		}

		if item.Dependent != nil {
			en.Dependents = append(en.Dependents, item.Dependent)
		}
	}

	nodes = make([]*evalNode, 0, len(tree))
	for _, en := range tree {
		nodes = append(nodes, en)
	}
	return
}

type evalFunc func(Node, map[string]interface{}, []Result) Result

func evalTask(params map[string]interface{}, inputs []Result, node Node, f evalFunc, ch chan<- Result) {
	defer close(ch)
	ch <- f(node, params, inputs)
}

func Eval(params map[string]interface{}, targets []Node) (final []Result) {
	results := make(map[Node]Result)
	nodes := buildTree(targets)

	for len(nodes) > 0 {
		nextNodes := make([]*evalNode, 0, len(nodes))
		channels := make(map[*evalNode]chan Result)

		for _, en := range nodes {
			if en.Ready() {
				// Build inputs
				deps := en.Node.Dependencies()
				inputs := make([]Result, len(deps))
				for i, d := range deps {
					inputs[i] = results[d]
				}
				// Start goroutine
				ch := make(chan Result, 1)
				channels[en] = ch
				go evalTask(params, inputs, en.Node, Node.Eval, ch)
			} else {
				nextNodes = append(nextNodes, en)
			}
		}

		// Collect results
		for en, ch := range channels {
			results[en.Node] = <-ch
			for _, d := range en.Dependents {
				d.ClearDependency()
			}
		}

		// Go to next generation
		nodes = nextNodes
	}

	final = make([]Result, len(targets))
	for i, t := range targets {
		final[i] = results[t]
	}
	return
}
