package main

import (
	"fmt"
	"github.com/dairovolzhas/dar-internship/task1/imageResizer"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var (
	port = "8080"
)



func main(){

	r := mux.NewRouter()

	r.Methods("POST").Path("/image").HandlerFunc(imageResizer.ImageProcessingHandler)

	fmt.Printf("Server started at localhost:%s\n", port)

	log.Fatal(http.ListenAndServe(":"+port, r))
}

