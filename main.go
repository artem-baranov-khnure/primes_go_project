// Програма підраховує кількість простих чисел у діапазоні [2, N]
// послідовно та паралельно (горутини + канали + sync.WaitGroup)
// і порівнює час виконання.
package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// isPrime перевіряє число на простоту методом пробного ділення.
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n%2 == 0 {
		return n == 2
	}
	for d := 3; d*d <= n; d += 2 {
		if n%d == 0 {
			return false
		}
	}
	return true
}

// countRange рахує прості числа в напівінтервалі [lo, hi).
func countRange(lo, hi int) int {
	count := 0
	for i := lo; i < hi; i++ {
		if isPrime(i) {
			count++
		}
	}
	return count
}

// countSequential - послідовна версія (еталон для порівняння).
func countSequential(n int) int {
	return countRange(2, n+1)
}

// countParallelStatic ділить діапазон на workers рівних частин,
// кожну частину обробляє окрема горутина,
// результати збираються через канал.
func countParallelStatic(n, workers int) int {
	results := make(chan int, workers)
	var wg sync.WaitGroup

	size := (n - 1) / workers // у діапазоні [2, n] є n-1 чисел
	lo := 2
	for w := 0; w < workers; w++ {
		hi := lo + size
		if w == workers-1 {
			hi = n + 1 // останній блок забирає залишок
		}
		wg.Add(1)
		go func(lo, hi int) {
			defer wg.Done()
			results <- countRange(lo, hi)
		}(lo, hi)
		lo = hi
	}

	// Канал закривається, коли всі горутини завершаться.
	go func() {
		wg.Wait()
		close(results)
	}()

	total := 0
	for c := range results {
		total += c
	}
	return total
}

// measure виконує f runs разів і повертає мінімальний час.
func measure(runs int, f func() int) (int, time.Duration) {
	best := time.Duration(1<<63 - 1)
	res := 0
	for r := 0; r < runs; r++ {
		start := time.Now()
		res = f()
		if el := time.Since(start); el < best {
			best = el
		}
	}
	return res, best
}

func main() {
	n := flag.Int("n", 10_000_000, "верхня межа діапазону [2, n]")
	runs := flag.Int("runs", 3, "кількість повторів (мінімум часу)")
	flag.Parse()

	fmt.Printf("CPU: %d логічних ядер, GOMAXPROCS = %d\n",
		runtime.NumCPU(), runtime.GOMAXPROCS(0))
	fmt.Printf("N = %d, повторів = %d\n\n", *n, *runs)

	seqRes, seqTime := measure(*runs, func() int {
		return countSequential(*n)
	})
	fmt.Printf("%-20s %7s %10s %9s %10s\n",
		"Варіант", "Горутин", "Час, мс", "Прискор.", "Результат")
	printRow("Послідовний", 1, seqTime, seqTime, seqRes)

	workerCounts := []int{2, 4, 8, 16}
	if c := runtime.NumCPU(); c > 16 {
		workerCounts = append(workerCounts, c)
	}

	for _, w := range workerCounts {
		res, t := measure(*runs, func() int {
			return countParallelStatic(*n, w)
		})
		check(res, seqRes)
		printRow("Паралельний", w, t, seqTime, res)
	}
}

// printRow друкує рядок таблиці результатів.
func printRow(name string, g int, t, seq time.Duration, res int) {
	ms := float64(t.Microseconds()) / 1000
	speedup := float64(seq) / float64(t)
	fmt.Printf("%-20s %7d %10.1f %9.2f %10d\n",
		name, g, ms, speedup, res)
}

// check перевіряє збіг паралельного результату з послідовним.
func check(got, want int) {
	if got != want {
		panic(fmt.Sprintf("помилка: %d != %d", got, want))
	}
}
