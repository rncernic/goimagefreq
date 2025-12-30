// Validation and debuging
package goimagefreq

import (
	"image/png"
	"math"
	"os"
	"sync"
)

// MaxAbsError computes the maximum absolute
// pixel-wise difference between two images.
func MaxAbsError(a, b [][]float32) float32 {
	h := len(a)
	w := len(a[0])
	maxErr := float32(0)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			e := float32(math.Abs(float64(a[y][x] - b[y][x])))
			if e > maxErr {
				maxErr = e
			}
		}
	}
	return maxErr
}

// DiffImage returns the absolute difference image
// for visual inspection.
func DiffImage(a, b [][]float32) [][]float32 {
	h := len(a)
	w := len(a[0])
	out := make([][]float32, h)
	for y := 0; y < h; y++ {
		out[y] = make([]float32, w)
		for x := 0; x < w; x++ {
			out[y][x] = float32(math.Abs(float64(a[y][x] - b[y][x])))
		}
	}
	return out
}

// SaveF32PNG saves a float32 image as an 8-bit PNG
// after linear normalization (debug/visualization only).
func SaveF32PNG(path string, img [][]float32) error {
	out := F32ToGray(img)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, out)
}

// SaveF32PNGRGB saves three float32 channels as a PNG to path.
func SaveF32PNGRGB(path string, rImg, gImg, bImg [][]float32) error {
	out := F32ToRGB(rImg, gImg, bImg)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, out)
}

// MaxAbsErrorRGB returns the maximum absolute error per channel (as triple).
func MaxAbsErrorRGB(aR, aG, aB, bR, bG, bB [][]float32) (errR, errG, errB float32) {
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		errR = MaxAbsError(aR, bR)
	}()
	go func() {
		defer wg.Done()
		errG = MaxAbsError(aG, bG)
	}()
	go func() {
		defer wg.Done()
		errB = MaxAbsError(aB, bB)
	}()
	return
}

// DiffImageRGB returns absolute-difference images per channel.
func DiffImageRGB(aR, aG, aB, bR, bG, bB [][]float32) (dR, dG, dB [][]float32) {
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		dR = DiffImage(aR, bR)
	}()
	go func() {
		defer wg.Done()
		dG = DiffImage(aG, bG)
	}()
	go func() {
		defer wg.Done()
		dB = DiffImage(aB, bB)
	}()
	return
}
