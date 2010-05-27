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

func Abs(f float) float {
	if f >= 0 {
		return f
	}
	return -f
}

func Cos(f float) float {
	return callUnary(math.Cos, f)
}

func IsInf(n float) bool {
	return math.IsInf(float64(n), 0)
}

func Min(f1, f2 float) float {
	if f1 <= f2 {
		return f1
	}
	return f2
}

func Max(f1, f2 float) float {
	if f1 >= f2 {
		return f1
	}
	return f2
}

func Mod(f1, f2 float) float {
	return callBinary(math.Fmod, f1, f2)
}

func Sin(f float) float {
	return callUnary(math.Sin, f)
}

func Sqrt(f float) float {
	return callUnary(math.Sqrt, f)
}

func Tan(f float) float {
	return callUnary(math.Tan, f)
}
