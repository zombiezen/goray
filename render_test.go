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

package goray

import (
	"testing"
	"bitbucket.org/zombiezen/goray/color"
)

func doAcquireBench(b *testing.B, queueSize int) {
	b.SetBytes(32) // sizeof(float64) * 4

	img := NewImage(b.N, 1)
	ch := make(chan Fragment, queueSize)
	go func() {
		defer close(ch)
		for i := 0; i < b.N; i++ {
			ch <- Fragment{color.RGBA{0.1, 0.2, 0.3, 0.5}, i, 0}
		}
	}()

	b.StartTimer()
	img.Acquire(ch)
}

func BenchmarkSyncAcquire(b *testing.B) {
	b.StopTimer()
	doAcquireBench(b, 0)
}

func BenchmarkAsyncAcquire(b *testing.B) {
	b.StopTimer()
	doAcquireBench(b, 100)
}
