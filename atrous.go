// À trous wavelet
package goimagefreq

import "sync"

// Base B3-spline kernel used in à trous wavelets
var atrousKernel = []float64{1.0 / 16, 4.0 / 16, 6.0 / 16, 4.0 / 16, 1.0 / 16}

// AtrousDilateKernel inserts zeros between kernel taps
// according to the wavelet scale.
//
// Scale w increases the effective kernel size without
// downsampling the image.
func AtrousDilateKernel(w int) []float64 {
	// insert 2^(w-1) zeros between taps
	if w == 0 {
		return atrousKernel
	}

	step := 1 << (w - 1)
	out := []float64{}
	for i := 0; i < len(atrousKernel); i++ {
		out = append(out, atrousKernel[i])
		if i < len(atrousKernel)-1 {
			for z := 0; z < step; z++ {
				out = append(out, 0) // zeros
			}
		}
	}
	return out
}

// AtrousWavelet performs an undecimated wavelet transform.
//
// Returns:
//
//	details[level] = band-pass layer
//	residual       = final smooth image
func AtrousWavelet(src [][]float32, levels int) (details [][][]float32, residual [][]float32) {
	current := src
	for level := 0; level < levels; level++ {
		k := AtrousDilateKernel(level)
		tmp := Convolve1D(current, k, true)
		smooth := Convolve1D(tmp, k, false)

		// detail = current - smooth
		h := len(src)
		w := len(src[0])
		detail := make([][]float32, h)
		for y := 0; y < h; y++ {
			detail[y] = make([]float32, w)
			for x := 0; x < w; x++ {
				detail[y][x] = current[y][x] - smooth[y][x]
			}
		}
		details = append(details, detail)
		current = smooth
	}
	residual = current
	return
}

// AtrousReconstruct reconstructs the image
// by summing residual + all detail layers.
func AtrousReconstruct(details [][][]float32, residual [][]float32) [][]float32 {
	h := len(residual)
	w := len(residual[0])
	out := make([][]float32, h)
	for y := range out {
		out[y] = make([]float32, w)
		for x := 0; x < w; x++ {
			out[y][x] = residual[y][x]
		}
	}

	for _, d := range details {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				out[y][x] += d[y][x]
			}
		}
	}
	return out
}

// AtrousWaveletRGB runs AtrousWavelet on each channel and returns details as
// slices [levels][h][w] for each channel plus residuals.
func AtrousWaveletRGB(r, g, b [][]float32, levels int) (rDetails, gDetails, bDetails [][][]float32, rResid, gResid, bResid [][]float32) {
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		rDetails, rResid = AtrousWavelet(r, levels)
	}()
	go func() {
		defer wg.Done()
		gDetails, gResid = AtrousWavelet(g, levels)
	}()
	go func() {
		defer wg.Done()
		bDetails, bResid = AtrousWavelet(b, levels)
	}()
	return
}

// AtrousReconstructRGB reconstructs the RGB image from per-channel details+residual.
func AtrousReconstructRGB(rDetails, gDetails, bDetails [][][]float32, rResid, gResid, bResid [][]float32) (r, g, b [][]float32) {
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		r = AtrousReconstruct(rDetails, rResid)
	}()
	go func() {
		defer wg.Done()
		g = AtrousReconstruct(gDetails, gResid)
	}()
	go func() {
		defer wg.Done()
		b = AtrousReconstruct(bDetails, bResid)
	}()
	return
}
