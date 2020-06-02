package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"
)

var (
	port = "8080"
	cnt = 0
)

func main(){
	r := mux.NewRouter()

	r.Methods("POST").Path("/image").HandlerFunc(ImageProcessingHandler)

	fmt.Printf("Server started at localhost:%s\n", port)

	log.Fatal(http.ListenAndServe(":"+port, r))
}

func ImageProcessingHandler(w http.ResponseWriter, r *http.Request) {

	img,_,err := image.Decode(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	go func(){
		newImage := resize.Resize(69, 69, img, resize.Lanczos3)

		buf := new(bytes.Buffer)
		err = jpeg.Encode(buf, newImage, nil)
		if err != nil {
			log.Fatal(err)
		}
		f, err := os.Create("images/image"+strconv.Itoa(cnt)+".png")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		err = png.Encode(f, newImage)
		if err != nil {
			log.Fatal(err)
		}
		cnt += 1
	}()
	w.WriteHeader(http.StatusCreated)
}
