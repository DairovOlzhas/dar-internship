package imageResizer

import (
	"encoding/json"
	"net/http"
)


// Proceeds got image and return links to saved resulting images
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

	nrmlImgPath, origImgPath, tbnlImgPath, err := imageResizer.SaveImages()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Sorry: " + err.Error()))
		return
	}

	result := Result{
		NrmlImgPath: nrmlImgPath,
		OrigImgPath: origImgPath,
		TbnlImgPath: tbnlImgPath,
	}
	data, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Sorry: " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

