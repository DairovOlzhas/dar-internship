package imageResizer

import (
	"os"
	"strings"
	"testing"
)

var testImages = []struct{
	name		string
	filepath	string
}{
	{"BigImage", "testdata/BigImage.jpg"},
	{"MediumImage", "testdata/MediumImage.jpg"},
	{"SmallImage", "testdata/SmallImage.jpg"},
	{"PNGImage", "testdata/PNGImage.png"},
	{"JPEGImage", "testdata/JPEGImage.jpeg"},
}


func BenchmarkNewImageResizer(b *testing.B) {
	for _, bm := range testImages {
		b.Run(bm.name, func(b *testing.B) {

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

//func BenchmarkImageProcessingHandler(b *testing.B) {
//
//}


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
