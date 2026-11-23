package main

import "testing"

// Еталонні значення π(n) - кількості простих чисел <= n.
var cases = []struct{ n, want int }{
	{10, 4}, {100, 25}, {1000, 168},
	{100_000, 9592}, {1_000_000, 78498},
}

func TestCounts(t *testing.T) {
	for _, c := range cases {
		if got := countSequential(c.n); got != c.want {
			t.Errorf("sequential(%d) = %d, want %d",
				c.n, got, c.want)
		}
		for _, w := range []int{1, 2, 3, 4, 8} {
			if got := countParallelStatic(c.n, w); got != c.want {
				t.Errorf("parallel(%d, %d) = %d, want %d",
					c.n, w, got, c.want)
			}
		}
	}
}

func BenchmarkSequential(b *testing.B) {
	for i := 0; i < b.N; i++ {
		countSequential(2_000_000)
	}
}

func BenchmarkParallel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		countParallelStatic(2_000_000, 8)
	}
}
