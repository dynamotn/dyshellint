package lint

import (
	"os"
	"runtime"
	"sort"
	"sync"
)

// Parallel calls fn once for every index below n, over a pool of jobs
// workers, the number of CPUs when jobs is zero or less. The indexes are handed
// out in the order given by order, or in increasing order when order is nil.
func Parallel(n, jobs int, order []int, fn func(i int)) {
	if jobs <= 0 {
		jobs = runtime.NumCPU()
	}
	next := make(chan int)
	var wg sync.WaitGroup
	for range min(jobs, n) {
		wg.Go(func() {
			for i := range next {
				fn(i)
			}
		})
	}
	for k := range n {
		if order != nil {
			next <- order[k]
		} else {
			next <- k
		}
	}
	close(next)
	wg.Wait()
}

// eachFile runs check on every file over a pool of jobs workers and gathers
// the findings. The largest files go first: the run lasts as long as its
// slowest file, so that one has to start before the pool fills with small
// ones. The first error, in the order of files, wins.
func eachFile(files []string, jobs int, check func(file string) ([]Finding, error)) ([]Finding, error) {
	results := make([][]Finding, len(files))
	errs := make([]error, len(files))
	Parallel(len(files), jobs, largestFirst(files), func(i int) {
		results[i], errs[i] = check(files[i])
	})
	var findings []Finding
	for i := range files {
		if errs[i] != nil {
			return nil, errs[i]
		}
		findings = append(findings, results[i]...)
	}
	return findings, nil
}

// largestFirst returns the indexes of files ordered by decreasing size, the
// cheap stand-in for how long each one takes to check.
func largestFirst(files []string) []int {
	sizes := make([]int64, len(files))
	order := make([]int, len(files))
	for i, file := range files {
		order[i] = i
		if info, err := os.Stat(file); err == nil {
			sizes[i] = info.Size()
		}
	}
	sort.SliceStable(order, func(a, b int) bool { return sizes[order[a]] > sizes[order[b]] })
	return order
}
