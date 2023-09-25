package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"

	combine "github.com/elzilrac/go-combine-image/combine"
	"github.com/elzilrac/go-combine-image/preprocess"
)

func getFileType(imgPath string) string {
	imgFile, err := os.Open(imgPath)
	if err != nil {
		log.Fatal(err)
	}
	defer imgFile.Close()

	_, imType, err := image.Decode(imgFile)
	if err != nil {
		log.Fatal(err)
	}
	return imType
}

func openImage(imgPath string) (image.Image, error) {
	imType := getFileType(imgPath)
	imgFile, err := os.Open(imgPath)
	if err != nil {
		log.Fatal(err)
	}
	defer imgFile.Close()

	var img image.Image
	switch imType {
	case "jpeg":
		img, err = jpeg.Decode(imgFile)
	case "png":
		img, err = png.Decode(imgFile)
	default:
		err = fmt.Errorf("image: image type %s not supported", imType)
	}
	return img, err
}

func main() {
	var inputImages string
	flag.StringVar(&inputImages, "images", "", "one or more images separated with a comma")

	var doCombine bool
	flag.BoolVar(&doCombine, "combine", false, "combine two images")

	var doCombineMask bool
	flag.BoolVar(&doCombineMask, "combine-subject", false, "combine two images using 'importance' mask to estimate the subjects of the images")

	var doCreateMask bool
	flag.BoolVar(&doCreateMask, "mask", false, "create a mask of the image of important areas")

	flag.Parse()

	var images []image.Image

	for _, filePath := range strings.Split(inputImages, ",") {
		img, err := openImage(filePath)
		if err != nil {
			log.Fatal(err)
		}
		images = append(images, img)
	}

	// 2nd filePath1 := "/home/nodda/Pictures/2022/08/14/DSC_0360.JPG"
	// filePath1 := "/home/nodda/Pictures/2022/10/28/DSC_1264.JPG"
	// filePath2 := "/home/nodda/Pictures/2022/10/29/DSC_1321.JPG"

	var outImg *image.RGBA
	if doCombine {
		// go run main.go --combine --images=/home/nodda/Pictures/2022/10/28/DSC_1264.JPG,/home/nodda/Pictures/2022/10/29/DSC_1321.JPG
		if len(images) != 2 {
			err := fmt.Errorf("combine: %d images provided, need 2", len(images))
			log.Fatal(err)
		}
		outImg = combine.Combine(images[0], images[1], combine.Luminance)
	} else if doCombineMask {
		if len(images) != 2 {
			err := fmt.Errorf("combine: %d images provided, need 2", len(images))
			log.Fatal(err)
		}
		outImg = combine.Combine(images[0], images[1], combine.Mask)
	} else if doCreateMask {
		if len(images) != 1 {
			err := fmt.Errorf("mask: %d images provided, need 1", len(images))
			log.Fatal(err)
		}
		fmt.Println("interpolating image sharpness")
		outInterpolatedImg := preprocess.InterpolatedBlurDetect(images[0])
		fmt.Println("done creating downsampled blurdetect")
		outImg = outInterpolatedImg.ToImage()

	} else {
		err := fmt.Errorf("transform: Please pick an action")
		log.Fatal(err)
	}

	outFile, err := os.Create("outImg.jpeg")
	if err != nil {
		log.Fatal(err)
	}
	options := jpeg.Options{Quality: 100}
	jpeg.Encode(outFile, outImg, &options)
}
