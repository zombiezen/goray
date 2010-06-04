//
//  fmath.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

package fmath

import "math"

var Inf = float(math.Inf(1))

type unaryFunc func(float64) float64

func callUnary(f unaryFunc, n float) float {
	return float(f(float64(n)))
}

type binaryFunc func(float64, float64) float64

func callBinary(f binaryFunc, n1, n2 float) float {
	return float(f(float64(n1), float64(n2)))
}

func Abs(f float) float { return callUnary(abs64, f) }

func abs64(f float64) float64 {
	if f >= 0 {
		return f
	}
	return -f
}

func Eq(a, b float) bool {
	const epsilon = 0.000001
	return nearlyEqual(float64(a), float64(b), epsilon)
}

/*
   nearlyEqual checks whether two numbers are equivalent.
   This is taken from http://floating-point-gui.de/
*/
func nearlyEqual(a, b, epsilon float64) bool {
	absA, absB := abs64(a), abs64(b)
	diff := abs64(a - b)

	if a*b == 0 {
		// A or B is zero; relative error is not meaningful
		return diff < epsilon*epsilon
	}
	// Use relative error
	return diff/(absA+absB) < epsilon
}

func IsInf(n float) bool {
	return math.IsInf(float64(n), 0)
}

func Min(f1, f2 float, fn ...float) (min float) {
	min = f1
	if f2 < f1 {
		min = f2
	}
	for _, f := range fn {
		if f < min {
			min = f
		}
	}
	return
}

func Max(f1, f2 float, fn ...float) (max float) {
	max = f1
	if f2 > f1 {
		max = f2
	}
	for _, f := range fn {
		if f > max {
			max = f
		}
	}
	return
}

func Mod(f1, f2 float) float {
	return callBinary(math.Fmod, f1, f2)
}

func Sqrt(f float) float {
	return callUnary(math.Sqrt, f)
}

func Asin(f float) float     { return callUnary(math.Asin, f) }
func Acos(f float) float     { return callUnary(math.Acos, f) }
func Atan(f float) float     { return callUnary(math.Atan, f) }
func Atan2(y, x float) float { return callBinary(math.Atan2, y, x) }

func Sin(f float) float { return callUnary(math.Sin, f) }
func Cos(f float) float { return callUnary(math.Cos, f) }
func Tan(f float) float { return callUnary(math.Tan, f) }

func Log(f float) float   { return callUnary(math.Log, f) }
func Log2(f float) float  { return callUnary(math.Log2, f) }
func Log10(f float) float { return callUnary(math.Log10, f) }
