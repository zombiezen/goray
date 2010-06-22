//
//  goray/core/kdtree/split.go
//  goray
//
//  Created by Ross Light on 2010-06-21.
//

package kdtree

import (
	container "container/vector"
	"sort"
	"goray/fmath"
	"goray/logging"
	"goray/core/bound"
)

type SplitFunc func([]Value, *bound.Bound, BuildState) (axis int, pivot float, cost float)

func DefaultSplit(vals []Value, bd *bound.Bound, state BuildState) (axis int, pivot float, cost float) {
	const pigeonThreshold = 128

	if len(vals) > pigeonThreshold {
		return PigeonSplit(vals, bd, state)
	}
	return MinimalSplit(vals, bd, state)
}

func SimpleSplit(vals []Value, bd *bound.Bound, state BuildState) (axis int, pivot float, cost float) {
	axis = bd.GetLargestAxis()
	data := make([]float, 0, len(vals)*2)
	for _, v := range vals {
		min, max := state.GetDimension(v, axis)
		if fmath.Eq(min, max) {
			i := len(data)
			data = data[0 : len(data)+1]
			data[i] = min
		} else {
			i := len(data)
			data = data[0 : len(data)+2]
			data[i], data[i+1] = min, max
		}
	}
	sort.SortFloats(data)
	pivot = data[len(data)/2]
	return
}

func PigeonSplit(vals []Value, bd *bound.Bound, state BuildState) (bestAxis int, bestPivot float, bestCost float) {
	const numBins = 1024
	type pigeonBin struct {
		n           int
		left, right int
		bleft, both int
		t           float
	}

	var bins [numBins + 1]pigeonBin
	d := [3]float{bd.GetXLength(), bd.GetYLength(), bd.GetZLength()}
	bestCost = fmath.Inf
	totalSA := d[0]*d[1] + d[0]*d[2] + d[1]*d[2]
	invTotalSA := 0.0
	if !fmath.Eq(totalSA, 0.0) {
		invTotalSA = 1.0 / totalSA
	}

	for axis := 0; axis < 3; axis++ {
		s := numBins / d[axis]
		min := bd.GetMin().GetComponent(axis)

		for _, v := range vals {
			tLow, tHigh := state.GetDimension(v, axis)
			bLeft, bRight := int((tLow-min)*s), int((tHigh-min)*s)
			if bLeft < 0 {
				bLeft = 0
			} else if bLeft > numBins {
				bLeft = numBins
			}
			if bRight < 0 {
				bRight = 0
			} else if bRight > numBins {
				bRight = numBins
			}

			if tLow == tHigh {
				if bins[bLeft].n == 0 || tLow >= bins[bLeft].t {
					bins[bLeft].t = tLow
					bins[bLeft].both++
				} else {
					bins[bLeft].left++
					bins[bLeft].right++
				}
				bins[bLeft].n += 2
			} else {
				if bins[bLeft].n == 0 || tLow > bins[bLeft].t {
					bins[bLeft].t = tLow
					bins[bLeft].left += bins[bLeft].both + bins[bLeft].bleft
					bins[bLeft].right += bins[bLeft].both
					bins[bLeft].both, bins[bLeft].bleft = 0, 0
					bins[bLeft].bleft++
				} else if tLow == bins[bLeft].t {
					bins[bLeft].bleft++
				} else {
					bins[bLeft].left++
				}

				bins[bLeft].n++
				bins[bRight].right++
				if bins[bRight].n == 0 || tHigh > bins[bRight].t {
					bins[bRight].t = tHigh
					bins[bRight].left += bins[bRight].both + bins[bRight].bleft
					bins[bRight].right += bins[bRight].both
					bins[bRight].both, bins[bRight].bleft = 0, 0
				}
				bins[bRight].n++
			}
		}

		capArea := d[(axis+1)%3] * d[(axis+2)%3]
		capPerim := d[(axis+1)%3] + d[(axis+2)%3]

		nBelow, nAbove := 0, len(vals)
		// Cumulate values and evaluate cost
		for _, b := range bins {
			if b.n != 0 {
				nBelow += b.left
				nAbove -= b.right
				// Cost:
				edget := b.t
				if edget > bd.GetMin().GetComponent(axis) && edget < bd.GetMax().GetComponent(axis) {
					cost := computeCost(axis, bd, capArea, capPerim, invTotalSA, nBelow, nAbove, edget)
					if cost < bestCost {
						bestAxis, bestPivot, bestCost = axis, edget, cost
					}
				}

				nBelow += b.both + b.bleft
				nAbove -= b.both
			}
		}

		if nBelow != len(vals) || nAbove != 0 {
			// SCREWED.
			logging.Error(state.Log, "Pigeon cost failed; %d above and %d below (should be %d)", nAbove, nBelow, len(vals))
			panic("Cost function mismatch")
		}

		// Reset all bins
		for i, _ := range bins {
			bins[i] = pigeonBin{}
		}
	}

	return
}

func computeCost(axis int, bd *bound.Bound, capArea, capPerim, invTotalSA float, nBelow, nAbove int, edget float) float {
	const emptyBonus = 0.33
	const costRatio = 0.35

	l1, l2 := edget-bd.GetMin().GetComponent(axis), bd.GetMax().GetComponent(axis)-edget
	belowSA, aboveSA := capArea+l1*capPerim, capArea+l2*capPerim
	rawCosts := belowSA*float(nBelow) + aboveSA*float(nAbove)

	d := 0.0
	switch axis {
	case 0:
		d = bd.GetXLength()
	case 1:
		d = bd.GetYLength()
	case 2:
		d = bd.GetZLength()
	}

	eb := 0.0
	if nAbove == 0 {
		eb = (0.1 + l2/d) * emptyBonus * rawCosts
	} else if nBelow == 0 {
		eb = (0.1 + l1/d) * emptyBonus * rawCosts
	}

	return costRatio + invTotalSA*(rawCosts-eb)
}

type boundEdge struct {
	position float
	boundEnd int
}

func (e boundEdge) Less(other interface{}) bool {
	f, ok := other.(boundEdge)
	if !ok {
		return false
	}
	if fmath.Eq(e.position, f.position) {
		return e.boundEnd > f.boundEnd
	}
	return e.position < f.position
}

const (
	lowerB = iota
	bothB
	upperB
)

func MinimalSplit(vals []Value, bd *bound.Bound, state BuildState) (bestAxis int, bestPivot float, bestCost float) {
	d := [3]float{bd.GetXLength(), bd.GetYLength(), bd.GetZLength()}
	bestCost = fmath.Inf
	totalSA := d[0]*d[1] + d[0]*d[2] + d[1]*d[2]
	invTotalSA := 0.0
	if !fmath.Eq(totalSA, 0.0) {
		invTotalSA = 1.0 / totalSA
	}

	for axis := 0; axis < 3; axis++ {
		edges := new(container.Vector)
		edges.Resize(0, len(vals)*2)
		for _, v := range vals {
			min, max := state.GetDimension(v, axis)
			if fmath.Eq(min, max) {
				edges.Push(boundEdge{min, bothB})
			} else {
				edges.Push(boundEdge{min, lowerB})
				edges.Push(boundEdge{max, upperB})
			}
		}
		sort.Sort(edges)

		capArea := d[(axis+1)%3] * d[(axis+2)%3]
		capPerim := d[(axis+1)%3] + d[(axis+2)%3]

		nBelow, nAbove := 0, len(vals)
		for tmp := range edges.Iter() {
			e := tmp.(boundEdge)
			if e.boundEnd == upperB {
				nAbove--
			}

			if e.position > bd.GetMin().GetComponent(axis) && e.position < bd.GetMax().GetComponent(axis) {
				cost := computeCost(axis, bd, capArea, capPerim, invTotalSA, nAbove, nBelow, e.position)
				if cost < bestCost {
					bestAxis, bestPivot, bestCost = axis, e.position, cost
				}
			}

			if e.boundEnd != upperB {
				nBelow++
				if e.boundEnd == bothB {
					nAbove--
				}
			}
		}

		if nBelow != len(vals) || nAbove != 0 {
			logging.Error(state.Log, "Minimal cost failed; %d above and %d below (should be %d)", nAbove, nBelow, len(vals))
			panic("Cost function mismatch")
		}
	}

	return
}
