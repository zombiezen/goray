/*
	Copyright (c) 2011 Ross Light.
	Copyright (c) 2005 Mathias Wein, Alejandro Conty, and Alfredo de Greef.

	This file is part of goray.

	goray is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	goray is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with goray.  If not, see <http://www.gnu.org/licenses/>.
*/

// Package time is a convenient way to time an operation.
package time

import (
	"fmt"
	"math"
	stdtime "time"
)

// Different units of time
const (
	Nanosecond = 1e-9
	Second     = 1
	Minute     = 60 * Second
	Hour       = 60 * Minute
	Day        = 24 * Hour
)

// Stopwatch calls a function and returns how long it took for the function to return.
func Stopwatch(f func()) Time {
	startTime := stdtime.Nanoseconds()
	f()
	endTime := stdtime.Nanoseconds()
	return Time(float64(endTime-startTime) * Nanosecond)
}

// Time is a type that represents a duration of time.
type Time float64

// Split returns the components of the time in hours, minutes, and seconds.
func (t Time) Split() (hours, minutes int, seconds float64) {
	hours = int(t / Hour)
	minutes = int(math.Fmod(float64(t), Hour) / Minute)
	seconds = math.Fmod(float64(t), Minute)
	return
}

// String returns a human-readable representation of the time.
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
