// PixInsight-like MLT
package goimagefreq

// MLTParams defines PixInsight-style MLT controls
type MLTParams struct {
	Gain []float64 // per scale
	Bias []float64 // per scale (negative suppresses noise)
}

// ApplyMLTLuminance applies MLT to L* channel only
func ApplyMLTLuminance(L [][]float32, params MLTParams) [][]float32 {
	levels := len(params.Gain)
	details, residual := AtrousWavelet(L, levels)

	for i := 0; i < levels; i++ {
		for y := range details[i] {
			for x := range details[i][y] {
				v := float64(details[i][y][x])
				v = params.Bias[i] + params.Gain[i]*v
				details[i][y][x] = float32(v)
			}
		}
	}

	return AtrousReconstruct(details, residual)
}
