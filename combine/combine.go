package combine

import (
	"fmt"
	"image"
	"image/color"

	preprocess "github.com/elzilrac/go-combine-image/preprocess"
)

type Operation string

const (
	Luminance Operation = "luminance"
	Mask      Operation = "mask"
)

// colorTransfBasic darkens c1 with c2, ratio is the balance between c1 and c2
func colorTransfBasic(c1, c2 uint32, ratio float64) uint8 {
	m1 := float64(c1) * ratio
	m2 := float64(c2) * (1.0 - ratio)
	return uint8(int64(m1+m2) >> 8)
}

// // colorTransfBlue provides a cool negative effect
// func colorTransfBlue(c1, c2 uint32) uint8 {
// 	return uint8(int32((float32(c1)/2)-(float32(c2)/2)) >> 8)
// }

// combine two images. The first image is the more diminant one.
func combineLuminance(c1, c2 color.Color) color.RGBA {
	var ratio float64 = 0.5
	lum1 := preprocess.Luminance(c1)
	lum2 := preprocess.Luminance(c2)

	img2ShadowThreshold := 0.3
	img2LightThreshold := 0.8

	if lum1 > 0.9 {
		ratio = lum1

	} else if lum2 < img2ShadowThreshold {
		// Shadow from c2
		// we want it to go from 0.5 to 0.0 gradually
		// y-y1 = m(x-x1) where y, x = 0, and y1 = 0.5 and x1 = img2ShadowThreshold
		ratio = (0.5 / img2ShadowThreshold) * lum2

	} else if lum2 > img2LightThreshold {
		// Highlight from c2 is not as powerful
		// Want it to go from 0.5 to 0.2 gracefully
		// y-y1 = m(x-x1)
		// y = 0.2, y1 = 0.5 upper and lower bounds of "ratio"
		// x = 1.0 and x1 = img2LightThreshold (luminance)
		// m = (y - y1)/(x - x1)
		// c = intercept = y - m
		m := (0.2 - 0.5) / (1.0 - img2LightThreshold)
		c := 0.2 - m
		ratio = (m * lum2) + c
	}

	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	outColor := color.RGBA{
		colorTransfBasic(r1, r2, ratio),
		colorTransfBasic(g1, g2, ratio),
		colorTransfBasic(b1, b2, ratio),
		uint8((a1 + a2) >> 8),
	}
	return outColor
}

// combine two images. The first image is the more diminant one.
func combineMask(c1, c2 color.Color, mask1Value, mask2Value float64) color.RGBA {
	ratio := (mask1Value + (1 - mask2Value)) / 2

	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	outColor := color.RGBA{
		colorTransfBasic(r1, r2, ratio),
		colorTransfBasic(g1, g2, ratio),
		colorTransfBasic(b1, b2, ratio),
		uint8((a1 + a2) >> 8),
	}
	return outColor
}

func Combine(img1, img2 image.Image, op Operation) *image.RGBA {
	outImg := image.NewRGBA(image.Rect(0, 0, img1.Bounds().Dx(), img1.Bounds().Dy()))
	mask1 := preprocess.InterpolatedBlurDetect(img1)
	mask2 := preprocess.InterpolatedBlurDetect(img2)

	var outColor color.RGBA
	for y := img1.Bounds().Min.Y; y < img1.Bounds().Max.Y; y++ {
		for x := img1.Bounds().Min.X; x < img1.Bounds().Max.X; x++ {
			if op == Mask {
				outColor = combineMask(img1.At(x, y), img2.At(x, y), mask1.At(x, y), mask2.At(x, y))
			} else if op == Luminance {
				outColor = combineLuminance(img1.At(x, y), img2.At(x, y))
			}

			outImg.Set(x, y, outColor)
		}
	}
	fmt.Println("completed combining")
	return outImg
}
