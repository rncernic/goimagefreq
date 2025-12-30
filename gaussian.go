// Gaussian kernels
package goimagefreq

import "math"

// GaussianKernel generates a 1D normalized Gaussian kernel.
//
// Radius is chosen as 3*sigma, which captures >99%
// of the Gaussian energy.
func GaussianKernel(sigma float64) []float64 {
	r := int(math.Ceil(3 * sigma))
	size := 2*r + 1
	k := make([]float64, size)
	total := 0.0
	for i := -r; i <= r; i++ {
		v := math.Exp(-float64(i*i) / (2 * sigma * sigma))
		k[i+r] = v
		total += v
	}
	// Normalize kernel so sum = 1
	for i := range k {
		k[i] /= total
	}
	return k
}
