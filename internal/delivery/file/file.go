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
	err := r.ParseMultipartForm(30 << 20) // Максимальный размер файла 30 MB
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

	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	err = h.usecase.Upload(r.Context(), file.File{
		File:        f,
		Filename:    handler.Filename,
		Size:        handler.Size,
		ContentType: handler.Header["Content-Type"][0],
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully: %s", handler.Filename)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	fileType := r.URL.Query().Get("type")
	query := r.URL.Query().Get("query")

	fmt.Printf("Search Query: %s, File Type: %s\n", query, fileType)

	results, err := h.usecase.Search(r.Context(), file.SearchQuery{
		Query: query,
		Type:  fileType,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)

	filenames := []string{}
	for _, file := range results {
		filenames = append(filenames, file.Filename)
	}

	response := struct {
		Body []string `json:"body"`
	}{
		Body: filenames,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
