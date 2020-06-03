package imageResizer

import (
	"errors"
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	savingPath = "/home/dheyoungb/go/src/github.com/dairovolzhas/dar-internship/task1/imageResizer/resizedImages/"
	normalWidth, normalHeight uint = 800, 800
	thumbnailWidth, thumbnailHeight uint = 200, 200
)

// image formats
const (
	JPEG = "jpeg"
	PNG = "png"
)

func NewImageResizer(file io.Reader, fileName string, fileSize int64) (ir *ImageResizer, err error) {
	if fileSize > 2*5242880 {
		return nil, errors.New("Image too large!!! Maximum file size 5 MB.")
	}

	ir = &ImageResizer{}

	// Determine image format and decode
	switch  {
	case strings.HasSuffix(fileName, ".jpeg") || strings.HasSuffix(fileName, ".jpg"):
		ir.imageFormat = JPEG
		ir.fileName = fileName[:len(fileName)-len(ir.imageFormat)-1]
		ir.originalImg, err = jpeg.Decode(file)
		if err != nil {
			return nil, err
		}
	case strings.HasSuffix(fileName, ".png"):
		ir.imageFormat = PNG
		ir.fileName = fileName[:len(fileName)-len(ir.imageFormat)-1]
		ir.originalImg, err = png.Decode(file)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Unsupported image format!!!")
	}

	return ir, nil
}

func (ir *ImageResizer) GetOriginalImg() image.Image {
	return ir.originalImg
}

func (ir *ImageResizer) GetThumbnailImg() (image.Image, error) {
	if ir.thumbnailImg == nil {
		if ir.normalImg == nil {
			_, err := ir.GetNormalImg()
			if err != nil {
				return nil, err
			}
		}
		ir.thumbnailImg = resize.Thumbnail(thumbnailWidth, thumbnailHeight, ir.normalImg, resize.NearestNeighbor)
	}
	return ir.thumbnailImg, nil
}

func (ir *ImageResizer) GetNormalImg() (image.Image, error) {
	if ir.normalImg == nil {
		size := ir.originalImg.Bounds().Size()

		width, height := size.X, size.Y

		config := cutter.Config{
			Width:   min(width, height),
			Height:  min(width, height),
			Mode:    cutter.Centered,
		}

		croppedImg, err := cutter.Crop(ir.originalImg, config)
		if err != nil {
			return nil, err
		}

		ir.normalImg = resize.Resize(normalWidth, normalHeight, croppedImg, resize.NearestNeighbor)
	}

	return ir.normalImg, nil
}

func (ir *ImageResizer) SaveImages() error {
	err := saveImage(ir.normalImg, ir.fileName+ "_normal", ir.imageFormat)
	if err != nil {
		return err
	}
	err = saveImage(ir.originalImg, ir.fileName+ "_original", ir.imageFormat)
	if err != nil {
		return err
	}
	err = saveImage(ir.thumbnailImg, ir.fileName+ "_thumbnail", ir.imageFormat)
	if err != nil {
		return err
	}
	return nil
}

func saveImage(image image.Image, name string, imageFormat string) error {

	cur, err := os.Getwd()
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(cur, savingPath)
	if err != nil {
		return err
	}
	f, err := os.Create(rel +"/"+ name + "." + imageFormat)
	if err != nil {
		return err
	}
	defer f.Close()

	switch imageFormat {
	case JPEG:
		opt := jpeg.Options{
			Quality: 90,
		}
		err =  jpeg.Encode(f, image, &opt)
		if err != nil {
			return err
		}
	case PNG:
		err =  png.Encode(f, image)
		if err != nil {
			return err
		}
	}

	return nil
}

func min(a,b int) int {
	if a > b {
		return b
	}
	return a
}