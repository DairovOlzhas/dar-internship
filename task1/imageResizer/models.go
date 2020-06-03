package imageResizer

import "image"

type ImageResizer struct {
	originalImg  image.Image
	normalImg    image.Image
	thumbnailImg image.Image
	imageFormat  string
	fileName     string
}