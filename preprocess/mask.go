package preprocess

import (
	"fmt"
	"image"

	stats "github.com/montanaflynn/stats"
)

func InterpolatedBlurDetect(img image.Image) InterpolatedFloatData {
	xDim := img.Bounds().Max.X - img.Bounds().Min.X
	yDim := img.Bounds().Max.Y - img.Bounds().Min.Y
	downsampleRate := xDim / 50
	radius := xDim / 30

	fmt.Println("InterpolatedBlurDetect", xDim, yDim, downsampleRate)
	interpData := NewInterpolatedFloatData(xDim, yDim, downsampleRate)
	lumMap := LuminanceMap(img)

	fmt.Println("lum map complete")
	fmt.Println("starting sampling", xDim/downsampleRate, yDim/downsampleRate)

	for y := 0; y < (yDim / downsampleRate); y++ {
		for x := 0; x < (xDim / downsampleRate); x++ {

			px, _ := localBlurDetect(lumMap, x*downsampleRate, y*downsampleRate, radius)
			interpData.Data[y][x] = px
		}
	}
	fmt.Println("Downsampling completed")

	interpData.Normalize()

	fmt.Println(("normalized"))

	return interpData
}

func localBlurDetect(lumMap [][]float64, dx, dy, radius int) (float64, error) {
	startY := max(dy-radius, 0)
	startX := max(dx-radius, 0)
	endY := min(dy+radius, len(lumMap))
	endX := min(dx+radius, len(lumMap[0]))
	var lums []float64

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			lums = append(lums, lumMap[y][x])
		}
	}
	return stats.VarP(lums)
}
