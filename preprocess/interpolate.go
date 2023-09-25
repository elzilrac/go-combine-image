package preprocess

import (
	"image"
	"image/color"
	"math"
	"sort"
)

type InterpolatedValueAt interface {
	At(dx, dy int) float64
}

type InterpolatedFloatData struct {
	Data           [][]float64
	DownsampleRate int // ie "1 in DownsampleRate"
	OriginalX      int
	OriginalY      int
}

func (d InterpolatedFloatData) BasicAt(dx, dy int) float64 {
	x := math.Floor(float64(dx) / float64(d.DownsampleRate))
	y := math.Floor(float64(dy) / float64(d.DownsampleRate))
	return d.Data[int(y)][int(x)]
}

// At uses bilinear interpolation
func (d InterpolatedFloatData) At(dx, dy int) float64 {
	// We use 4 sample points to create 2 proxy values
	// which we then use again to determine the final
	/*
		Q11 ----- P1 ----- Q12
		 |         |		|
		 |-------- x ------ |
		 |		   | 		|
		Q21------ P2 ------ Q22
	*/
	x_ := float64(dx) / float64(d.DownsampleRate)
	y_ := float64(dy) / float64(d.DownsampleRate)

	x1 := (math.Ceil(float64(dx) / float64(d.DownsampleRate)))
	y1 := (math.Ceil(float64(dy) / float64(d.DownsampleRate)))
	x2 := (math.Floor(float64(dx) / float64(d.DownsampleRate)))
	y2 := (math.Floor(float64(dy) / float64(d.DownsampleRate)))

	// Boundary checks
	if int(y1) >= len(d.Data) {
		y1 = float64(len(d.Data) - 1)
	}
	if int(y2) >= len(d.Data) {
		y2 = float64(len(d.Data) - 1)
	}
	if int(x2) >= len(d.Data[0]) {
		x2 = float64(len(d.Data[0]) - 1)
	}
	if int(x1) >= len(d.Data[0]) {
		x1 = float64(len(d.Data[0]) - 1)
	}

	q11 := d.Data[int(y1)][int(x1)]
	q12 := d.Data[int(y2)][int(x1)]
	q21 := d.Data[int(y1)][int(x2)]
	q22 := d.Data[int(y2)][int(x2)]

	r1 := ((x2 - x_) / (x2 - x1) * q11) + ((x_ - x1) / (x2 - x1) * q21)
	r2 := ((x2 - x_) / (x2 - x1) * q12) + ((x_ - x1) / (x2 - x1) * q22)
	if x_ == math.Floor(x_) {
		r1 = q11
		r2 = q22
	}
	out := (y2-y_)/(y2-y1)*r1 + (y_-y1)/(y2-y1)*r2
	if y_ == math.Floor(y_) {
		return r1
	}

	return out
}

// IDWAt uses inverse distance weighted interpolation
func (d InterpolatedFloatData) IDWAt(dx, dy int) float64 {
	var sumWeightedValues float64
	var denominator float64
	// Only use N closest points
	// Depending on the number of points you use, you get very different
	// looking artifcats when your sample points are on a grid
	power := 1.0                   // power changes the "dropoff" effect due to distance
	var closestPoints [][2]float64 // first one is the distance

	for iy := range d.Data {
		for ix := range d.Data[iy] {
			dist := d.distance(dx, dy, ix, iy)
			value := d.Data[iy][ix]
			closestPoints = append(closestPoints, [2]float64{dist, value})
			if dist == 0 {
				return value
			}
		}
	}

	sort.Slice(closestPoints, func(i, j int) bool {
		return closestPoints[i][0] < closestPoints[j][0]
	})

	for i := 0; i < 2; i++ {
		dPow := math.Pow(closestPoints[i][0], power)
		sumWeightedValues += (closestPoints[i][1] / dPow)
		denominator += (1 / dPow)
	}

	return sumWeightedValues / denominator
}

func NewInterpolatedFloatData(dx, dy, downsampleRate int) InterpolatedFloatData {
	xDim := dx / downsampleRate
	yDim := dy / downsampleRate
	data := make([][]float64, yDim+1)
	for i := range data {
		data[i] = make([]float64, xDim+1)
	}
	return InterpolatedFloatData{
		Data:           data,
		DownsampleRate: downsampleRate,
		OriginalX:      dx,
		OriginalY:      dy,
	}
}

func (d InterpolatedFloatData) ToImage() *image.RGBA {
	outImg := image.NewRGBA(image.Rect(0, 0, d.OriginalX, d.OriginalY))

	for y := 0; y < d.OriginalY; y++ {
		for x := 0; x < d.OriginalX; x++ {
			value := d.At(x, y)
			outImg.Set(x, y, color.Gray{uint8(value * 255)})
		}
	}
	return outImg
}

func (d InterpolatedFloatData) Normalize() {
	max := float64(0)
	min := math.MaxFloat64
	// Find the max and min
	for y := 0; y < len(d.Data); y++ {
		for x := 0; x < len(d.Data[0]); x++ {
			if d.Data[y][x] > max {
				max = d.Data[y][x]
			}
			if d.Data[y][x] < min {
				min = d.Data[y][x]
			}
		}
	}
	// Normalize
	for y := 0; y < len(d.Data); y++ {
		for x := 0; x < len(d.Data[0]); x++ {
			v := d.Data[y][x]
			newV := (v - min) / (max - min)
			d.Data[y][x] = newV
		}
	}
}

// distance calculates the distance between a given point (dx, dy), and
// the interpolated value (ix, iy) which is downsampled
func (d InterpolatedFloatData) distance(dx, dy, ix, iy int) float64 {
	xComponent := math.Pow(float64(dx-(ix*d.DownsampleRate)), 2)
	yComponent := math.Pow(float64(dy-(iy*d.DownsampleRate)), 2)
	return math.Sqrt(xComponent + yComponent)
}
