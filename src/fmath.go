//
//  fmath.go
//  goray
//
//  Created by Ross Light on 2010-05-23.
//

package fmath

import "math"

func Abs(f float) float {
    if f >= 0 {
        return f
    }
    return -f
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

func Sqrt(f float) float {
    return float(math.Sqrt(float64(f)))
}
