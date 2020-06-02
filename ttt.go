package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

var (
	url = "http://localhost:8080/image"
	path = "task1/image.jpg"

)

func main(){
	var n int
	fmt.Scanf("%d", &n)

	ItersPerSecond(n)
	//DurationForNIters(n)
}
func DurationForNIters(n int){
	t := time.Now()
	f, _ := os.Open(path)
	for i := 0; i < n; i++ {
		http.Post(url, "image/jpeg", f)
	}
	f.Close()
	fmt.Println(time.Now().Sub(t).String())
}


func ItersPerSecond(n int){
	cnt := 0
	t := time.Now()
	for int(time.Now().Sub(t).Seconds()) < n {
		f, _ := os.Open(path)
		http.Post(url, "image/jpeg", f)
		f.Close()
		cnt +=1
	}
	fmt.Println(cnt)
}