// Blur functions
package goimagefreq

import (
	"runtime"
	"sync"
)

// GaussianBlur applies a full 2D Gaussian blur
// using separable convolution (horizontal + vertical).
func GaussianBlur(src [][]float32, sigma float64) [][]float32 {
	k := GaussianKernel(sigma)
	tmp := Convolve1D(src, k, true)
	return Convolve1D(tmp, k, false)
}

// GaussianBlurYCbCr blurs only luminance (Y channel).
func GaussianBlurYCbCr(img RGBImage, sigma float64) RGBImage {
	h, w := img.H, img.W

	Y := make([][]float32, h)
	Cb := make([][]float32, h)
	Cr := make([][]float32, h)

	for y := 0; y < h; y++ {
		Y[y] = make([]float32, w)
		Cb[y] = make([]float32, w)
		Cr[y] = make([]float32, w)
	}

	// RGB → YCbCr (parallel)
	wg := sync.WaitGroup{}
	workers := runtime.GOMAXPROCS(0)
	jobs := make(chan int, h)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for y := range jobs {
				for x := 0; x < w; x++ {
					Y[y][x], Cb[y][x], Cr[y][x] =
						RGBToYCbCr(img.R[y][x], img.G[y][x], img.B[y][x])
				}
			}
		}()
	}

	for y := 0; y < h; y++ {
		jobs <- y
	}
	close(jobs)
	wg.Wait()

	// Blur luminance only
	Yb := GaussianBlur(Y, sigma)

	// Recombine
	out := RGBImage{
		W: w,
		H: h,
		R: make([][]float32, h),
		G: make([][]float32, h),
		B: make([][]float32, h),
	}

	for y := 0; y < h; y++ {
		out.R[y] = make([]float32, w)
		out.G[y] = make([]float32, w)
		out.B[y] = make([]float32, w)
		for x := 0; x < w; x++ {
			out.R[y][x], out.G[y][x], out.B[y][x] =
				YCbCrToRGB(Yb[y][x], Cb[y][x], Cr[y][x])
		}
	}

	return out
}

// GaussianBlurLab blurs only the L* channel (perceptual luminance).
func GaussianBlurLab(img RGBImage, sigma float64) RGBImage {
	h, w := img.H, img.W

	L := make([][]float32, h)
	A := make([][]float32, h)
	B := make([][]float32, h)

	for y := 0; y < h; y++ {
		L[y] = make([]float32, w)
		A[y] = make([]float32, w)
		B[y] = make([]float32, w)
	}

	wg := sync.WaitGroup{}
	workers := runtime.GOMAXPROCS(0)
	jobs := make(chan int, h)

	// RGB → Lab
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for y := range jobs {
				for x := 0; x < w; x++ {
					L[y][x], A[y][x], B[y][x] =
						RGBToLab(img.R[y][x], img.G[y][x], img.B[y][x])
				}
			}
		}()
	}

	for y := 0; y < h; y++ {
		jobs <- y
	}
	close(jobs)
	wg.Wait()

	// Blur luminance only
	Lb := GaussianBlur(L, sigma)

	// Recombine
	out := RGBImage{
		W: w,
		H: h,
		R: make([][]float32, h),
		G: make([][]float32, h),
		B: make([][]float32, h),
	}

	for y := 0; y < h; y++ {
		out.R[y] = make([]float32, w)
		out.G[y] = make([]float32, w)
		out.B[y] = make([]float32, w)
		for x := 0; x < w; x++ {
			out.R[y][x], out.G[y][x], out.B[y][x] =
				LabToRGB(Lb[y][x], A[y][x], B[y][x])
		}
	}

	return out
}
