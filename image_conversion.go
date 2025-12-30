// Image conversions
package goimagefreq

import (
	"image"
	"image/color"
	"sync"
)

type RGBImage struct {
	W, H    int
	R, G, B [][]float32
}

// ToGrayF32 converts a generic image.Image into a
// 2D float32 grayscale buffer.
//
// Output values are normalized to [0,1] using
// ITU-R BT.709 luminance coefficients:
//
//	Y = 0.2126 R + 0.7152 G + 0.0722 B
//
// This preserves perceived brightness and is suitable
// for frequency-domain operations.
func ToGrayF32(img image.Image) [][]float32 {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := make([][]float32, h)
	for y := 0; y < h; y++ {
		out[y] = make([]float32, w)
		for x := 0; x < w; x++ {
			r, g, bb, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			out[y][x] = float32(0.2126*float64(r)/65535.0 +
				0.7152*float64(g)/65535.0 +
				0.0722*float64(bb)/65535.0)
		}
	}
	return out
}

// F32ToGray converts a float32 grayscale matrix into
// an 8-bit image.Gray for visualization.
//
// The data is linearly normalized to [0,255].
// This is intended ONLY for display/debugging,
// not for scientific output.
func F32ToGray(img [][]float32) *image.Gray {
	h := len(img)
	w := 0
	if h > 0 {
		w = len(img[0])
	}
	out := image.NewGray(image.Rect(0, 0, w, h))

	// Find min/max for linear normalization
	minv, maxv := float32(1e9), float32(-1e9)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if img[y][x] < minv {
				minv = img[y][x]
			}
			if img[y][x] > maxv {
				maxv = img[y][x]
			}
		}
	}

	// Avoid division by zero
	scale := float32(1.0)
	if maxv != minv {
		scale = float32(1.0) / (maxv - minv)
	} else {
		scale = 0
	}

	// Normalize and convert to uint8
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := uint8(255 * (img[y][x] - minv) * scale)
			out.SetGray(x, y, color.Gray{Y: v})
		}
	}
	return out
}

// ToRGBF32 converts image.Image to three float32 matrices (R, G, B) in range [0,1].
// It respects arbitrary image bounds.
func ToRGBF32(img image.Image) (rOut, gOut, bOut [][]float32) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	rOut = make([][]float32, h)
	gOut = make([][]float32, h)
	bOut = make([][]float32, h)

	var wg sync.WaitGroup
	wg.Add(h)

	for y := 0; y < h; y++ {
		y := y
		go func() {
			defer wg.Done()
			rRow := make([]float32, w)
			gRow := make([]float32, w)
			bRow := make([]float32, w)

			for x := 0; x < w; x++ {
				rr, gg, bb, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
				// Normalize from 0..65535 to 0..1
				rRow[x] = float32(rr) / 65535.0
				gRow[x] = float32(gg) / 65535.0
				bRow[x] = float32(bb) / 65535.0
			}

			rOut[y] = rRow
			gOut[y] = gRow
			bOut[y] = bRow
		}()
	}

	wg.Wait()
	return
}

// F32ToRGB converts three float32 matrices (R,G,B) into an *image.NRGBA suitable for PNG encoding.
// Each channel is normalized independently (linear min/max).
// FIXME: Add luminance normalization
func F32ToRGB(rImg, gImg, bImg [][]float32) *image.NRGBA {
	h := len(rImg)
	if h == 0 {
		return image.NewNRGBA(image.Rect(0, 0, 0, 0))
	}
	w := len(rImg[0])
	out := image.NewNRGBA(image.Rect(0, 0, w, h))

	// find per-channel min/max
	minR, maxR := float32(1e9), float32(-1e9)
	minG, maxG := float32(1e9), float32(-1e9)
	minB, maxB := float32(1e9), float32(-1e9)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := rImg[y][x]
			if v < minR {
				minR = v
			}
			if v > maxR {
				maxR = v
			}
			v = gImg[y][x]
			if v < minG {
				minG = v
			}
			if v > maxG {
				maxG = v
			}
			v = bImg[y][x]
			if v < minB {
				minB = v
			}
			if v > maxB {
				maxB = v
			}
		}
	}
	// avoid division by zero
	scaleR := float32(1.0)
	scaleG := float32(1.0)
	scaleB := float32(1.0)
	if maxR != minR {
		scaleR = 1.0 / (maxR - minR)
	}
	if maxG != minG {
		scaleG = 1.0 / (maxG - minG)
	}
	if maxB != minB {
		scaleB = 1.0 / (maxB - minB)
	}

	var wg sync.WaitGroup
	wg.Add(h)

	for y := 0; y < h; y++ {
		y := y
		go func() {
			defer wg.Done()
			rowOff := y * out.Stride
			for x := 0; x < w; x++ {
				out.Pix[rowOff+x*4+0] = uint8(255 * (rImg[y][x] - minR) * scaleR)
				out.Pix[rowOff+x*4+1] = uint8(255 * (gImg[y][x] - minG) * scaleG)
				out.Pix[rowOff+x*4+2] = uint8(255 * (bImg[y][x] - minB) * scaleB)
				// Set as fully opaque
				out.Pix[rowOff+x*4+3] = 0xFF
			}
		}()
	}

	wg.Wait()
	return out
}
