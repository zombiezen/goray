//
//  goray/partition.go
//  goray
//
//  Created by Ross Light on 2010-05-29.
//

package partition

import (
	"math"
)

import (
	"./goray/bound"
	"./goray/color"
	"./goray/primitive"
	"./goray/render"
	"./goray/ray"
)

type Partitioner interface {
	Intersect(r ray.Ray, dist float) (hit bool, prim primitive.Primitive, z float)
	IntersectS(r ray.Ray, dist float) (hit bool, prim primitive.Primitive)
	IntersectTS(state render.State, r ray.Ray, maxDepth int, dist float) (hit bool, prim primitive.Primitive, filt color.Color)
	GetBound() *bound.Bound
}

const (
    lowerB = 0
    upperB = 2
    bothB = 1
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

type bin struct {
    n int
    cLeft, cRight int
    cBLeft, cBoth int
    t float
}

func (b bin) Empty() bool { return n == 0 }
func (b *bin) Reset() { b.n = 0; b.cLeft = 0; b.cRight = 0; b.cBoth = 0; b.cBLeft = 0 }

type kdTree struct {
	costRatio   float // node transversal cost divided by primitive intersection cost
	emptyBonus  float
	maxDepth    int
	maxLeafSize uint
	treeBound   *bound.Bound
	nodes       []interface{}

	prims     []primitive.Primitive
	allBounds []*bound.Bound
	clip      []int
	clipData  []byte

	// Statistics!
	depthLimitReached, numBadSplits int
}

func NewKDTree(prims []primitive.Primitive, depth, leafSize int, costRatio, emptyBonus float) Partitioner {
	// Constants
	const triClipThreshold = 32
	const boundFudge = 0.001
	// Create tree!
	tree := &kdTree{costRatio: costRatio, emptyBonus: emptyBonus, maxDepth: depth}
	tree.nodes = make([]interface{}, 0, 256)
	// Calculate maximum depth
	if tree.maxDepth <= 0 {
		tree.maxDepth = int(7.0 + 1.66*math.Log(float64(len(prims))))
	}
	// Calculate leaf size
	logLeaves := math.Log2(float64(len(prims)))
	if leafSize <= 0 {
		mls := int(logLeaves - 16.0)
		if mls <= 0 {
			mls = 1
		}
		tree.maxLeafSize = uint(mls)
	} else {
		tree.maxLeafSize = leafSize
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
	edges := make([][]boundEdge, 3)
	for i, _ := range edges {
		edges[i] = make([]boundEdge, 514)
	}
	tree.clip = make([]int, tree.maxDepth+2)
	tree.clipData = make([]byte, len(tree.clip)*triClipThreshold*clipDataSize)
	// Prepare data
	for i, _ := range prims {
		leftPrims[i] = i
	}
	for i, _ := range tree.clip {
		tree.clip[i] = -1
	}
	// Build tree
	tree.prims = prims
	tree.build()
	return tree
}

func (tree *kdTree) build(primNums []int, nodeBound *bound.Bound, leftPrims, rightPrims []int, edges [][]boundEdge, depth, badRefines int) int {
	const triClip = false
	if len(tree.nodes) == cap(tree.nodes) {
		newCap := 2 * cap(tree.nodes)
		if newCap > 0x100000 {
			newCap += 0x80000
		}
		n := make([]interface{}, len(tree.nodes), newCap)
		copy(n, tree.nodes)
		tree.nodes = n
	}
	if triClip && len(primNums) <= triClipThreshold {
		// TODO
	}
	// << Check if leaf criteria met >>
	if len(primNums) <= tree.maxLeafSize || depth >= tree.maxDepth {
		tree.nodes = tree.nodes[0 : len(tree.nodes)+1]
		tree.nodes[len(tree.nodes)-1] = NewLeaf(primNums)
		return 0
	}
	// << Calculate cost for all axes and choose minimum >>
    split := splitCost{bestAxis:-1, bestOffset:-1}
	baseBonus := tree.emptyBonus
    tree.emptyBonus *= 1.1 - float(depth) / float(tree.maxDepth)
    switch {
        case len(prims) > 128:
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
    if (split.bestCost > 1.6 * split.oldCost && len(prims) < 16) || split.bestAxis == -1 || badRefines == 2 {
        tree.nodes = tree.nodes[0 : len(tree.nodes)+1]
        tree.nodes[len(tree.nodes)-1] = NewLeaf(primNums)
        if badRefines == 2 {
            tree.numBadSplits++
        }
        return 0
    }
    
    // Allocate more memory, if we need it
    oldRightPrims, newRightPrims := rightPrims, rightPrims
    if len(prims) > cap(rightPrims) || triClipThreshold * 2 > cap(rightPrims) {
        newRightPrims = make([]int, len(primNums) * 3)
    }
    
    // Classify primitives with respect to split
    var splitPos float
    n0, n1 := 0, 0
    switch {
        case len(primNums) > 128: // we did pigeonhole
            for i, pn := range prims {
                bd := allBounds[pn]
                if a, _ := bd.Get(); a.GetAxis(split.bestAxis) >= split.t {
                    newRightPrims[n1] = pn
                    n1++
                } else {
                    leftPrims[n0] = pn
                    n0++
                    if _, g := bd.Get(); g.GetAxis(split.bestAxis) > split.t {
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
                e := edges[split.bestAxis][pos]
                if e.end != endVal {
                    prims[*pos] = e
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
            splitPos = edges[split.bestAxis][split.bestOffset].pos
    }
}

func (tree *kdTree) pigeonMinCost(primNums []int, nodeBound *bound.Bound, split *splitCost) {
    const kdBins = 1024
    const axisLUT = [3][3]int{[3]int{0,1,2}, [3]int{1,2,0}, [3]int{2,0,1}}
    
    bins := make([]bin, kdBins + 1)
    d := [3]float{nodeBound.GetLengthX(), nodeBound.GetLengthY(), nodeBound.GetLengthZ()}
    split.oldCost = float(len(primNums))
    split.bestCost = fmath.Inf
    invTotalSA := 1.0 / (d[0]*d[1] + d[0]*d[2] + d[1]*d[2])
    
    for axis := 0; axis < 3; axis++ {
        s := kdBins / d[axis]
        min := nodeBound.GetMin().GetAxis(axis)
        // Pigeonhole Sort
        for i, primNum := range primNums {
            bbox := tree.allBounds[primNum]
            tLow, tUp := bbox.GetMin().GetAxis(axis), bbox.GetMax().GetAxis(axis)
            bLeft, bRight := int((tLow - min) * s), int((tUp - min) * s)
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
                    bins[bRight].cBoth = bins[bRight].cBLeft = 0
                }
                bins[bRight].n++
            }
        }
        
        capArea := d[axisLUT[1][axis]] * d[axisLut[2][axis]]
        capPerim := d[axisLut[1][axis]] + d[axisLut[2][axis]]
        
        // Accumulate primitives and evaluate cost
        nBelow, nAbove := 0, nPrims
        for i, b := range bins {
            if !b.Empty() {
                nBelow += b.cLeft
                nAbove -= b.cRight
                // Cost
                edget := b.t
                if edget > nodeBound.GetMin().GetAxis(axis) && edget < nodeBound.GetMax().GetAxis(axis) {
                    // Compute cost at ith edge
                    l1 := edget - nodeBound.GetMin().GetAxis(axis)
                    l2 := nodeBound.GetMax().GetAxis(axis) - edget
                    belowSA := capArea + l1 * capPerim
                    aboveSA := capArea + l2 * capPerim
                    rawCosts := belowSA * nBelow + aboveSA * nAbove
                    eb := 0.0
                    if nAbove == 0 {
                        eb = (0.1 + l2 / d[axis]) * tree.emptyBonus * rawCosts
                    } else if nBelow == 0 {
                        eb = (0.1 + l1 / d[axis]) * tree.emptyBonus * rawCosts
                    }
                    // Update best split if this is lowest cost so far
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

func (tree *kdTree) minimalCost(primNums []int, nodeBound *bound.Bound, pBounds []*bound.Bound, edges [3]boundEdge, split *splitCost) {
}

type kdInteriorNode struct {
	division   float
	axis       int
	rightChild int
}

func newInterior(axis int, d float) *kdInteriorNode {
	return &kdInteriorNode{division: d, axis: axis}
}

func (node kdInteriorNode) GetSplitPos() float   { return node.division }
func (node kdInteriorNode) GetSplitAxis() int    { return node.axis }
func (node kdInteriorNode) GetRightChild() int   { return node.rightChild }
func (node *kdInteriorNode) SetRightChild(i int) { node.rightChild = i }

type kdLeafNode struct {
	primitives []int
}

func newLeaf(prims []int) *kdLeafNode {
	return &kdLeafNode{prims}
}
func (node kdLeafNode) GetPrimitives() []int { return node.axis }
