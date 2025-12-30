// Deconvolution
package goimagefreq

func RichardsonLucy(
	L [][]float32,
	kx []float64,
	ky []float64,
	iterations int,
) [][]float32 {

	h := len(L)
	w := len(L[0])

	// Initial estimate = observed image
	estimate := make([][]float32, h)
	for y := range estimate {
		estimate[y] = make([]float32, w)
		copy(estimate[y], L[y])
	}

	// Flipped PSF (adjoint operator)
	kxFlip := flipKernel1D(kx)
	kyFlip := flipKernel1D(ky)

	const eps = 1e-6

	for it := 0; it < iterations; it++ {

		// Blur current estimate
		blur := Convolve2DSeparable(estimate, kx, ky)

		// Ratio image
		ratio := make([][]float32, h)
		for y := 0; y < h; y++ {
			ratio[y] = make([]float32, w)
			for x := 0; x < w; x++ {
				if blur[y][x] > eps {
					ratio[y][x] = L[y][x] / blur[y][x]
				} else {
					ratio[y][x] = 0
				}
			}
		}

		// Back-project correction
		corr := Convolve2DSeparable(ratio, kxFlip, kyFlip)

		// Update estimate
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				estimate[y][x] *= corr[y][x]
			}
		}
	}

	return estimate
}

// func flipKernel(k [][]float32) [][]float32 {
// 	h := len(k)
// 	w := len(k[0])

// 	out := make([][]float32, h)
// 	for y := 0; y < h; y++ {
// 		out[y] = make([]float32, w)
// 		for x := 0; x < w; x++ {
// 			out[y][x] = k[h-1-y][w-1-x]
// 		}
// 	}
// 	return out
// }

func flipKernel1D(k []float64) []float64 {
	out := make([]float64, len(k))
	for i := range k {
		out[i] = k[len(k)-1-i]
	}
	return out
}
