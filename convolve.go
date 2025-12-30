// Parallel separable convolution
package goimagefreq

import (
	"runtime"
	"sync"
)

// Convolve1D performs separable 1D convolution
// either horizontally or vertically.
//
// Edge handling: clamp-to-edge (replication).
//
// This is the building block for Gaussian blur,
// wavelets, and multiband decomposition.
func Convolve1D(src [][]float32, kernel []float64, horizontal bool) [][]float32 {
	h := len(src)
	w := len(src[0])
	r := len(kernel) / 2

	out := make([][]float32, h)
	for y := range out {
		out[y] = make([]float32, w)
	}

	workers := runtime.GOMAXPROCS(0)
	jobs := make(chan int, h)
	wg := sync.WaitGroup{}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for y := range jobs {
				if horizontal {
					for x := 0; x < w; x++ {
						acc := float64(0)
						for k := -r; k <= r; k++ {
							xx := x + k
							if xx < 0 {
								xx = 0
							}
							if xx >= w {
								xx = w - 1
							}
							acc += float64(src[y][xx]) * kernel[k+r]
						}
						out[y][x] = float32(acc)
					}
				} else {
					for x := 0; x < w; x++ {
						acc := float64(0)
						for k := -r; k <= r; k++ {
							yy := y + k
							if yy < 0 {
								yy = 0
							}
							if yy >= h {
								yy = h - 1
							}
							acc += float64(src[yy][x]) * kernel[k+r]
						}
						out[y][x] = float32(acc)
					}
				}
			}
		}()
	}

	for y := 0; y < h; y++ {
		jobs <- y
	}
	close(jobs)
	wg.Wait()

	return out
}

func Convolve2DGeneric(src [][]float32, kernel [][]float32) [][]float32 {
	h := len(src)
	w := len(src[0])
	kh := len(kernel)
	kw := len(kernel[0])

	ry := kh / 2
	rx := kw / 2

	out := make([][]float32, h)
	for y := range out {
		out[y] = make([]float32, w)
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
					acc := float32(0)
					for ky := -ry; ky <= ry; ky++ {
						yy := y + ky
						if yy < 0 {
							yy = 0
						}
						if yy >= h {
							yy = h - 1
						}
						for kx := -rx; kx <= rx; kx++ {
							xx := x + kx
							if xx < 0 {
								xx = 0
							}
							if xx >= w {
								xx = w - 1
							}
							acc += src[yy][xx] *
								kernel[ky+ry][kx+rx]
						}
					}
					out[y][x] = acc
				}
			}
		}()
	}

	for y := 0; y < h; y++ {
		jobs <- y
	}
	close(jobs)
	wg.Wait()

	return out
}

// Convolve2DSeparable performs a full 2D convolution using
// two 1D convolutions (horizontal then vertical).
//
// This is mathematically equivalent to 2D convolution
// ONLY if the kernel is separable:
//
//	K(x, y) = ky(y) * kx(x)
//
// This is the preferred method for Gaussian / PSF convolution.
func Convolve2DSeparable(
	src [][]float32,
	kx []float64,
	ky []float64,
) [][]float32 {

	// Horizontal pass
	tmp := Convolve1D(src, kx, true)

	// Vertical pass
	return Convolve1D(tmp, ky, false)
}
