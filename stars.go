// TODO: implement star related functions
// migrate to a different package
// added here just to test frequency-domain functions

package goimagefreq

import "math"

type Star struct {
	X, Y int
	Peak float32
}

func DetectStars(
	L [][]float32,
	threshold float32,
	minDist int,
) []Star {

	h := len(L)
	w := len(L[0])

	var stars []Star

	for y := minDist; y < h-minDist; y++ {
		for x := minDist; x < w-minDist; x++ {

			v := L[y][x]
			if v < threshold {
				continue
			}

			// Local maximum test
			isMax := true
			for dy := -1; dy <= 1 && isMax; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dy == 0 && dx == 0 {
						continue
					}
					if L[y+dy][x+dx] >= v {
						isMax = false
						break
					}
				}
			}

			if isMax {
				stars = append(stars, Star{x, y, v})
			}
		}
	}

	return stars
}

func ExtractPatch(
	L [][]float32,
	x, y, r int,
) [][]float32 {

	size := 2*r + 1
	patch := make([][]float32, size)

	for j := -r; j <= r; j++ {
		row := make([]float32, size)
		for i := -r; i <= r; i++ {
			row[i+r] = L[y+j][x+i]
		}
		patch[j+r] = row
	}
	return patch
}

func NormalizePatch(p [][]float32) {
	var sum float32
	for y := range p {
		for x := range p[y] {
			sum += p[y][x]
		}
	}
	if sum == 0 {
		return
	}
	inv := 1 / sum
	for y := range p {
		for x := range p[y] {
			p[y][x] *= inv
		}
	}
}

func StackPatches(patches [][][]float32) [][]float32 {

	n := len(patches)
	h := len(patches[0])
	w := len(patches[0][0])

	out := make([][]float32, h)
	for y := range out {
		out[y] = make([]float32, w)
	}

	for _, p := range patches {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				out[y][x] += p[y][x]
			}
		}
	}

	inv := float32(1.0 / float64(n))
	for y := range out {
		for x := range out[y] {
			out[y][x] *= inv
		}
	}

	return out
}

func RadialProfile(psf [][]float32) []float64 {

	h := len(psf)
	w := len(psf[0])
	cx := float64(w-1) / 2
	cy := float64(h-1) / 2

	maxR := int(math.Min(cx, cy))
	sum := make([]float64, maxR+1)
	cnt := make([]int, maxR+1)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			r := int(math.Hypot(dx, dy))
			if r <= maxR {
				sum[r] += float64(psf[y][x])
				cnt[r]++
			}
		}
	}

	profile := make([]float64, maxR+1)
	for i := range profile {
		if cnt[i] > 0 {
			profile[i] = sum[i] / float64(cnt[i])
		}
	}
	return profile
}

func FitMoffat(profile []float64) (alpha, beta float64) {

	bestErr := math.Inf(1)

	for a := 0.5; a <= 6; a += 0.1 {
		for b := 1.5; b <= 6; b += 0.1 {

			var err float64
			for r, v := range profile {
				model := math.Pow(1+(float64(r)*float64(r))/(a*a), -b)
				d := v - model
				err += d * d
			}

			if err < bestErr {
				bestErr = err
				alpha = a
				beta = b
			}
		}
	}

	return
}

func MoffatKernel1D(alpha, beta float64, radius int) []float64 {

	size := 2*radius + 1
	k := make([]float64, size)

	var sum float64
	for i := -radius; i <= radius; i++ {
		r := float64(i)
		v := math.Pow(1+(r*r)/(alpha*alpha), -beta)
		k[i+radius] = v
		sum += v
	}

	for i := range k {
		k[i] /= sum
	}

	return k
}

func EstimatePSF(
	L [][]float32,
	threshold float32,
) (kx, ky []float64) {

	stars := DetectStars(L, threshold, 10)

	var patches [][][]float32
	for _, s := range stars {
		p := ExtractPatch(L, s.X, s.Y, 10)
		NormalizePatch(p)
		patches = append(patches, p)
		if len(patches) >= 50 {
			break
		}
	}

	psf := StackPatches(patches)
	profile := RadialProfile(psf)
	alpha, beta := FitMoffat(profile)

	k := MoffatKernel1D(alpha, beta, 10)

	return k, k
}
