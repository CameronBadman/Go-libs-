package skiplist

import (
	"math/rand"
	"testing"
)

func buildSequential(n int) *Skiplist {
	sl := New(20)
	for i := 0; i < n; i++ {
		if err := sl.InsertAt(sl.Len(), intData(i)); err != nil {
			panic(err)
		}
	}
	return sl
}

func BenchmarkSkiplistAtRandom(b *testing.B) {
	sizes := []int{1000, 10000, 100000}
	for _, size := range sizes {
		b.Run("N="+itoa(size), func(b *testing.B) {
			sl := buildSequential(size)
			rng := rand.New(rand.NewSource(42))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				idx := rng.Intn(size)
				if _, ok := sl.At(idx); !ok {
					b.Fatalf("missing index %d", idx)
				}
			}
		})
	}
}

func BenchmarkSkiplistInsertAtMiddle(b *testing.B) {
	sizes := []int{1000, 10000, 100000}
	for _, size := range sizes {
		b.Run("N="+itoa(size), func(b *testing.B) {
			sl := buildSequential(size)
			idx := size / 2
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := sl.InsertAt(idx, intData(-1)); err != nil {
					b.Fatal(err)
				}
				if _, err := sl.DeleteAt(idx); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkSkiplistInsertAtHead(b *testing.B) {
	sizes := []int{1000, 10000, 100000}
	for _, size := range sizes {
		b.Run("N="+itoa(size), func(b *testing.B) {
			sl := buildSequential(size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := sl.InsertAt(0, intData(-1)); err != nil {
					b.Fatal(err)
				}
				if _, err := sl.DeleteAt(0); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkSkiplistInsertAtTail(b *testing.B) {
	sizes := []int{1000, 10000, 100000}
	for _, size := range sizes {
		b.Run("N="+itoa(size), func(b *testing.B) {
			sl := buildSequential(size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := sl.InsertAt(sl.Len(), intData(-1)); err != nil {
					b.Fatal(err)
				}
				if _, err := sl.DeleteAt(sl.Len() - 1); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkSkiplistInsertAtRandom(b *testing.B) {
	sizes := []int{1000, 10000, 100000}
	for _, size := range sizes {
		b.Run("N="+itoa(size), func(b *testing.B) {
			sl := buildSequential(size)
			rng := rand.New(rand.NewSource(42))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				idx := rng.Intn(sl.Len() + 1)
				if err := sl.InsertAt(idx, intData(-1)); err != nil {
					b.Fatal(err)
				}
				if _, err := sl.DeleteAt(idx); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkSkiplistDeleteAtMiddle(b *testing.B) {
	sizes := []int{1000, 10000, 100000}
	for _, size := range sizes {
		b.Run("N="+itoa(size), func(b *testing.B) {
			sl := buildSequential(size)
			idx := size / 2
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				v, err := sl.DeleteAt(idx)
				if err != nil {
					b.Fatal(err)
				}
				if err := sl.InsertAt(idx, v); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkSkiplistDeleteAtHead(b *testing.B) {
	sizes := []int{1000, 10000, 100000}
	for _, size := range sizes {
		b.Run("N="+itoa(size), func(b *testing.B) {
			sl := buildSequential(size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				v, err := sl.DeleteAt(0)
				if err != nil {
					b.Fatal(err)
				}
				if err := sl.InsertAt(0, v); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkSkiplistDeleteAtTail(b *testing.B) {
	sizes := []int{1000, 10000, 100000}
	for _, size := range sizes {
		b.Run("N="+itoa(size), func(b *testing.B) {
			sl := buildSequential(size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				idx := sl.Len() - 1
				v, err := sl.DeleteAt(idx)
				if err != nil {
					b.Fatal(err)
				}
				if err := sl.InsertAt(sl.Len(), v); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkSkiplistDeleteAtRandom(b *testing.B) {
	sizes := []int{1000, 10000, 100000}
	for _, size := range sizes {
		b.Run("N="+itoa(size), func(b *testing.B) {
			sl := buildSequential(size)
			rng := rand.New(rand.NewSource(42))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				idx := rng.Intn(sl.Len())
				v, err := sl.DeleteAt(idx)
				if err != nil {
					b.Fatal(err)
				}
				if err := sl.InsertAt(idx, v); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}

	neg := v < 0
	if neg {
		v = -v
	}

	buf := [20]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + (v % 10))
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
