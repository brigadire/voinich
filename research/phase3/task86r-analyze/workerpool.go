package main

import (
	"runtime"
	"sync"
)

// parallelFor runs fn(i) for i in [0,n) using a bounded worker pool. Results
// must be written by each fn call into a caller-owned, index-addressed
// slice (never a shared map) so output order is independent of completion
// order -- required for byte-identical regeneration across GOMAXPROCS and
// worker-count variation (G1_SEED_CONTRACT.md).
func parallelFor(n int, fn func(i int)) {
	workers := runtime.GOMAXPROCS(0)
	if workers < 1 {
		workers = 1
	}
	if workers > n {
		workers = n
	}
	if workers <= 1 {
		for i := 0; i < n; i++ {
			fn(i)
		}
		return
	}
	var wg sync.WaitGroup
	next := make(chan int)
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := range next {
				fn(i)
			}
		}()
	}
	for i := 0; i < n; i++ {
		next <- i
	}
	close(next)
	wg.Wait()
}
