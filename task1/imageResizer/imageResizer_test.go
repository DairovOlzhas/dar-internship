package imageResizer

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var testImages = []struct{
	name		string
	filepath	string
}{
	{"BigImage", "testdata/BigImage.jpg", },
	{"MediumImage", "testdata/MediumImage.jpg"},
	{"SmallImage", "testdata/SmallImage.jpg"},
	{"PNGImage", "testdata/PNGImage.png"},
	{"JPEGImage", "testdata/JPEGImage.jpeg"},
}

func TestImageProcessingHandler(t *testing.T) {
	for _, tt := range testImages {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(ImageProcessingHandler))
			defer ts.Close()

			file, err := os.Open(tt.filepath)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("image", strings.ReplaceAll(tt.filepath, "/", ""))
			io.Copy(part, file)
			writer.Close()

			req, err := http.NewRequest("POST", ts.URL, body)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-type", writer.FormDataContentType())
			client := &http.Client{}
			resp, err := client.Do(req)

			result := Result{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				t.Fatal(err)
			}

			if status := resp.StatusCode; status != http.StatusCreated {
				t.Errorf("Got %v, but expected %v", status, http.StatusCreated)
			}
			if _,ok := os.Stat(relPath(result.NrmlImgPath)); os.IsNotExist(ok) {
				t.Errorf("Normal image doesn't exist in the response path")
			}
			if _,ok := os.Stat(relPath(result.OrigImgPath)); os.IsNotExist(ok) {
				t.Errorf("Original image doesn't exist in the response path")
			}
			if _,ok := os.Stat(relPath(result.TbnlImgPath)); os.IsNotExist(ok) {
				t.Errorf("Thumbnail image doesn't exist in the response path")
			}
		})
	}
}

func TestNewImageResizer(t *testing.T) {
	for _, tt := range testImages {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.Open(tt.filepath)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			fi, err := file.Stat()
			if err != nil {
				t.Fatal(err)
			}
			ir, err := NewImageResizer(file, file.Name(), fi.Size())
			if err != nil {
				t.Fatal(err)
			}


		})
	}
}

func BenchmarkNewImageResizer(b *testing.B) {
	for _, bm := range testImages {
		b.Run(bm.name, func(b *testing.B) {
			b.StopTimer()
			b.ResetTimer()

			for i:=0; i < b.N; i++ {
				file, err := os.Open(bm.filepath)
				if err != nil {
					b.Fatal(err)
				}
				defer file.Close()
				fi, err := file.Stat()
				if err != nil {
					b.Fatal(err)
				}
				_, err = NewImageResizer(file, file.Name(), fi.Size())
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkImageResizer_GetOriginalImg(b *testing.B) {
	for _, bm := range testImages {
		b.Run(bm.name, func(b *testing.B) {

			ir, err := imageResizerFromImagePath(bm.filepath)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()

			for i:=0; i < b.N; i++ {
				_ = ir.GetOriginalImg()
			}
		})
	}
}

func BenchmarkImageResizer_GetNormalImg(b *testing.B) {
	for _, bm := range testImages {
		b.Run(bm.name, func(b *testing.B) {

			ir, err := imageResizerFromImagePath(bm.filepath)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()

			for i:=0; i < b.N; i++ {
				_, err = ir.GetNormalImg()
				if err != nil {
					b.Fatal(err)
				}
				ir.normalImg = nil
			}
		})
	}
}

func BenchmarkImageResizer_GetThumbnailImg(b *testing.B) {
	for _, bm := range testImages {
		b.Run(bm.name, func(b *testing.B) {

			ir, err := imageResizerFromImagePath(bm.filepath)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()

			for i:=0; i < b.N; i++ {
				_, err = ir.GetThumbnailImg()
				if err != nil {
					b.Fatal(err)
				}
				//ir.normalImg = nil
				ir.thumbnailImg = nil
			}
		})
	}
}

func BenchmarkImageResizer_SaveImages(b *testing.B) {
	for _, bm := range testImages {
		b.Run(bm.name, func(b *testing.B) {

			ir, err := imageResizerFromImagePath(bm.filepath)
			if err != nil {
				b.Fatal(err)
			}
			ir.fileName = strings.ReplaceAll(ir.fileName, "/", "")

			_, err = ir.GetNormalImg()
			if err != nil {
				b.Fatal(err)
			}
			_, err = ir.GetThumbnailImg()
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()

			for i:=0; i < b.N; i++ {
				_, _, _, err = ir.SaveImages()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkImageProcessingHandler(b *testing.B) {
	for _, bm := range testImages {
		b.Run(bm.name, func(b *testing.B) {
			ts := httptest.NewServer(http.HandlerFunc(ImageProcessingHandler))
			defer ts.Close()
			b.StopTimer()
			b.ResetTimer()
			for i:=0; i < b.N; i++ {
				file, err := os.Open(bm.filepath)
				if err != nil {
					b.Fatal(err)
				}
				defer file.Close()

				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("image", strings.ReplaceAll(bm.filepath, "/", ""))
				io.Copy(part, file)
				writer.Close()

				req, err := http.NewRequest("POST", ts.URL, body)
				if err != nil {
					b.Fatal(err)
				}
				req.Header.Set("Content-type", writer.FormDataContentType())
				client := &http.Client{}
				b.StartTimer()
				resp, err := client.Do(req)
				b.StopTimer()
				if err != nil {
					b.Fatal(err)
				}
				if status := resp.StatusCode; status != http.StatusCreated {
					b.Errorf("Got %v, but expected %v", status, http.StatusCreated)
				}
			}
		})
	}
}
	

func imageResizerFromImagePath(path string) (*ImageResizer, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	ir, err := NewImageResizer(file, file.Name(), fi.Size())
	if err != nil {
		return nil, err
	}

	return ir, nil
}

func relPath(path string) string{
	cur, _ := os.Getwd()
	rel, _ := filepath.Rel(cur, path)
	return rel
}