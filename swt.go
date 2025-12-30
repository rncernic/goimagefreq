// Stationary wavelet transform
package goimagefreq

import "math"

type SWTLayer struct {
	Detail [][]float32
	Sigma  float32
}

type SWTResult struct {
	Layers   []SWTLayer
	Residual [][]float32
}

var b3Spline = []float64{
	1.0 / 16,
	4.0 / 16,
	6.0 / 16,
	4.0 / 16,
	1.0 / 16,
}

func swtKernel(level int) []float64 {
	if level == 0 {
		return b3Spline
	}
	step := 1 << level
	var k []float64
	for i := 0; i < len(b3Spline); i++ {
		k = append(k, b3Spline[i])
		if i < len(b3Spline)-1 {
			for z := 0; z < step; z++ {
				k = append(k, 0)
			}
		}
	}
	return k
}

func SWTDecompose(
	src [][]float32,
	levels int,
) SWTResult {

	current := src
	var layers []SWTLayer

	for i := 0; i < levels; i++ {
		k := swtKernel(i)

		tmp := Convolve1D(current, k, true)
		smooth := Convolve1D(tmp, k, false)

		h := len(src)
		w := len(src[0])

		detail := make([][]float32, h)
		for y := 0; y < h; y++ {
			detail[y] = make([]float32, w)
			for x := 0; x < w; x++ {
				detail[y][x] = current[y][x] - smooth[y][x]
			}
		}

		layers = append(layers, SWTLayer{
			Detail: detail,
		})

		current = smooth
	}

	return SWTResult{
		Layers:   layers,
		Residual: current,
	}
}

func EstimateNoiseMAD(layer [][]float32) float32 {

	var values []float64
	for y := range layer {
		for x := range layer[y] {
			values = append(values, float64(layer[y][x]))
		}
	}

	med := quickMedian(values)

	for i := range values {
		values[i] = math.Abs(values[i] - med)
	}

	mad := quickMedian(values)

	// Gaussian equivalent
	return float32(1.4826 * mad)
}

func shrink(v, t float32, soft bool) float32 {
	av := float32(math.Abs(float64(v)))
	if av < t {
		return 0
	}
	if soft {
		return float32(math.Copysign(float64(av-t), float64(v)))
	}
	return v
}

func SWTDenoise(
	src [][]float32,
	sigmas []float32,
	soft bool,
) [][]float32 {

	res := SWTDecompose(src, len(sigmas))

	for i := range res.Layers {
		sigma := sigmas[i]
		if sigma <= 0 {
			continue
		}

		layer := res.Layers[i].Detail
		noise := EstimateNoiseMAD(layer)
		th := sigma * noise

		for y := range layer {
			for x := range layer[y] {
				layer[y][x] = shrink(layer[y][x], th, soft)
			}
		}
	}

	// reconstruct
	h := len(src)
	w := len(src[0])
	out := make([][]float32, h)

	for y := 0; y < h; y++ {
		out[y] = make([]float32, w)
		for x := 0; x < w; x++ {
			out[y][x] = res.Residual[y][x]
		}
	}

	for _, l := range res.Layers {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				out[y][x] += l.Detail[y][x]
			}
		}
	}

	return out
}
