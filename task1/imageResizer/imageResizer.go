package imageResizer

import (
	"errors"
	"fmt"
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
	// path for saving images locally
	savingPath = "/home/dheyoungb/go/src/github.com/dairovolzhas/dar-internship/task1/imageResizer/resizedImages/"
	// Size of normal image
	normalWidth, normalHeight uint = 800, 800
	// Size of thumbnail image
	thumbnailWidth, thumbnailHeight uint = 200, 200
	// Maximum image size in megabytes
	maxImageSize int64 = 5
)

// image formats
const (
	JPEG = "jpeg"
	PNG = "png"
)

// Returns ImageResizer which contains only original image and their
// field such as imageFormat and fileName. Normal and thumbnail sized images
// will be proceed when needed it.
func NewImageResizer(file io.Reader, fileName string, fileSize int64) (ir *ImageResizer, err error) {
	// check for image size
	if fileSize > maxImageSize*1024*1024 {
		return nil, errors.New(fmt.Sprintf("Image too large!!! Maximum file size %d MB.", maxImageSize))
	}

	ir = &ImageResizer{}

	// Determine image format and decode.
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

// Returns original image without changes.
func (ir *ImageResizer) GetOriginalImg() image.Image {
	return ir.originalImg
}

// Returns resized image via downscaling normal image.
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

// Returns cropped by center and resized image of normal size.
// There are checking for existing of normal image.
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
// Saves images.
// Returns paths to saved images.
func (ir *ImageResizer) SaveImages() (nrmlImgPath, origImgPath, tbnlImgPath string, err error) {
	nrmlImgPath, err = saveImage(ir.normalImg, ir.fileName+ "_normal", ir.imageFormat)
	if err != nil {
		return
	}
	origImgPath, err = saveImage(ir.originalImg, ir.fileName+ "_original", ir.imageFormat)
	if err != nil {
		return
	}
	tbnlImgPath, err = saveImage(ir.thumbnailImg, ir.fileName+ "_thumbnail", ir.imageFormat)
	if err != nil {
		return
	}
	return
}

// Returns path where image saved.
func saveImage(image image.Image, name string, imageFormat string) (string, error) {
	// getting relative path between runtime and image saving dirs
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(cur, savingPath)
	if err != nil {
		return "", err
	}

	f, err := os.Create(rel +"/"+ name + "." + imageFormat)
	if err != nil {
		return "", errors.New(fmt.Sprintf("%v\ncur: %v\nsaving: %v\nrel: %v\n", err.Error(), cur, savingPath, rel))
	}
	defer f.Close()

	switch imageFormat {
	case JPEG:
		opt := jpeg.Options{
			Quality: 90,
		}
		err =  jpeg.Encode(f, image, &opt)
		if err != nil {
			return "", err
		}
	case PNG:
		err =  png.Encode(f, image)
		if err != nil {
			return "", err
		}
	}


	return cur + "/" + rel + "/"+ name + "." + imageFormat, nil
}

func min(a,b int) int {
	if a > b {
		return b
	}
	return a
}