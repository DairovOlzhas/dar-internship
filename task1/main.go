package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	port = "8080"
	normalWidth, normalHeight uint = 800, 800
	thumbnailWidth, thumbnailHeight uint = 128, 128
)

// image formats
const (
	JPEG = "jpeg"
	PNG = "png"
)

func ImageProcessingHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("image")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Sorry: " + err.Error()))
		return
	}
	defer file.Close()

	if header.Size > 5242880 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Image too large!!! Maximum file size 5 MB."))
		return
	}

	var originalImg image.Image
	var imageFormat string

	// Determine image format
	switch  {
	case strings.HasSuffix(header.Filename, ".jpeg") || strings.HasSuffix(header.Filename, ".jpg"):
		imageFormat = JPEG
		originalImg, err = jpeg.Decode(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Sorry can't decode jpeg: " + err.Error()))
			return
		}
	case strings.HasSuffix(header.Filename, ".png"):
		imageFormat = PNG
		originalImg, err = png.Decode(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Sorry can't decode png: " + err.Error()))
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported image format!!!"))
		return
	}


	normalImg, err := resizeAndCropToNormal(originalImg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Sorry: " + err.Error()))
		return
	}
	thumbnailImg := resizeToThumbnail(normalImg)

	err = saveImage(originalImg, header.Filename + "_original", imageFormat)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Sorry: " + err.Error()))
		return
	}
	err = saveImage(normalImg, header.Filename + "_normal", imageFormat)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Sorry: " + err.Error()))
		return
	}
	err = saveImage(thumbnailImg, header.Filename + "_thumbnail", imageFormat)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Sorry: " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func main(){
	r := mux.NewRouter()

	r.Methods("POST").Path("/image").HandlerFunc(ImageProcessingHandler)

	fmt.Printf("Server started at localhost:%s\n", port)

	log.Fatal(http.ListenAndServe(":"+port, r))
}

func resizeToThumbnail(originalImg image.Image) image.Image {
	newImage := resize.Thumbnail(thumbnailWidth, thumbnailHeight, originalImg, resize.NearestNeighbor)
	return newImage
}

func resizeAndCropToNormal(originalImg image.Image) (image.Image, error) {
	size := originalImg.Bounds().Size()

	width, height := size.X, size.Y
	config := cutter.Config{
		Width:   min(width, height),
		Height:  min(width, height),
		Mode:    cutter.Centered,
	}

	croppedImg, err := cutter.Crop(originalImg, config)
	if err != nil {
		return nil, err
	}

	newImage := resize.Resize(normalWidth, normalHeight, croppedImg, resize.NearestNeighbor)

	return newImage, nil
}

func saveImage(image image.Image, name string, imageFormat string) error {
	f, err := os.Create("images/" + name + "." + imageFormat)
	if err != nil {
		log.Fatal(err)
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