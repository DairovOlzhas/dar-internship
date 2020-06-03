package main

import (
	"fmt"
	"github.com/dairovolzhas/dar-intership/task1/imageResizer"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var (
	port = "8080"
)



func main(){
	//file, err := os.Open("testdata/BigImage.jpg")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer file.Close()
	//fi, err := file.Stat()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//_, err = imageResizer.NewImageResizer(file, file.Name(), fi.Size())
	//if err != nil {
	//	log.Fatal(err)
	//}


	r := mux.NewRouter()

	r.Methods("POST").Path("/image").HandlerFunc(imageResizer.ImageProcessingHandler)

	fmt.Printf("Server started at localhost:%s\n", port)

	log.Fatal(http.ListenAndServe(":"+port, r))
}

