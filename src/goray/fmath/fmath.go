//
//	goray/fmath/fmath.go
//	goray
//
//	Created by Ross Light on 2010-05-23.
//

/*
	The fmath package provides convenient functions for performing
	floating-point math.

	Most functionality could be done with the standard math package (and indeed,
	many functions here will use them internally), but using this package allows
	us to use the compiler's native amount of precision.
*/
package fmath

import "math"

// Inf stores positive infinity.
var Inf = float(math.Inf(1))

type unaryFunc func(float64) float64

func callUnary(f unaryFunc, n float) float {
	return float(f(float64(n)))
}

type binaryFunc func(float64, float64) float64

func callBinary(f binaryFunc, n1, n2 float) float {
	return float(f(float64(n1), float64(n2)))
}

// Abs returns the absolute value of its argument.
func Abs(f float) float { return callUnary(abs64, f) }

func abs64(f float64) float64 {
	if f >= 0 {
		return f
	}
	return -f
}

// IsInf returns whether its argument is one of the infinities.
func IsInf(n float) bool {
	return math.IsInf(float64(n), 0)
}

// Min returns the argument that is closest to negative infinity.
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

// Max returns the argument that is closest to positive infinity.
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

// Mod performs a floating-point modulus operation.
func Mod(f1, f2 float) float {
	return callBinary(math.Fmod, f1, f2)
}

// Sqrt returns the square root of its argument.
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
