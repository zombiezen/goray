//
//  goray/kdtree.go
//  goray
//
//  Created by Ross Light on 2010-06-02.
//

package kdtree

import (
	"math"
	"./fmath"
)

import (
	"./goray/bound"
	"./goray/color"
	"./goray/partition"
	"./goray/primitive"
	"./goray/render"
	"./goray/ray"
	"./goray/vector"
)

const (
	lowerB = 0
	upperB = 2
	bothB  = 1
)

type boundEdge struct {
	position     float
	primitiveNum int
	end          int
}

func (edge boundEdge) Less(i interface{}) bool {
	switch other := i.(type) {
	case boundEdge:
		if edge.position == other.position {
			return edge.end > other.end
		} else {
			return edge.position < other.position
		}
	}
	return false
}

type splitCost struct {
	bestAxis              int
	bestOffset            int
	bestCost, oldCost     float
	t                     float
	nBelow, nAbove, nEdge int
}

const triClipThreshold = 32
const maxKdStack = 64 // this probably needs to be tweaked for Go

type kdTree struct {
	costRatio   float // node transversal cost divided by primitive intersection cost
	emptyBonus  float
	maxDepth    int
	maxLeafSize uint
	treeBound   *bound.Bound
	nodes       []kdNode

	prims     []primitive.Primitive
	allBounds []*bound.Bound
	clip      []int
	clipData  []float

	// Statistics!
	depthLimitReached, numBadSplits int
}

func New(prims []primitive.Primitive, depth, leafSize int, costRatio, emptyBonus float) partition.Partitioner {
	// Constants
	const boundFudge = 0.001
	const clipDataSize = 36
	// Create tree!
	tree := &kdTree{costRatio: costRatio, emptyBonus: emptyBonus, maxDepth: depth}
	tree.nodes = make([]kdNode, 0, 256)
	// Calculate maximum depth
	if tree.maxDepth <= 0 {
		tree.maxDepth = int(7.0 + 1.66*math.Log(float64(len(prims))))
	}
	// Calculate leaf size
	logLeaves := fmath.Log2(float(len(prims)))
	if leafSize <= 0 {
		mls := int(logLeaves - 16.0)
		if mls <= 0 {
			mls = 1
		}
		tree.maxLeafSize = uint(mls)
	} else {
		tree.maxLeafSize = uint(leafSize)
	}
	// TODO: if (maxDepth > KD_MAX_STACK)
	if logLeaves > 16.0 {
		tree.costRatio += 0.25 * (logLeaves - 16.0)
	}
	// Calculate bounds
	tree.allBounds = make([]*bound.Bound, len(prims)+triClipThreshold+1)
	for i, prim := range prims {
		b := prim.GetBound()
		tree.allBounds[i] = b
		if i > 0 {
			tree.treeBound = bound.Union(tree.treeBound, b)
		} else {
			tree.treeBound = b
		}
	}
	// Slightly increase tree bound to prevent errors with primitives
	// lying in a bound plane (still slight bug with trees where one dimension is
	// zero)
	{
		a, g := tree.treeBound.Get()
		fudge := vector.New(
			tree.treeBound.GetXLength()*boundFudge,
			tree.treeBound.GetYLength()*boundFudge,
			tree.treeBound.GetZLength()*boundFudge,
		)
		tree.treeBound = bound.New(vector.Sub(a, fudge), vector.Add(g, fudge))
	}
	// Get working memory for tree construction
	leftPrimsSize := len(prims)
	if triClipThreshold*2 > leftPrimsSize {
		leftPrimsSize = triClipThreshold * 2
	}
	leftPrims := make([]int, leftPrimsSize)
	rightPrims := make([]int, len(prims)*3) // just a rough guess, allocating worst case is insane!
	var edges [3][]boundEdge
	for i, _ := range edges {
		edges[i] = make([]boundEdge, 514)
	}
	tree.clip = make([]int, tree.maxDepth+2)
	tree.clipData = make([]float, len(tree.clip)*triClipThreshold*clipDataSize)
	// Prepare data
	for i, _ := range prims {
		leftPrims[i] = i
	}
	for i, _ := range tree.clip {
		tree.clip[i] = -1
	}
	// Build tree
	tree.prims = prims
	tree.build(leftPrims, tree.treeBound, leftPrims, rightPrims, edges, 0, 0)
	return tree
}

func (tree *kdTree) build(primNums []int, nodeBound *bound.Bound, leftPrims, rightPrims []int, edges [3][]boundEdge, depth, badRefines int) int {
	const triClip = false
	// Ensure that the nodes array can fit at least one more node
	if len(tree.nodes) == cap(tree.nodes) {
		newCap := 2 * cap(tree.nodes)
		if newCap > 0x100000 {
			newCap += 0x80000
		}
		n := make([]kdNode, len(tree.nodes), newCap)
		copy(n, tree.nodes)
		tree.nodes = n
	}
	if triClip && len(primNums) <= triClipThreshold {
		// TODO
	}
	// << Check if leaf criteria met >>
	if uint(len(primNums)) <= tree.maxLeafSize || depth >= tree.maxDepth {
		tree.nodes = tree.nodes[0 : len(tree.nodes)+1]
		tree.nodes[len(tree.nodes)-1] = newLeaf(primNums)
		return 0
	}
	// << Calculate cost for all axes and choose minimum >>
	split := splitCost{bestAxis: -1, bestOffset: -1}
	baseBonus := tree.emptyBonus
	tree.emptyBonus *= 1.1 - float(depth)/float(tree.maxDepth)
	switch {
	case len(primNums) > 128:
		tree.pigeonMinCost(primNums, nodeBound, &split)
	case triClip:
		if len(primNums) > triClipThreshold {
			tree.minimalCost(primNums, nodeBound, tree.allBounds, edges, &split)
		} else {
			// TODO: Check if this is right
			tree.minimalCost(primNums, nodeBound, tree.allBounds[len(primNums):], edges, &split)
		}
	default:
		tree.minimalCost(primNums, nodeBound, tree.allBounds, edges, &split)
	}
	tree.emptyBonus = baseBonus // Restore emptyBonus
	// << if minimum > leafcost increase bad refines >>
	if split.bestCost > split.oldCost {
		badRefines++
	}
	if (split.bestCost > 1.6*split.oldCost && len(primNums) < 16) || split.bestAxis == -1 || badRefines == 2 {
		tree.nodes = tree.nodes[0 : len(tree.nodes)+1]
		tree.nodes[len(tree.nodes)-1] = newLeaf(primNums)
		if badRefines == 2 {
			tree.numBadSplits++
		}
		return 0
	}

	// Allocate more memory, if we need it
	newRightPrims := rightPrims
	if len(primNums) > cap(rightPrims) || triClipThreshold*2 > cap(rightPrims) {
		newRightPrims = make([]int, len(primNums)*3)
	}

	// Classify primitives with respect to split
	var splitPos float
	n0, n1 := 0, 0
	switch {
	case len(primNums) > 128: // we did pigeonhole
		for _, pn := range primNums {
			bd := tree.allBounds[pn]
			if a, _ := bd.Get(); a.GetComponent(split.bestAxis) >= split.t {
				newRightPrims[n1] = pn
				n1++
			} else {
				leftPrims[n0] = pn
				n0++
				if _, g := bd.Get(); g.GetComponent(split.bestAxis) > split.t {
					newRightPrims[n1] = pn
					n1++
				}
			}
		}
		splitPos = split.t
	case len(primNums) <= triClipThreshold:
		// TODO
	default: // we did "normal" cost function
		partition := func(prims []int, pos *int, i, endVal int) {
			e := edges[split.bestAxis][*pos]
			if e.end != endVal {
				prims[*pos] = e.primitiveNum
				(*pos)++
			}
		}
		for i := 0; i < split.bestOffset; i++ {
			partition(leftPrims, &n0, i, upperB)
		}
		partition(newRightPrims, &n1, split.bestOffset, bothB)
		for i := split.bestOffset + 1; i < split.nEdge; i++ {
			partition(newRightPrims, &n1, i, lowerB)
		}
		splitPos = edges[split.bestAxis][split.bestOffset].position
	}

	currNode := len(tree.nodes)
	tree.nodes = tree.nodes[0 : len(tree.nodes)+1]
	tree.nodes[currNode] = newInterior(split.bestAxis, splitPos)
	boundL, boundR := bound.New(nodeBound.Get()), bound.New(nodeBound.Get())
	switch split.bestAxis {
	case 0:
		boundL.SetMaxX(splitPos)
		boundR.SetMinX(splitPos)
	case 1:
		boundL.SetMaxY(splitPos)
		boundR.SetMinY(splitPos)
	case 2:
		boundL.SetMaxZ(splitPos)
		boundR.SetMinZ(splitPos)
	}

	if triClip && len(primNums) <= triClipThreshold {
		// TODO
	} else {
		// << Recurse below child >>
		tree.build(leftPrims[0:n0], boundL, leftPrims, newRightPrims, edges, depth+1, badRefines)
		// << Recurse above child >>
		tree.nodes[currNode].(*kdInteriorNode).SetRightChild(len(tree.nodes))
		tree.build(newRightPrims[0:n1], boundR, leftPrims, newRightPrims[n1:], edges, depth+1, badRefines)
	}

	return 1
}

type bin struct {
	n             int
	cLeft, cRight int
	cBLeft, cBoth int
	t             float
}

func (b bin) Empty() bool { return b.n == 0 }
func (b *bin) Reset()     { b.n = 0; b.cLeft = 0; b.cRight = 0; b.cBoth = 0; b.cBLeft = 0 }

func (tree *kdTree) pigeonMinCost(primNums []int, nodeBound *bound.Bound, split *splitCost) {
	const kdBins = 1024
	axisLUT := [3][3]int{[3]int{0, 1, 2}, [3]int{1, 2, 0}, [3]int{2, 0, 1}}

	bins := make([]bin, kdBins+1)
	d := [3]float{nodeBound.GetXLength(), nodeBound.GetYLength(), nodeBound.GetZLength()}
	split.oldCost = float(len(primNums))
	split.bestCost = fmath.Inf
	invTotalSA := 1.0 / (d[0]*d[1] + d[0]*d[2] + d[1]*d[2])

	for axis := 0; axis < 3; axis++ {
		s := kdBins / d[axis]
		min := nodeBound.GetMin().GetComponent(axis)
		// Pigeonhole Sort
		for _, primNum := range primNums {
			bbox := tree.allBounds[primNum]
			tLow, tUp := bbox.GetMin().GetComponent(axis), bbox.GetMax().GetComponent(axis)
			bLeft, bRight := int((tLow-min)*s), int((tUp-min)*s)
			{
				clamp := func(b int) int {
					switch {
					case b < 0:
						return 0
					case b > kdBins:
						return kdBins
					}
					return b
				}
				bLeft, bRight = clamp(bLeft), clamp(bRight)
			}
			if tLow == tUp {
				if bins[bLeft].Empty() || tLow >= bins[bLeft].t {
					bins[bLeft].t = tLow
					bins[bLeft].cBoth++
				} else {
					bins[bLeft].cLeft++
					bins[bLeft].cRight++
				}
				bins[bLeft].n += 2
			} else {
				switch {
				case bins[bLeft].Empty() || tLow > bins[bLeft].t:
					bins[bLeft].t = tLow
					bins[bLeft].cLeft += bins[bLeft].cBoth + bins[bLeft].cBLeft
					bins[bLeft].cRight += bins[bLeft].cBoth
					bins[bLeft].cBoth, bins[bLeft].cBLeft = 0, 1
				case tLow == bins[bLeft].t:
					bins[bLeft].cBLeft++
				default:
					bins[bLeft].cLeft++
				}
				bins[bLeft].n++

				bins[bRight].cRight++
				if bins[bRight].Empty() || tUp > bins[bRight].t {
					bins[bRight].t = tUp
					bins[bRight].cLeft += bins[bRight].cBoth + bins[bRight].cBLeft
					bins[bRight].cRight += bins[bRight].cBoth
					bins[bRight].cBoth, bins[bRight].cBLeft = 0, 0
				}
				bins[bRight].n++
			}
		}

		capArea := d[axisLUT[1][axis]] * d[axisLUT[2][axis]]
		capPerim := d[axisLUT[1][axis]] + d[axisLUT[2][axis]]

		// Accumulate primitives and evaluate cost
		nBelow, nAbove := 0, len(primNums)
		for i, b := range bins {
			if !b.Empty() {
				nBelow += b.cLeft
				nAbove -= b.cRight
				// Cost
				edget := b.t
				if edget > nodeBound.GetMin().GetComponent(axis) && edget < nodeBound.GetMax().GetComponent(axis) {
					// Compute cost at ith edge
					l1 := edget - nodeBound.GetMin().GetComponent(axis)
					l2 := nodeBound.GetMax().GetComponent(axis) - edget
					belowSA := capArea + l1*capPerim
					aboveSA := capArea + l2*capPerim
					rawCosts := belowSA*float(nBelow) + aboveSA*float(nAbove)
					eb := 0.0
					if nAbove == 0 {
						eb = (0.1 + l2/d[axis]) * tree.emptyBonus * rawCosts
					} else if nBelow == 0 {
						eb = (0.1 + l1/d[axis]) * tree.emptyBonus * rawCosts
					}
					// Update best split if this is lowest cost so far
					cost := tree.costRatio + invTotalSA*(rawCosts-eb)
					if cost < split.bestCost {
						split.t = edget
						split.bestCost = cost
						split.bestAxis = axis
						split.bestOffset = i // kinda useless...
						split.nBelow = nBelow
						split.nAbove = nAbove
					}
				}
				nBelow += b.cBoth + b.cBLeft
				nAbove -= b.cBoth
			}
		}

		if nBelow != len(primNums) || nAbove != 0 {
			// SCREWED.
			panic("Cost function mismatch")
		}

		for i, _ := range bins {
			bins[i].Reset()
		}
	}
}

func (tree *kdTree) minimalCost(primNums []int, nodeBound *bound.Bound, pBounds []*bound.Bound, edges [3][]boundEdge, split *splitCost) {
	// TODO
}

type kdStackFrame struct {
	node kdNode
	t    float
	pb   vector.Vector3D
	prev int
}

type collideResult struct {
	prim  primitive.Primitive
	depth float
}

func (tree *kdTree) collide(r ray.Ray, minDist, maxDist float) (<-chan collideResult, chan<- bool) {
	ch := make(chan collideResult)
	signal := make(chan bool, 1)

	// Quick check: If we're not even in the ballpark, then don't spawn a
	// goroutine.
	var a, b float
	var crosses bool
	if a, b, crosses = tree.treeBound.Cross(r.From, r.Dir, maxDist); !crosses {
		close(ch)
		return ch, signal
	}

	// Now start the interesting stuff.
	go func() {
		defer close(ch) // This channel should last as long as the goroutine.

		var t float
		invDir := vector.New(1.0/r.Dir.X, 1.0/r.Dir.Y, 1.0/r.Dir.Z)
		stack := make([]kdStackFrame, maxKdStack)
		currIndex := 0
		farIndex := 0

		enterPt := 0
		stack[enterPt].t = a

		if a >= 0.0 {
			// Ray with external origin
			stack[enterPt].pb = vector.Add(r.From, vector.ScalarMul(r.Dir, a))
		} else {
			// Ray with internal origin
			stack[enterPt].pb = r.From
		}

		// Setup initial entry and exit point in stack
		exitPt := 1
		stack[exitPt].t = b
		stack[exitPt].pb = vector.Add(r.From, vector.ScalarMul(r.Dir, b))
		stack[exitPt].node = nil

		for currIndex != -1 {
			currNode := tree.nodes[currIndex]
			if maxDist < stack[enterPt].t {
				break
			}
			for !currNode.IsLeaf() {
				axis := currNode.(*kdInteriorNode).GetSplitAxis()
				splitVal := currNode.(*kdInteriorNode).GetSplitPos()
				if stack[enterPt].pb.GetComponent(axis) <= splitVal {
					if stack[exitPt].pb.GetComponent(axis) <= splitVal {
						currIndex++
						continue
					} else if stack[exitPt].pb.GetComponent(axis) == splitVal {
						currIndex = currNode.(*kdInteriorNode).GetRightChild()
						continue
					} else {
						farIndex = currNode.(*kdInteriorNode).GetRightChild()
						currIndex++
						currNode = tree.nodes[currIndex]
					}
				} else {
					if splitVal < stack[exitPt].pb.GetComponent(axis) {
						currIndex = currNode.(*kdInteriorNode).GetRightChild()
						continue
					}
					farIndex = currIndex + 1
					currIndex = currNode.(*kdInteriorNode).GetRightChild()
					currNode = tree.nodes[currIndex]
				}
				// Traverse both children
				t = (splitVal - r.From.GetComponent(axis)) * invDir.GetComponent(axis)
				// Set up the new exit point
				prevExitPt := exitPt
				exitPt++
				// Possibly skip current entry point so not to overwrite the data
				if exitPt == enterPt {
					exitPt++
				}
				// Push values onto the stack
				nextAxis := (axis + 1) % 3
				prevAxis := (axis + 2) % 3
				stack[exitPt].prev = prevExitPt
				stack[exitPt].t = t
				stack[exitPt].node = tree.nodes[farIndex]
				// TODO: SetAxis?
				_, _ = nextAxis, prevAxis
				//stack[exitPt].pb[axis] = splitVal
			}

			// Check for intersections inside leaf node
			for _, index := range currNode.(*kdLeafNode).GetPrimitives() {
				mp := tree.prims[index]
				depth, hit := mp.Intersect(r)
				if hit && depth < maxDist && depth > minDist {
					// It's a hit!  Send it back!
					ch <- collideResult{mp, depth}
					// Now check to see whether we can stop.
					if !<-signal {
						// Caller wants us to terminate.
						return
					}
				}
			}
			enterPt, currIndex = exitPt, exitPt
			exitPt = stack[enterPt].prev
		}
	}()
	return ch, signal
}

func (tree *kdTree) Intersect(r ray.Ray, dist float) (hit bool, prim primitive.Primitive, z float) {
	ch, signal := tree.collide(r, r.TMin, dist)
	signal <- false
	result := <-ch
	if result.prim == nil {
		hit = false
	} else {
		hit = true
		prim = result.prim
		z = result.depth
	}
	return
}

func (tree *kdTree) IntersectS(r ray.Ray, dist float) (hit bool, prim primitive.Primitive) {
	ch, signal := tree.collide(r, r.TMin, dist)
	signal <- false
	result := <-ch
	if result.prim == nil {
		hit = false
	} else {
		hit = true
		prim = result.prim
	}
	return
}

func (tree *kdTree) IntersectTS(state *render.State, r ray.Ray, maxDepth int, dist float, filt *color.Color) (hit bool, prim primitive.Primitive) {
	ch, signal := tree.collide(r, r.TMin, dist)
	filtered := make(map[primitive.Primitive]bool)
	depth := 0
	for result := range ch {
		hit = true
		prim = result.prim
		mat := prim.GetMaterial()
		if !mat.IsTransparent() {
			signal <- false
			return
		}
		if found, _ := filtered[prim]; !found {
			filtered[prim] = true
			if depth < maxDepth {
				h := vector.Add(r.From, vector.ScalarMul(r.Dir, result.depth))
				sp := prim.GetSurface(h)
				*filt = color.Mul(*filt, mat.GetTransparency(state, sp, r.Dir))
				depth++
			} else {
				// We've hit the depth limit.  Cut it off.
				signal <- false
				return
			}
		}
	}
	return
}

func (tree *kdTree) GetBound() *bound.Bound { return tree.treeBound }

type kdNode interface {
	IsLeaf() bool
}

type kdInteriorNode struct {
	division   float
	axis       int
	rightChild int
}

func newInterior(axis int, d float) *kdInteriorNode {
	return &kdInteriorNode{division: d, axis: axis}
}

func (node *kdInteriorNode) IsLeaf() bool        { return false }
func (node *kdInteriorNode) GetSplitPos() float  { return node.division }
func (node *kdInteriorNode) GetSplitAxis() int   { return node.axis }
func (node *kdInteriorNode) GetRightChild() int  { return node.rightChild }
func (node *kdInteriorNode) SetRightChild(i int) { node.rightChild = i }

type kdLeafNode struct {
	primitives []int
}

func newLeaf(prims []int) *kdLeafNode {
	return &kdLeafNode{prims}
}

func (node *kdLeafNode) IsLeaf() bool         { return false }
func (node *kdLeafNode) GetPrimitives() []int { return node.primitives }
