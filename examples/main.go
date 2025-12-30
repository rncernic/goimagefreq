package main

import (
	"fmt"
	freq "goimagefreq"
	"image/png"
	"math"
	"os"
)

func main() {
	// f, _ := os.Open("images/baboon.png")
	f, _ := os.Open("images/t1.png")

	img, _ := png.Decode(f)

	// gray image
	src := freq.ToGrayF32(img)

	// color image
	r, g, b := freq.ToRGBF32(img)

	// =============================
	// 1. LOW/HIGH TESTFREQUENCY
	// =============================
	low, high := freq.SplitLowHigh(src, 2.0)
	freq.SaveF32PNG("output/low.png", low)
	freq.SaveF32PNG("output/high.png", high)

	recLH := freq.ReconstructLowHigh(low, high)
	errLH := freq.MaxAbsError(src, recLH)
	freq.SaveF32PNG("output/lh_reconstructed.png", recLH)
	fmt.Println("Low/High max error:", errLH)

	diffLH := freq.DiffImage(src, recLH)
	freq.SaveF32PNG("output/diff_lowhigh.png", diffLH)

	// =============================================================
	// 2. À TROUS WAVELET (4 LEVELS)
	// =============================================================
	details, residual := freq.AtrousWavelet(src, 4)

	for i, d := range details {
		fname := fmt.Sprintf("output/atrous_level_%d.png", i)
		freq.SaveF32PNG(fname, d)
	}
	freq.SaveF32PNG("output/atrous_residual.png", residual)

	recAtrous := freq.AtrousReconstruct(details, residual)
	errAtrous := freq.MaxAbsError(src, recAtrous)
	freq.SaveF32PNG("output/atrous_reconstructed.png", recAtrous)
	fmt.Println("Atrous max error:", errAtrous)

	diffAtrous := freq.DiffImage(src, recAtrous)
	freq.SaveF32PNG("output/diff_atrous.png", diffAtrous)

	// ... color image
	rAtr, rResAtr := freq.AtrousWavelet(r, 4)
	gAtr, gResAtr := freq.AtrousWavelet(g, 4)
	bAtr, bResAtr := freq.AtrousWavelet(b, 4)

	rAtrRec := freq.AtrousReconstruct(rAtr, rResAtr)
	gAtrRec := freq.AtrousReconstruct(gAtr, gResAtr)
	bAtrRec := freq.AtrousReconstruct(bAtr, bResAtr)

	_ = freq.SaveF32PNGRGB("output/atrous_color_reconstructed.png", rAtrRec, gAtrRec, bAtrRec)

	// =============================================================
	// 3. MULTI-BAND GAUSSIAN PYRAMID (5 LEVELS)
	// =============================================================

	// ... gray image
	bands, resid := freq.MultiBand(src, 5, 1.5)

	for i, b := range bands {
		fname := fmt.Sprintf("output/band_%d.png", i)
		freq.SaveF32PNG(fname, b)
	}
	freq.SaveF32PNG("output/band_residual.png", resid)

	recMB := freq.MultiBandReconstruct(bands, resid)
	errMB := freq.MaxAbsError(src, recMB)
	freq.SaveF32PNG("output/mb_gray_reconstructed.png", recMB)
	fmt.Println("Multiband max error:", errMB)

	diffMB := freq.DiffImage(src, recMB)
	freq.SaveF32PNG("output/diff_multiband.png", diffMB)

	// ... color image
	rBands, rres := freq.MultiBand(r, 4, 1.0)
	gBands, gres := freq.MultiBand(g, 4, 1.0)
	bBands, bres := freq.MultiBand(b, 4, 1.0)

	rBandsRec := freq.MultiBandReconstruct(rBands, rres)
	gBandsRec := freq.MultiBandReconstruct(gBands, gres)
	bBandsRec := freq.MultiBandReconstruct(bBands, bres)

	_ = freq.SaveF32PNGRGB("output/mb_color_reconstructed.png", rBandsRec, gBandsRec, bBandsRec)

	// =============================================================
	// 4. L*-ONLY GAUSSIAN BLUR (COLOR SAFE)
	// =============================================================

	// Convert RGB → Lab
	L, a, b2 := freq.RGBToLabImage(r, g, b)
	// Blur ONLY luminance
	Lblur := freq.GaussianBlur(L, 1.5)
	// Back to RGB
	rBlur, gBlur, bBlur := freq.LabToRGBImage(Lblur, a, b2)
	// Save
	_ = freq.SaveF32PNGRGB("output/lab_luminance_blur.png", rBlur, gBlur, bBlur)

	// =============================================================
	// 5. L*-ONLY WAVELET DENOISE (À TROUS)
	// =============================================================

	L, a, b2 = freq.RGBToLabImage(r, g, b)
	// Decompose
	details, res := freq.AtrousWavelet(L, 5)
	// Soft-threshold fine scales
	for i := 0; i < 3; i++ { // fine layers
		for y := range details[i] {
			for x := range details[i][y] {
				if math.Abs(float64(details[i][y][x])) < 0.01 {
					details[i][y][x] = 0
				}
			}
		}
	}
	// Reconstruct
	Lden := freq.AtrousReconstruct(details, res)
	// Back to RGB
	rDen, gDen, bDen := freq.LabToRGBImage(Lden, a, b2)
	_ = freq.SaveF32PNGRGB("output/lab_wavelet_denoise.png", rDen, gDen, bDen)

	// =============================================================
	// 6. PIXINSIGHT-STYLE MLT DENOISE (L* ONLY)
	// =============================================================

	L, a, b2 = freq.RGBToLabImage(r, g, b)

	// Mild denoise (DSO broadband)
	// sigma := []float32{0.015, 0.010, 0.005, 0}

	// Strong denoise (narrowband)
	// sigma := []float32{0.03, 0.02, 0.01, 0.005}

	// Per-layer noise thresholds (fine → coarse)
	sigma := []float32{
		0.015, // layer 1
		0.010, // layer 2
		0.005, // layer 3
		0.000, // layer 4+
	}

	Lmlt := freq.WaveletDenoiseMLT(L, sigma)

	// Back to RGB
	rMLT, gMLT, bMLT := freq.LabToRGBImage(Lmlt, a, b2)

	_ = freq.SaveF32PNGRGB("output/mlt_luminance_denoise.png", rMLT, gMLT, bMLT)

	// =============================================================
	// 7. COLOR-SAFE RICHARDSON–LUCY (L* ONLY)
	// =============================================================

	// Convert to Lab
	L, a, b2 = freq.RGBToLabImage(r, g, b)

	// Estimate PSF from stars
	kx, ky := freq.EstimatePSF(L, 0.01)

	// Deconvolve luminance only
	Ldeconv := freq.RichardsonLucy(L, kx, ky, 25)

	// Optional: mild post-blur to control ringing
	Ldeconv = freq.GaussianBlur(Ldeconv, 0.5)

	// Back to RGB
	rRL, gRL, bRL := freq.LabToRGBImage(Ldeconv, a, b2)

	_ = freq.SaveF32PNGRGB("output/rl_luminance_deconv.png", rRL, gRL, bRL)

	// =============================================================
	// 8. SWT (L* ONLY)
	// =============================================================

	// Convert to Lab
	L, a, b = freq.RGBToLabImage(r, g, b)

	// Denoise
	Lden = freq.SWTDenoise(
		L,
		[]float32{3.0, 2.0, 1.0, 0.5, 0.0},
		true, // soft threshold
	)

	// Back to rgb
	rOut, gOut, bOut := freq.LabToRGBImage(Lden, a, b)

	_ = freq.SaveF32PNGRGB("output/swt_denoise.png", rOut, gOut, bOut)

}
