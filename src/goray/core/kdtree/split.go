//
//	goray/core/kdtree/split.go
//	goray
//
//	Created by Ross Light on 2010-06-21.
//

package kdtree

import (
	"math"
	"sort"
	"goray/logging"
	"goray/core/bound"
	"goray/core/vector"
)

type SplitFunc func([]Value, *bound.Bound, BuildState) (axis vector.Axis, pivot float64, cost float64)

func DefaultSplit(vals []Value, bd *bound.Bound, state BuildState) (axis vector.Axis, pivot float64, cost float64) {
	const pigeonThreshold = 128

	if len(vals) > pigeonThreshold {
		return PigeonSplit(vals, bd, state)
	}
	return MinimalSplit(vals, bd, state)
}

type float64array []float64

func (a float64array) Len() int             { return len(a) }
func (a float64array) Less(i1, i2 int) bool { return a[i1] < a[i2] }
func (a float64array) Swap(i1, i2 int)      { a[i1], a[i2] = a[i2], a[i1] }

func SimpleSplit(vals []Value, bd *bound.Bound, state BuildState) (axis vector.Axis, pivot float64, cost float64) {
	axis = bd.LargestAxis()
	data := make([]float64, 0, len(vals)*2)
	for i, v := range vals {
		min, max := state.getClippedDimension(i, v, axis)
		if min == max {
			data = append(data, min)
		} else {
			data = append(data, min, max)
		}
	}
	sort.Sort(float64array(data))
	pivot = data[len(data)/2]
	return
}

func PigeonSplit(vals []Value, bd *bound.Bound, state BuildState) (bestAxis vector.Axis, bestPivot float64, bestCost float64) {
	const numBins = 1024
	type pigeonBin struct {
		n           int
		left, right int
		bleft, both int
		t           float64
	}

	var bins [numBins + 1]pigeonBin
	d := [3]float64{bd.LengthX(), bd.LengthY(), bd.LengthZ()}
	bestCost = math.Inf(1)
	totalSA := d[0]*d[1] + d[0]*d[2] + d[1]*d[2]
	invTotalSA := 0.0
	if totalSA != 0.0 {
		invTotalSA = 1.0 / totalSA
	}

	for axis := vector.X; axis <= vector.Z; axis++ {
		s := numBins / d[axis]
		min := bd.Min()[axis]

		for i, v := range vals {
			tLow, tHigh := state.getClippedDimension(i, v, axis)
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

		capArea := d[axis.Next()] * d[axis.Prev()]
		capPerim := d[axis.Next()] + d[axis.Prev()]

		nBelow, nAbove := 0, len(vals)
		// Cumulate values and evaluate cost
		for _, b := range bins {
			if b.n != 0 {
				nBelow += b.left
				nAbove -= b.right
				// Cost:
				edget := b.t
				if edget > bd.Min()[axis] && edget < bd.Max()[axis] {
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

func computeCost(axis vector.Axis, bd *bound.Bound, capArea, capPerim, invTotalSA float64, nBelow, nAbove int, edget float64) float64 {
	const emptyBonus = 0.33
	const costRatio = 0.35

	l1, l2 := edget-bd.Min()[axis], bd.Max()[axis]-edget
	belowSA, aboveSA := capArea+l1*capPerim, capArea+l2*capPerim
	rawCosts := belowSA*float64(nBelow) + aboveSA*float64(nAbove)

	d := bd.Size()[axis]

	var eb float64
	if nAbove == 0 {
		eb = (0.1 + l2/d) * emptyBonus * rawCosts
	} else if nBelow == 0 {
		eb = (0.1 + l1/d) * emptyBonus * rawCosts
	}

	return costRatio + invTotalSA*(rawCosts-eb)
}

type boundEdge struct {
	position float64
	boundEnd int
}

type boundEdgeArray []boundEdge

func (a boundEdgeArray) Len() int {
	return len(a)
}

func (a boundEdgeArray) Swap(i1, i2 int) {
	a[i1], a[i2] = a[i2], a[i1]
}

func (a boundEdgeArray) Less(i1, i2 int) bool {
	e, f := a[i1], a[i2]
	if e.position == f.position {
		return e.boundEnd > f.boundEnd
	}
	return e.position < f.position
}

const (
	lowerB = iota
	bothB
	upperB
)

func MinimalSplit(vals []Value, bd *bound.Bound, state BuildState) (bestAxis vector.Axis, bestPivot float64, bestCost float64) {
	d := bd.Size()
	bestCost = math.Inf(1)
	totalSA := d[0]*d[1] + d[0]*d[2] + d[1]*d[2]
	invTotalSA := 0.0
	if totalSA != 0.0 {
		invTotalSA = 1.0 / totalSA
	}

	for axis := vector.X; axis <= vector.Z; axis++ {
		edges := make(boundEdgeArray, 0, len(vals)*2)
		for i, v := range vals {
			min, max := state.getClippedDimension(i, v, axis)
			if min == max {
				edges = append(edges, boundEdge{min, bothB})
			} else {
				edges = append(edges, boundEdge{min, lowerB}, boundEdge{max, upperB})
			}
		}
		sort.Sort(edges)

		capArea := d[axis.Next()] * d[axis.Prev()]
		capPerim := d[axis.Next()] + d[axis.Prev()]

		nBelow, nAbove := 0, len(vals)
		for _, e := range edges {
			if e.boundEnd == upperB {
				nAbove--
			}

			if e.position > bd.Min()[axis] && e.position < bd.Max()[axis] {
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
