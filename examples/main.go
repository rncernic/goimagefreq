package main

import (
	"fmt"
	freq "goimagefreq"
	"image/png"
	"os"
)

func main() {
	f, _ := os.Open("images/baboon.png")
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
}
