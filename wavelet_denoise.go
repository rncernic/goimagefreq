// MAD-based wavelet denoising
package goimagefreq

import (
	"math"
	"math/rand"
)

// Median Absolute Deviation (robust sigma estimator)
func MAD(img [][]float32) float64 {
	var vals []float64
	for y := range img {
		for x := range img[y] {
			vals = append(vals, float64(img[y][x]))
		}
	}
	median := quickMedian(vals)
	for i := range vals {
		vals[i] = math.Abs(vals[i] - median)
	}
	return quickMedian(vals)
}

// Soft-threshold shrinkage
func softThreshold(v, t float32) float32 {
	if math.Abs(float64(v)) <= float64(t) {
		return 0
	}
	if v > 0 {
		return float32(v - t)
	}
	return float32(v + t)
}

// AtrousWaveletDenoiseL applies wavelet denoising to luminance only.
func AtrousWaveletDenoiseL(L [][]float32, levels int, strength []float32) [][]float32 {
	details, residual := AtrousWavelet(L, levels)

	for i := 0; i < levels; i++ {
		sigma := MAD(details[i]) / 0.6745
		thr := strength[i] * float32(sigma)

		for y := range details[i] {
			for x := range details[i][y] {
				details[i][y][x] =
					softThreshold(details[i][y][x], thr)
			}
		}
	}

	return AtrousReconstruct(details, residual)
}

// quickMedian returns the median value of the slice.
// The input slice is modified.
func quickMedian(a []float64) float64 {
	n := len(a)
	if n == 0 {
		return 0
	}
	k := n / 2

	lo, hi := 0, n-1
	for {
		if lo == hi {
			return a[lo]
		}
		p := partition(a, lo, hi)
		if k == p {
			return a[k]
		} else if k < p {
			hi = p - 1
		} else {
			lo = p + 1
		}
	}
}

func partition(a []float64, lo, hi int) int {
	pivot := a[lo+rand.Intn(hi-lo+1)]
	i, j := lo, hi

	for {
		for a[i] < pivot {
			i++
		}
		for a[j] > pivot {
			j--
		}
		if i >= j {
			return j
		}
		a[i], a[j] = a[j], a[i]
		i++
		j--
	}
}

func WaveletDenoiseMLT(
	L [][]float32,
	sigma []float32,
) [][]float32 {

	levels := len(sigma)

	// À trous decomposition
	details, residual := AtrousWavelet(L, levels)

	// Threshold each detail layer
	for i := 0; i < len(details); i++ {
		t := sigma[min(i, len(sigma)-1)]

		if t <= 0 {
			continue
		}

		for y := range details[i] {
			for x := range details[i][y] {
				details[i][y][x] = softThreshold(details[i][y][x], t)
			}
		}
	}

	// Reconstruct
	return AtrousReconstruct(details, residual)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
