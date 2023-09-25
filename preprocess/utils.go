package preprocess

import (
	"image"
	"image/color"
	"math"
)

// Luminance returns the precieved brightness of a color (0,1)
func Luminance(c color.Color) float64 {
	// src: https://stackoverflow.com/questions/596216/formula-to-determine-perceived-brightness-of-rgb-color
	// sqrt( 0.299*R^2 + 0.587*G^2 + 0.114*B^2 )
	// This forumula is for color in [0-1] space
	r, g, b, _ := c.RGBA()
	r = r >> 8
	g = g >> 8
	b = b >> 8
	// Now it's in 255 space

	rComp := (0.299 * math.Pow(float64(r)/255, 2))
	gComp := (0.587 * math.Pow(float64(g)/255, 2))
	bComp := (0.114 * math.Pow(float64(b)/255, 2))
	return math.Sqrt(rComp + gComp + bComp)
}

// LuminanceMap creates a 2d slice of luminance values [0,1]
func LuminanceMap(img image.Image) [][]float64 {
	xDim := img.Bounds().Max.X - img.Bounds().Min.X
	yDim := img.Bounds().Max.Y - img.Bounds().Min.Y
	out := make([][]float64, yDim)
	for i := range out {
		out[i] = make([]float64, xDim)
	}

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			out[y][x] = Luminance(img.At(x, y))
		}
	}
	return out
}
