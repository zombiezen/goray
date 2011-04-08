package render

import "testing"
import "goray/core/color"

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
