//
//	goray/time.go
//	goray
//
//	Created by Ross Light on 2010-06-09.
//

/* The goray/time package provides a convenient way to time an operation. */
package time

import (
	"fmt"
	stdtime "time"
)

const (
	Second = 1
	Minute = 60 * Second
	Hour   = 60 * Minute
	Day    = 24 * Hour
)

/* Stopwatch calls a function and returns how long it took for the function to return. */
func Stopwatch(f func()) Time {
	startTime := stdtime.Nanoseconds()
	f()
	endTime := stdtime.Nanoseconds()
	return Time(float64(endTime-startTime) * 1e-9)
}

/* Time is a type that represents a duration of time. */
type Time float64

/* Add returns the sum of two or more Time values. */
func Add(t1, t2 Time, tn ...Time) Time {
	accum := float64(t1) + float64(t2)
	for _, t := range tn {
		accum += float64(t)
	}
	return Time(accum)
}

/* Split returns the components of the time in hours, minutes, and seconds. */
func (t Time) Split() (hours, minutes int, seconds float64) {
	seconds = float64(t)
	hours = int(t / Hour)
	seconds -= float64(hours * Hour)
	minutes = int(seconds / Minute)
	seconds -= float64(minutes * Minute)
	return
}

func (t Time) String() string {
	h, m, s := t.Split()
	switch {
	case h == 0 && m == 0:
		return fmt.Sprintf("%.3fs", s)
	case h == 0:
		return fmt.Sprintf("%02d:%05.2f", m, s)
	}
	return fmt.Sprintf("%d:%02d:%05.2f", h, m, s)
}
