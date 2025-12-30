// Low / high frequency split
package goimagefreq

import (
	"runtime"
	"sync"
)

// SplitLowHigh decomposes an image into:
//
//	low  = GaussianBlur(src)
//	high = src - low
func SplitLowHigh(src [][]float32, sigma float64) (low, high [][]float32) {
	low = GaussianBlur(src, sigma)

	h := len(src)
	w := len(src[0])

	high = make([][]float32, h)
	for y := range high {
		high[y] = make([]float32, w)
	}

	workers := runtime.GOMAXPROCS(0)
	jobs := make(chan int, h)
	wg := sync.WaitGroup{}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for y := range jobs {
				for x := 0; x < w; x++ {
					high[y][x] = src[y][x] - low[y][x]
				}
			}
		}()
	}

	for y := 0; y < h; y++ {
		jobs <- y
	}
	close(jobs)
	wg.Wait()

	return
}

// ReconstructLowHigh perfectly reconstructs the image
// by summing low + high components.
func ReconstructLowHigh(low, high [][]float32) [][]float32 {
	h := len(low)
	w := len(low[0])
	out := make([][]float32, h)
	for y := 0; y < h; y++ {
		out[y] = make([]float32, w)
		for x := 0; x < w; x++ {
			out[y][x] = low[y][x] + high[y][x]
		}
	}
	return out
}

// SplitLowHighRGB applies SplitLowHigh per channel.
func SplitLowHighRGB(r, g, b [][]float32, sigma float64) (rLow, rHigh, gLow, gHigh, bLow, bHigh [][]float32) {
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		rLow, rHigh = SplitLowHigh(r, sigma)
	}()
	go func() {
		defer wg.Done()
		gLow, gHigh = SplitLowHigh(g, sigma)
	}()
	go func() {
		defer wg.Done()
		bLow, bHigh = SplitLowHigh(b, sigma)
	}()

	wg.Wait()
	return
}

// ReconstructLowHighRGB reconstructs per channel.
func ReconstructLowHighRGB(rLow, rHigh, gLow, gHigh, bLow, bHigh [][]float32) (r, g, b [][]float32) {
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		r = ReconstructLowHigh(rLow, rHigh)
	}()
	go func() {
		defer wg.Done()
		g = ReconstructLowHigh(gLow, gHigh)
	}()
	go func() {
		defer wg.Done()
		b = ReconstructLowHigh(bLow, bHigh)
	}()
	return
}
