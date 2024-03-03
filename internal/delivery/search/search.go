package search

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Handler struct {
}

func (h *Handler) searchHandler(w http.ResponseWriter, r *http.Request) {
	fileType := r.URL.Query().Get("type")
	query := r.URL.Query().Get("query")

	fmt.Printf("Search Query: %s, File Type: %s\n", query, fileType)

	// Примерный ответ
	response := struct {
		Status int      `json:"status"`
		Body   []string `json:"body"`
	}{
		Status: 200,
		Body:   []string{"filename1", "filename2"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
