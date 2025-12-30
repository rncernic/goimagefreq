// Color space conversions
package goimagefreq

import (
	"math"
	"sync"
)

// BT.709 coefficients (same as sRGB luminance)
func RGBToYCbCr(r, g, b float32) (y, cb, cr float32) {
	y = 0.2126*r + 0.7152*g + 0.0722*b
	cb = (b - y) * 0.5389
	cr = (r - y) * 0.6350
	return
}

func YCbCrToRGB(y, cb, cr float32) (r, g, b float32) {
	r = y + 1.5748*cr
	b = y + 1.8556*cb
	g = y - 0.1873*cb - 0.4681*cr
	return
}

func pivotXYZ(t float64) float64 {
	if t > 0.008856 {
		return math.Cbrt(t)
	}
	return (7.787 * t) + 16.0/116.0
}

func invPivotXYZ(t float64) float64 {
	t3 := t * t * t
	if t3 > 0.008856 {
		return t3
	}
	return (t - 16.0/116.0) / 7.787
}

// sRGB linear RGB → XYZ (D65)
func RGBToXYZ(r, g, b float64) (x, y, z float64) {
	x = 0.4124564*r + 0.3575761*g + 0.1804375*b
	y = 0.2126729*r + 0.7151522*g + 0.0721750*b
	z = 0.0193339*r + 0.1191920*g + 0.9503041*b
	return
}

func XYZToRGB(x, y, z float64) (r, g, b float64) {
	r = 3.2404542*x - 1.5371385*y - 0.4985314*z
	g = -0.9692660*x + 1.8760108*y + 0.0415560*z
	b = 0.0556434*x - 0.2040259*y + 1.0572252*z
	return
}

func RGBToLab(r, g, b float32) (L, a, bb float32) {
	x, y, z := RGBToXYZ(float64(r), float64(g), float64(b))

	// D65 white
	x /= 0.95047
	z /= 1.08883

	fx := pivotXYZ(x)
	fy := pivotXYZ(y)
	fz := pivotXYZ(z)

	L = float32(116*fy - 16)
	a = float32(500 * (fx - fy))
	bb = float32(200 * (fy - fz))
	return
}

func LabToRGB(L, a, bb float32) (r, g, b float32) {
	fy := (float64(L) + 16) / 116
	fx := fy + float64(a)/500
	fz := fy - float64(bb)/200

	x := invPivotXYZ(fx) * 0.95047
	y := invPivotXYZ(fy)
	z := invPivotXYZ(fz) * 1.08883

	rr, gg, bb2 := XYZToRGB(x, y, z)
	return float32(rr), float32(gg), float32(bb2)
}

// func RGBToLabImage(
// 	r, g, b [][]float32,
// ) (L, a, b2 [][]float32) {

// 	h := len(r)
// 	w := len(r[0])

// 	L = make([][]float32, h)
// 	a = make([][]float32, h)
// 	b2 = make([][]float32, h)

// 	for y := 0; y < h; y++ {
// 		L[y] = make([]float32, w)
// 		a[y] = make([]float32, w)
// 		b2[y] = make([]float32, w)

// 		for x := 0; x < w; x++ {
// 			L[y][x], a[y][x], b2[y][x] =
// 				RGBToLab(r[y][x], g[y][x], b[y][x])
// 		}
// 	}
// 	return
// }

// func LabToRGBImage(
// 	L, a, b2 [][]float32,
// ) (r, g, b [][]float32) {

// 	h := len(L)
// 	w := len(L[0])

// 	r = make([][]float32, h)
// 	g = make([][]float32, h)
// 	b = make([][]float32, h)

// 	for y := 0; y < h; y++ {
// 		r[y] = make([]float32, w)
// 		g[y] = make([]float32, w)
// 		b[y] = make([]float32, w)

// 		for x := 0; x < w; x++ {
// 			r[y][x], g[y][x], b[y][x] =
// 				LabToRGB(L[y][x], a[y][x], b2[y][x])
// 		}
// 	}
// 	return
// }

func RGBToLabImage(
	r, g, b [][]float32,
) (L, a, b2 [][]float32) {

	h := len(r)
	w := len(r[0])

	L = make([][]float32, h)
	a = make([][]float32, h)
	b2 = make([][]float32, h)

	var wg sync.WaitGroup
	wg.Add(h)

	for y := 0; y < h; y++ {
		y := y
		go func() {
			defer wg.Done()

			L[y] = make([]float32, w)
			a[y] = make([]float32, w)
			b2[y] = make([]float32, w)

			for x := 0; x < w; x++ {
				L[y][x], a[y][x], b2[y][x] =
					RGBToLab(r[y][x], g[y][x], b[y][x])
			}
		}()
	}

	wg.Wait()
	return
}

func LabToRGBImage(
	L, a, b2 [][]float32,
) (r, g, b [][]float32) {

	h := len(L)
	w := len(L[0])

	r = make([][]float32, h)
	g = make([][]float32, h)
	b = make([][]float32, h)

	var wg sync.WaitGroup
	wg.Add(h)

	for y := 0; y < h; y++ {
		y := y
		go func() {
			defer wg.Done()

			r[y] = make([]float32, w)
			g[y] = make([]float32, w)
			b[y] = make([]float32, w)

			for x := 0; x < w; x++ {
				r[y][x], g[y][x], b[y][x] =
					LabToRGB(L[y][x], a[y][x], b2[y][x])
			}
		}()
	}

	wg.Wait()
	return
}

func RGBToYCbCrImage(
	r, g, b [][]float32,
) (y, cb, cr [][]float32) {

	h := len(r)
	w := len(r[0])

	y = make([][]float32, h)
	cb = make([][]float32, h)
	cr = make([][]float32, h)

	var wg sync.WaitGroup
	wg.Add(h)

	for i := 0; i < h; i++ {
		i := i
		go func() {
			defer wg.Done()

			y[i] = make([]float32, w)
			cb[i] = make([]float32, w)
			cr[i] = make([]float32, w)

			for x := 0; x < w; x++ {
				y[i][x], cb[i][x], cr[i][x] =
					RGBToYCbCr(r[i][x], g[i][x], b[i][x])
			}
		}()
	}

	wg.Wait()
	return
}

func YCbCrToRGBImage(
	y, cb, cr [][]float32,
) (r, g, b [][]float32) {

	h := len(y)
	w := len(y[0])

	r = make([][]float32, h)
	g = make([][]float32, h)
	b = make([][]float32, h)

	var wg sync.WaitGroup
	wg.Add(h)

	for i := 0; i < h; i++ {
		i := i
		go func() {
			defer wg.Done()

			r[i] = make([]float32, w)
			g[i] = make([]float32, w)
			b[i] = make([]float32, w)

			for x := 0; x < w; x++ {
				r[i][x], g[i][x], b[i][x] =
					YCbCrToRGB(y[i][x], cb[i][x], cr[i][x])
			}
		}()
	}

	wg.Wait()
	return
}
