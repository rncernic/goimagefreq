# goImageFreq — Frequency-Domain Image Processing in Go

`goimagefreq` is a **pure Go**, **high-performance** image-processing library focused on **frequency-domain techniques**, with a strong emphasis on **astrophotography** and **color-preserving workflows**.

The package implements **Gaussian filtering**, **low/high frequency separation**, **à trous wavelets**, **PixInsight-like MLT**, **luminance-only denoising**, and **color-safe deconvolution**, all using **float32 linear pipelines** and **parallel goroutines**.

> No external dependencies.  
> No color shifts.  
> Perfect reconstruction.

It will be used as part of a bigger project I'm working on.

---

## Features

### Core frequency processing
- Separable **Gaussian blur**
- **Low / High frequency split**
- **Multiband frequency decomposition**
- **À trous undecimated wavelet transform**
- Rreconstruction

### Astrophotography-grade processing
- **Wavelet denoising (MAD-based, PixInsight style)**
- **Multi-Scale Linear Transform (MLT)**
- **Richardson–Lucy deconvolution**
- Noise estimation using **MAD / 0.6745**

### Color-safe pipelines
- **YCbCr luminance-only blur** (fast, preview-friendly)
- **CIELAB L\*-only processing** (perceptual, high quality)
- Guaranteed chroma preservation (no RGB channel blurring)

### Performance & design
- Fully **parallelized** using goroutines
- Row-based concurrency



