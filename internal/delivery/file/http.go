package file

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mmikhail2001/test-clever-search/internal/domain/file"
)

type Handler struct {
	usecase Usecase
}

func NewHandler(usecase Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	// TODO: это не работает
	// смог загрузить файл 500 МБ
	err := r.ParseMultipartForm(30 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	f, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer f.Close()

	folder := r.FormValue("folder")

	fmt.Printf("Uploaded File: %+v\n", handler.Filename)

	err = h.usecase.Upload(r.Context(), file.File{
		File:        f,
		Filename:    handler.Filename,
		Size:        handler.Size,
		Path:        folder + "/" + handler.Filename,
		ContentType: handler.Header["Content-Type"][0],
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully: %s", handler.Filename)
}

func (h *Handler) GetFiles(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	results, err := h.usecase.GetFiles(r.Context(), query)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)

	URLs := []string{}
	for _, file := range results {
		URLs = append(URLs, file.URL)
	}

	response := struct {
		Body []string `json:"body"`
	}{
		Body: URLs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
