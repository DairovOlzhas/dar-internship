package main

import (
	"net/http"
	"os"
	"testing"
)

func BenchmarkImageProcessingHandler(b *testing.B) {
	url := "http://localhost:8080/image"
	f, _ := os.Open("image.jpg")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		http.Post(url, "image/jpeg", f)
	}
}