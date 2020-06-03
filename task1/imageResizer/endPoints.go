package imageResizer

import (
	"net/http"
)



func ImageProcessingHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("image")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Sorry: " + err.Error()))
		return
	}
	defer file.Close()

	imageResizer, err := NewImageResizer(file, header.Filename, header.Size)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Sorry: " + err.Error()))
		return
	}
	_, err = imageResizer.GetNormalImg()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Sorry: " + err.Error()))
		return
	}

	_, err = imageResizer.GetThumbnailImg()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Sorry: " + err.Error()))
		return
	}

	err = imageResizer.SaveImages()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Sorry: " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

