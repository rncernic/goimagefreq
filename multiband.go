// Multiband frequency decomposition
package goimagefreq

import (
	"math"
	"sync"
)

// MultiBand performs a Gaussian pyramid-like decomposition
// using increasing sigma values.
//
// Each band captures a frequency range.
func MultiBand(src [][]float32, levels int, sigma0 float64) (bands [][][]float32, residual [][]float32) {
	current := src
	for i := 0; i < levels; i++ {
		sigma := sigma0 * math.Pow(2, float64(i)) // geometric growth
		low := GaussianBlur(current, sigma)
		h := len(src)
		w := len(src[0])
		band := make([][]float32, h)
		for y := 0; y < h; y++ {
			band[y] = make([]float32, w)
			for x := 0; x < w; x++ {
				band[y][x] = current[y][x] - low[y][x]
			}
		}
		bands = append(bands, band)
		current = low
	}
	residual = current
	return
}

// MultiBandReconstruct reconstrucs the image by
// summing all bands plus residual.
func MultiBandReconstruct(bands [][][]float32, residual [][]float32) [][]float32 {
	h := len(residual)
	w := len(residual[0])
	out := make([][]float32, h)
	for y := range out {
		out[y] = make([]float32, w)
		for x := range out[y] {
			out[y][x] = residual[y][x]
		}
	}
	for _, b := range bands {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				out[y][x] += b[y][x]
			}
		}
	}
	return out
}

// MultiBandRGB splits into multiband per channel.
func MultiBandRGB(r, g, b [][]float32, levels int, sigma0 float64) (rBands, gBands, bBands [][][]float32, rResid, gResid, bResid [][]float32) {
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		rBands, rResid = MultiBand(r, levels, sigma0)
	}()
	go func() {
		defer wg.Done()
		gBands, gResid = MultiBand(g, levels, sigma0)
	}()
	go func() {
		defer wg.Done()
		bBands, bResid = MultiBand(b, levels, sigma0)
	}()
	return
}

// MultiBandReconstructRGB reconstructs per channel.
func MultiBandReconstructRGB(rBands, gBands, bBands [][][]float32, rResid, gResid, bResid [][]float32) (r, g, b [][]float32) {
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		r = MultiBandReconstruct(rBands, rResid)
	}()
	go func() {
		defer wg.Done()
		g = MultiBandReconstruct(gBands, gResid)
	}()
	go func() {
		defer wg.Done()
		b = MultiBandReconstruct(bBands, bResid)
	}()
	return
}
